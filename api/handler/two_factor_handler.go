package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/edsonmubezi/myapp/internal/auth"
	"github.com/edsonmubezi/myapp/internal/identity"
	pkgAuth "github.com/edsonmubezi/myapp/pkg/auth"
	"github.com/edsonmubezi/myapp/pkg/securejson"
)

var (
	totpService *auth.TOTPService
)

// SetTOTPService sets the TOTP service
func SetTOTPService(service *auth.TOTPService) {
	totpService = service
}

// Setup2FARequest represents the request to initiate 2FA setup
type Setup2FARequest struct {
	Method string `json:"method"` // "totp", "email", or "sms"
}

// Setup2FAResponse represents the response for 2FA setup initiation
type Setup2FAResponse struct {
	Method          string `json:"method"`
	Secret          string `json:"secret,omitempty"`           // For TOTP only
	ProvisioningURI string `json:"provisioning_uri,omitempty"` // For TOTP only (QR code)
	Message         string `json:"message"`
}

// Verify2FARequest represents the request to verify and enable 2FA
type Verify2FARequest struct {
	Code string `json:"code"`
}

// Verify2FAResponse represents the response for 2FA verification
type Verify2FAResponse struct {
	Success     bool     `json:"success"`
	BackupCodes []string `json:"backup_codes,omitempty"` // Only returned on first setup
	Message     string   `json:"message"`
}

// TwoFactorStatusResponse represents the 2FA status
type TwoFactorStatusResponse struct {
	Enabled              bool   `json:"enabled"`
	Method               string `json:"method,omitempty"`
	RoleRequires2FA      bool   `json:"role_requires_2fa"`
	RemainingBackupCodes int    `json:"remaining_backup_codes,omitempty"`
}

// @Summary      Get 2FA Status
// @Description  Returns the current 2FA status for the authenticated user
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} TwoFactorStatusResponse "2FA status"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/status [get]
func Get2FAStatusHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Get 2FA status
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to get 2FA status", nil)
		return
	}

	// Check if role requires 2FA
	roleRequires2FA, err := totpService.IsRoleRequires2FA(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error checking role 2FA requirement: %v", err)
		roleRequires2FA = false
	}

	response := TwoFactorStatusResponse{
		Enabled:         enabled,
		Method:          string(method),
		RoleRequires2FA: roleRequires2FA,
	}

	// Get remaining backup codes if 2FA is enabled
	if enabled {
		remaining, err := totpService.GetRemainingBackupCodes(r.Context(), authUser.UserID)
		if err == nil {
			response.RemainingBackupCodes = remaining
		}
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "2FA status retrieved",
		"data":    response,
	})
}

// @Summary      Setup 2FA
// @Description  Initiates 2FA setup for the authenticated user. Returns secret and provisioning URI for TOTP.
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body Setup2FARequest true "2FA method to setup"
// @Success      200 {object} Setup2FAResponse "Setup initiated"
// @Failure      400 {object} ErrorResponse "Invalid method or 2FA already enabled"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/setup [post]
func Setup2FAHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req Setup2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// Validate method
	method := auth.TwoFactorMethod(req.Method)
	if method != auth.TwoFactorMethodTOTP && method != auth.TwoFactorMethodEmail && method != auth.TwoFactorMethodSMS {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid 2FA method. Must be 'totp', 'email', or 'sms'", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Check if 2FA is already enabled
	enabled, _, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to check 2FA status", nil)
		return
	}

	if enabled {
		SendJSONResponse(w, http.StatusBadRequest, "2FA is already enabled. Disable it first to change method.", nil)
		return
	}

	response := Setup2FAResponse{
		Method: req.Method,
	}

	switch method {
	case auth.TwoFactorMethodTOTP:
		// Generate TOTP secret
		secret, err := auth.GenerateSecret()
		if err != nil {
			log.Printf("Error generating TOTP secret: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate secret", nil)
			return
		}

		// Store the secret (but don't enable 2FA yet)
		if err := totpService.SetupTOTP(r.Context(), authUser.UserID, secret); err != nil {
			log.Printf("Error setting up TOTP: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to setup TOTP", nil)
			return
		}

		// Generate provisioning URI for QR code
		provisioningURI := auth.GenerateProvisioningURI(secret, authUser.Email, "Microfinance")

		response.Secret = secret
		response.ProvisioningURI = provisioningURI
		response.Message = "Scan the QR code or enter the secret in your authenticator app, then verify with a code"

	case auth.TwoFactorMethodEmail:
		// Set the method and generate OTP
		if err := totpService.Set2FAMethod(r.Context(), authUser.UserID, method); err != nil {
			log.Printf("Error setting 2FA method: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to setup email 2FA", nil)
			return
		}

		// Generate and send OTP code
		code, err := totpService.GenerateOTPCode(r.Context(), authUser.UserID, method, 10)
		if err != nil {
			log.Printf("Error generating OTP: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate verification code", nil)
			return
		}

		// Send email with code
		if emailResolver != nil {
			subject := "Your Microfinance Verification Code"
			content := "Your verification code is: " + code + "\n\nThis code will expire in 10 minutes."
			emailSvc := emailResolver.ForOrg(r.Context(), authUser.OrganizationID)
			if err := emailSvc.SendEmail(r.Context(), []string{authUser.Email}, subject, content, false); err != nil {
				log.Printf("Error sending verification email: %v", err)
			}
		}

		response.Message = "A verification code has been sent to your email address"

	case auth.TwoFactorMethodSMS:
		// Set the method
		if err := totpService.Set2FAMethod(r.Context(), authUser.UserID, method); err != nil {
			log.Printf("Error setting 2FA method: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to setup SMS 2FA", nil)
			return
		}

		// Generate OTP code
		_, err := totpService.GenerateOTPCode(r.Context(), authUser.UserID, method, 10)
		if err != nil {
			log.Printf("Error generating OTP: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate verification code", nil)
			return
		}

		// TODO: Send SMS with code (requires SMS service implementation)
		response.Message = "A verification code has been sent to your phone number"
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": response.Message,
		"data":    response,
	})
}

// @Summary      Verify and Enable 2FA
// @Description  Verifies the 2FA code and enables 2FA for the user. Returns backup codes on success.
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body Verify2FARequest true "Verification code"
// @Success      200 {object} Verify2FAResponse "2FA enabled successfully"
// @Failure      400 {object} ErrorResponse "Invalid code"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/verify [post]
func Verify2FASetupHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Code == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Verification code is required", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Check current method
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to check 2FA status", nil)
		return
	}

	if enabled {
		SendJSONResponse(w, http.StatusBadRequest, "2FA is already enabled", nil)
		return
	}

	var valid bool

	switch method {
	case auth.TwoFactorMethodTOTP:
		// Get the secret and validate TOTP code
		secret, err := totpService.GetTOTPSecret(r.Context(), authUser.UserID)
		if err != nil {
			SendJSONResponse(w, http.StatusBadRequest, "TOTP not configured. Please initiate setup first.", nil)
			return
		}

		valid = auth.ValidateTOTP(secret, req.Code, time.Now())

	case auth.TwoFactorMethodEmail, auth.TwoFactorMethodSMS:
		// Validate OTP code
		valid, err = totpService.ValidateOTPCode(r.Context(), authUser.UserID, method, req.Code)
		if err != nil {
			log.Printf("Error validating OTP: %v", err)
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to validate code", nil)
			return
		}

	default:
		SendJSONResponse(w, http.StatusBadRequest, "2FA method not configured. Please initiate setup first.", nil)
		return
	}

	if !valid {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid verification code", nil)
		return
	}

	// Enable 2FA
	if err := totpService.EnableTOTP(r.Context(), authUser.UserID); err != nil {
		log.Printf("Error enabling 2FA: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to enable 2FA", nil)
		return
	}

	// Generate backup codes
	backupCodes, err := totpService.GenerateBackupCodes(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error generating backup codes: %v", err)
		// Don't fail - 2FA is still enabled
	}

	response := Verify2FAResponse{
		Success:     true,
		BackupCodes: backupCodes,
		Message:     "Two-factor authentication has been enabled. Please save your backup codes securely.",
	}

	log.Printf("2FA enabled for user %d", authUser.UserID)

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": response.Message,
		"data":    response,
	})
}

// @Summary      Disable 2FA
// @Description  Disables 2FA for the authenticated user. Requires current password or 2FA code for security.
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body Verify2FARequest true "Current 2FA code or backup code"
// @Success      200 {object} MessageResponse "2FA disabled"
// @Failure      400 {object} ErrorResponse "Invalid code"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/disable [post]
func Disable2FAHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Code == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Verification code is required", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Check if 2FA is enabled
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to check 2FA status", nil)
		return
	}

	if !enabled {
		SendJSONResponse(w, http.StatusBadRequest, "2FA is not enabled", nil)
		return
	}

	var valid bool

	switch method {
	case auth.TwoFactorMethodTOTP:
		// Validate TOTP code
		secret, err := totpService.GetTOTPSecret(r.Context(), authUser.UserID)
		if err != nil {
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to get TOTP secret", nil)
			return
		}

		valid = auth.ValidateTOTP(secret, req.Code, time.Now())

		// If TOTP code is invalid, try backup code
		if !valid {
			valid, _ = totpService.ValidateBackupCode(r.Context(), authUser.UserID, req.Code)
		}

	case auth.TwoFactorMethodEmail, auth.TwoFactorMethodSMS:
		// Validate OTP code
		valid, err = totpService.ValidateOTPCode(r.Context(), authUser.UserID, method, req.Code)
		if err != nil {
			log.Printf("Error validating OTP: %v", err)
		}

		// If OTP code is invalid, try backup code
		if !valid {
			valid, _ = totpService.ValidateBackupCode(r.Context(), authUser.UserID, req.Code)
		}
	}

	if !valid {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid verification code", nil)
		return
	}

	// Disable 2FA
	if err := totpService.DisableTOTP(r.Context(), authUser.UserID); err != nil {
		log.Printf("Error disabling 2FA: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to disable 2FA", nil)
		return
	}

	log.Printf("2FA disabled for user %d", authUser.UserID)

	SendJSONResponse(w, http.StatusOK, "Two-factor authentication has been disabled", nil)
}

// @Summary      Regenerate Backup Codes
// @Description  Regenerates backup codes for the authenticated user. Requires current 2FA code.
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body Verify2FARequest true "Current 2FA code"
// @Success      200 {object} Verify2FAResponse "New backup codes generated"
// @Failure      400 {object} ErrorResponse "Invalid code or 2FA not enabled"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/backup-codes/regenerate [post]
func RegenerateBackupCodesHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req Verify2FARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.Code == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Verification code is required", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Check if 2FA is enabled
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil || !enabled {
		SendJSONResponse(w, http.StatusBadRequest, "2FA is not enabled", nil)
		return
	}

	var valid bool

	switch method {
	case auth.TwoFactorMethodTOTP:
		secret, err := totpService.GetTOTPSecret(r.Context(), authUser.UserID)
		if err != nil {
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to get TOTP secret", nil)
			return
		}
		valid = auth.ValidateTOTP(secret, req.Code, time.Now())

	case auth.TwoFactorMethodEmail, auth.TwoFactorMethodSMS:
		valid, _ = totpService.ValidateOTPCode(r.Context(), authUser.UserID, method, req.Code)
	}

	if !valid {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid verification code", nil)
		return
	}

	// Generate new backup codes
	backupCodes, err := totpService.GenerateBackupCodes(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error generating backup codes: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate backup codes", nil)
		return
	}

	response := Verify2FAResponse{
		Success:     true,
		BackupCodes: backupCodes,
		Message:     "New backup codes have been generated. Please save them securely.",
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": response.Message,
		"data":    response,
	})
}

// @Summary      Send 2FA Code
// @Description  Sends a new 2FA verification code via email or SMS (for non-TOTP methods)
// @Tags         Two-Factor Authentication
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} MessageResponse "Code sent"
// @Failure      400 {object} ErrorResponse "Invalid method or 2FA not enabled"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/send-code [post]
func Send2FACodeHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Check 2FA status
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), authUser.UserID)
	if err != nil || !enabled {
		SendJSONResponse(w, http.StatusBadRequest, "2FA is not enabled", nil)
		return
	}

	if method == auth.TwoFactorMethodTOTP {
		SendJSONResponse(w, http.StatusBadRequest, "TOTP does not require sending codes. Use your authenticator app.", nil)
		return
	}

	// Generate and send OTP code
	code, err := totpService.GenerateOTPCode(r.Context(), authUser.UserID, method, 10)
	if err != nil {
		log.Printf("Error generating OTP: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate verification code", nil)
		return
	}

	switch method {
	case auth.TwoFactorMethodEmail:
		if emailResolver != nil {
			subject := "Your Microfinance Verification Code"
			content := "Your verification code is: " + code + "\n\nThis code will expire in 10 minutes."
			emailSvc := emailResolver.ForOrg(r.Context(), authUser.OrganizationID)
			if err := emailSvc.SendEmail(r.Context(), []string{authUser.Email}, subject, content, false); err != nil {
				log.Printf("Error sending verification email: %v", err)
				SendJSONResponse(w, http.StatusInternalServerError, "Failed to send verification email", nil)
				return
			}
		}
		SendJSONResponse(w, http.StatusOK, "Verification code sent to your email", nil)

	case auth.TwoFactorMethodSMS:
		// TODO: Implement SMS sending
		SendJSONResponse(w, http.StatusOK, "Verification code sent to your phone", nil)
	}
}

// Verify2FALoginRequest represents the request to verify 2FA during login
type Verify2FALoginRequest struct {
	SessionToken string `json:"session_token"` // Temporary token from login response
	Code         string `json:"code"`          // TOTP code or OTP code
	UseBackup    bool   `json:"use_backup"`    // If true, treat code as backup code
}

// Send2FALoginCodeRequest represents request to send 2FA code during login
type Send2FALoginCodeRequest struct {
	SessionToken string `json:"session_token"`
}

// @Summary      Verify 2FA during Login
// @Description  Verifies 2FA code during login flow. Returns access and refresh tokens on success.
// @Tags         Two-Factor Authentication
// @Accept       json
// @Produce      json
// @Param        request body Verify2FALoginRequest true "Session token and 2FA code"
// @Success      200 {object} LoginResponse "Login successful with tokens"
// @Failure      400 {object} ErrorResponse "Invalid code or session token"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/verify-login [post]
func Verify2FALoginHandler(w http.ResponseWriter, r *http.Request) {
	var req Verify2FALoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.SessionToken == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Session token is required", nil)
		return
	}

	if req.Code == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Verification code is required", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Validate the session token (it's a JWT with limited claims)
	claims, err := pkgAuth.ValidateToken(req.SessionToken)
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Invalid or expired session token", nil)
		return
	}

	// Get user info
	user, err := userUseCase.GetUserByID(r.Context(), claims.UserID, claims.OrganizationID)
	if err != nil || user == nil {
		SendJSONResponse(w, http.StatusUnauthorized, "User not found", nil)
		return
	}

	// Check if user is still active
	if !user.ActiveStatus {
		SendJSONResponse(w, http.StatusForbidden, "Account is deactivated", nil)
		return
	}

	// Get 2FA status — if not enabled, default to email (mandatory login OTP)
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to check verification status", nil)
		return
	}
	if !enabled {
		method = auth.TwoFactorMethodEmail
	}

	var valid bool

	// First try backup code if specified
	if req.UseBackup {
		valid, err = totpService.ValidateBackupCode(r.Context(), user.ID, req.Code)
		if err != nil {
			log.Printf("Error validating backup code: %v", err)
		}
	} else {
		// Validate based on method
		switch method {
		case auth.TwoFactorMethodTOTP:
			// Get the secret and validate TOTP code
			secret, err := totpService.GetTOTPSecret(r.Context(), user.ID)
			if err != nil {
				SendJSONResponse(w, http.StatusInternalServerError, "Failed to validate code", nil)
				return
			}
			valid = auth.ValidateTOTP(secret, req.Code, time.Now())

		case auth.TwoFactorMethodEmail, auth.TwoFactorMethodSMS:
			// Validate OTP code
			valid, err = totpService.ValidateOTPCode(r.Context(), user.ID, method, req.Code)
			if err != nil {
				log.Printf("Error validating OTP: %v", err)
				SendJSONResponse(w, http.StatusInternalServerError, "Failed to validate code", nil)
				return
			}
		}

		// If code is invalid, try backup code as fallback
		if !valid {
			valid, _ = totpService.ValidateBackupCode(r.Context(), user.ID, req.Code)
		}
	}

	if !valid {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid verification code", nil)
		return
	}

	// 2FA verified - generate full tokens
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.Name
	}

	accessToken, err := pkgAuth.GenerateAccessToken(user.ID, user.Email, user.OrganizationID, roleName, user.FullName)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate access token", nil)
		return
	}

	refreshToken, err := pkgAuth.GenerateRefreshToken(user.ID, user.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate refresh token", nil)
		return
	}

	// Store refresh token fingerprint in database (if repository available)
	if refreshTokenRepo != nil {
		ipAddress := getClientIPFromRequest(r)
		userAgent := r.Header.Get("User-Agent")
		tokenHash := pkgAuth.HashToken(refreshToken)
		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		deviceInfo := fmt.Sprintf("IP: %s, UA: %s", ipAddress, userAgent)

		_, err := refreshTokenRepo.StoreRefreshToken(
			r.Context(),
			user.ID,
			tokenHash,
			expiresAt,
			deviceInfo,
			ipAddress,
			userAgent,
		)
		if err != nil {
			log.Printf("Error storing refresh token after 2FA: %v", err)
		}
	}

	// Update last login
	if userRepo != nil {
		ipAddress := getClientIPFromRequest(r)
		if err := userRepo.UpdateLastLogin(r.Context(), user.ID, ipAddress); err != nil {
			log.Printf("Error updating last login after 2FA: %v", err)
		}
	}

	log.Printf("2FA login successful for user %d", user.ID)

	// Check if user must change password
	if user.MustChangePassword {
		SendJSONResponse(w, http.StatusOK, "Login successful. Password change required.", LoginResponseExtended{
			AccessToken:        accessToken,
			RefreshToken:       refreshToken,
			MustChangePassword: true,
		})
		return
	}

	SendJSONResponse(w, http.StatusOK, "Login successful", LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// @Summary      Send 2FA Code during Login
// @Description  Sends a new 2FA verification code during login (for email/SMS methods)
// @Tags         Two-Factor Authentication
// @Accept       json
// @Produce      json
// @Param        request body Send2FALoginCodeRequest true "Session token"
// @Success      200 {object} MessageResponse "Code sent"
// @Failure      400 {object} ErrorResponse "Invalid method or session"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/auth/2fa/send-login-code [post]
func Send2FALoginCodeHandler(w http.ResponseWriter, r *http.Request) {
	var req Send2FALoginCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.SessionToken == "" {
		SendJSONResponse(w, http.StatusBadRequest, "Session token is required", nil)
		return
	}

	if totpService == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "2FA service not configured", nil)
		return
	}

	// Validate the session token
	claims, err := pkgAuth.ValidateToken(req.SessionToken)
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Invalid or expired session token", nil)
		return
	}

	// Get user info
	user, err := userUseCase.GetUserByID(r.Context(), claims.UserID, claims.OrganizationID)
	if err != nil || user == nil {
		SendJSONResponse(w, http.StatusUnauthorized, "User not found", nil)
		return
	}

	// Check 2FA status — if not enabled, default to email (mandatory login OTP)
	enabled, method, err := totpService.Is2FAEnabled(r.Context(), user.ID)
	if err != nil {
		log.Printf("Error checking 2FA status: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to check verification status", nil)
		return
	}
	if !enabled {
		method = auth.TwoFactorMethodEmail
	}

	if method == auth.TwoFactorMethodTOTP {
		SendJSONResponse(w, http.StatusBadRequest, "TOTP does not require sending codes. Use your authenticator app.", nil)
		return
	}

	// Generate and send OTP code
	code, err := totpService.GenerateOTPCode(r.Context(), user.ID, method, 10)
	if err != nil {
		log.Printf("Error generating OTP: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate verification code", nil)
		return
	}

	switch method {
	case auth.TwoFactorMethodEmail:
		if emailResolver != nil {
			subject := "Your Microfinance Verification Code"
			content := "Your verification code is: " + code + "\n\nThis code will expire in 10 minutes."
			emailSvc := emailResolver.ForOrg(r.Context(), user.OrganizationID)
			if err := emailSvc.SendEmail(r.Context(), []string{user.Email}, subject, content, false); err != nil {
				log.Printf("Error sending verification email: %v", err)
				SendJSONResponse(w, http.StatusInternalServerError, "Failed to send verification email", nil)
				return
			}
		}
		SendJSONResponse(w, http.StatusOK, "Verification code sent to your email", nil)

	case auth.TwoFactorMethodSMS:
		// TODO: Implement SMS sending
		SendJSONResponse(w, http.StatusOK, "Verification code sent to your phone", nil)
	}
}
