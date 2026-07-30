package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/internal/auth"
	"github.com/edsonmubezi/myapp/internal/platform/email"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var (
	resetTokenService *auth.ResetTokenService
	emailResolver     *email.Resolver
)

// SetResetTokenService sets the reset token service
func SetResetTokenService(service *auth.ResetTokenService) {
	resetTokenService = service
}

// SetEmailResolver sets the email resolver for org-specific SMTP with .env fallback.
func SetEmailResolver(resolver *email.Resolver) {
	emailResolver = resolver
}

// ForgotPasswordRequest represents the request body for forgot password
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents the request body for reset password
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// VerifyResetTokenResponse represents the response for token verification
type VerifyResetTokenResponse struct {
	Valid     bool   `json:"valid"`
	ExpiresAt string `json:"expires_at,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ForgotPasswordHandler handles forgot password requests
// @Summary      Request Password Reset
// @Description  Sends a password reset email to the user. Returns the same success message regardless of whether the email exists (security best practice).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body ForgotPasswordRequest true "Email address" example({"email":"user@example.com"})
// @Success      200 {object} PasswordResetSuccessResponse "Reset email sent (or email not found)"
// @Failure      400 {object} ErrorResponse "Invalid email format"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/auth/forgot-password [post]
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
		return
	}

	// Validate email
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Valid email address is required",
		})
		return
	}

	// Get user by email
	user, err := userUseCase.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		// Security: Don't reveal if user exists or not
		log.Printf("User not found for email: %s, error: %v", req.Email, err)
		securejson.JSON(w, http.StatusOK, map[string]interface{}{
			"status":  http.StatusOK,
			"message": "If the email exists, a password reset link has been sent",
		})
		return
	}

	// Get IP address and user agent for audit
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Generate reset token
	token, err := resetTokenService.CreateResetToken(
		r.Context(),
		user.ID,
		ipAddress,
		userAgent,
		30*time.Minute, // Token expires in 30 minutes
	)
	if err != nil {
		log.Printf("Failed to create reset token: %v", err)
		securejson.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Failed to process password reset request",
		})
		return
	}

	// Build reset link
	baseURL := getBaseURL()
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)

	// Send email (resolve org-specific SMTP, fallback to .env)
	if emailResolver != nil {
		companyName := getEnv("COMPANY_NAME", "Microfinance")
		emailData := email.PasswordResetData{
			UserName:    user.FullName,
			ResetLink:   resetLink,
			ExpiresIn:   "30 minutes",
			CompanyName: companyName,
		}

		// Fetch org email branding so reset emails match org theme
		branding := email.DefaultBranding()
		if Organize != nil {
			if bs, bErr := Organize.GetEmailBranding(r.Context(), user.OrganizationID); bErr == nil && bs != nil {
				branding = email.EmailBranding{
					PrimaryColor:    bs.PrimaryColor,
					HeaderTextColor: bs.HeaderTextColor,
					AccentColor:     bs.AccentColor,
					FontFamily:      bs.FontFamily,
					FooterText:      bs.FooterText,
					SignOff:         bs.SignOff,
				}
			}
		}

		body, renderErr := email.RenderTemplate(email.TemplatePasswordReset, emailData, branding)
		if renderErr != nil {
			log.Printf("Failed to render password reset template: %v", renderErr)
		}

		emailSvc := emailResolver.ForOrg(r.Context(), user.OrganizationID)
		err = emailSvc.SendEmail(
			r.Context(),
			[]string{user.Email},
			"Password Reset Request",
			body,
			true,
		)
		if err != nil {
			log.Printf("Failed to send reset email: %v", err)
			securejson.JSON(w, http.StatusInternalServerError, map[string]interface{}{
				"status":  http.StatusInternalServerError,
				"message": "Failed to send reset email",
			})
			return
		}
	} else {
		// Email service not configured - log the reset link for development
		log.Printf("Password reset link for %s: %s", user.Email, resetLink)
	}

	// Return success response (same response whether user exists or not for security)
	securejson.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  http.StatusOK,
		"message": "If the email exists, a password reset link has been sent",
	})
}

// VerifyResetTokenHandler verifies if a reset token is valid
// @Summary      Verify Reset Token
// @Description  Validates a password reset token without consuming it. Use this to check token validity before showing password reset form.
// @Tags         Authentication
// @Produce      json
// @Param        token path string true "Reset token" example("base64-encoded-token-string")
// @Success      200 {object} TokenVerificationResponse "Token is valid"
// @Failure      400 {object} ErrorResponse "Invalid or expired token"
// @Router       /api/auth/verify-reset-token/{token} [get]
func VerifyResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	token := vars["token"]

	if token == "" {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Token is required",
		})
		return
	}

	// Validate the token
	userID, err := resetTokenService.ValidateResetToken(r.Context(), token)
	if err != nil {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
			"valid":   false,
		})
		return
	}

	// Get token info
	tokenInfo, err := resetTokenService.GetTokenInfo(r.Context(), token)
	if err != nil {
		log.Printf("Failed to get token info: %v", err)
	}

	response := VerifyResetTokenResponse{
		Valid:   true,
		Message: "Token is valid",
	}

	if tokenInfo != nil {
		response.ExpiresAt = tokenInfo.ExpiresAt.Format(time.RFC3339)
	}

	log.Printf("Token verified for user ID: %d", userID)

	securejson.JSON(w, http.StatusOK, response)
}

// ResetPasswordHandler handles password reset with token
// @Summary      Reset Password
// @Description  Resets user password using a valid reset token. Token can only be used once and expires after 30 minutes.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body ResetPasswordRequest true "Reset token and new password" example({"token":"base64-token","new_password":"NewSecurePass123!"})
// @Success      200 {object} MessageResponse "Password reset successful"
// @Failure      400 {object} ErrorResponse "Invalid token or weak password"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/auth/reset-password [post]
func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid request body",
		})
		return
	}

	// Validate required fields
	if req.Token == "" || req.NewPassword == "" {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Token and new password are required",
		})
		return
	}

	// Validate password strength
	if err := validatePasswordStrength(req.NewPassword); err != nil {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": err.Error(),
		})
		return
	}

	// Validate the token and get user ID
	userID, err := resetTokenService.ValidateResetToken(r.Context(), req.Token)
	if err != nil {
		securejson.JSON(w, http.StatusBadRequest, map[string]interface{}{
			"status":  http.StatusBadRequest,
			"message": "Invalid or expired token",
		})
		return
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		securejson.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Failed to process password reset",
		})
		return
	}

	// Update user password
	err = userUseCase.UpdatePassword(r.Context(), userID, string(hashedPassword))
	if err != nil {
		log.Printf("Failed to update password for user %d: %v", userID, err)
		securejson.JSON(w, http.StatusInternalServerError, map[string]interface{}{
			"status":  http.StatusInternalServerError,
			"message": "Failed to reset password",
		})
		return
	}

	// Mark token as used
	err = resetTokenService.MarkTokenAsUsed(r.Context(), req.Token)
	if err != nil {
		log.Printf("Failed to mark token as used: %v", err)
		// Don't fail the request, password has already been changed
	}

	// Get user details for confirmation email
	user, err := userUseCase.GetUserByIDNoOrg(r.Context(), userID)
	if err != nil {
		log.Printf("Failed to get user details: %v", err)
	}

	// Send confirmation email (resolve org-specific SMTP, fallback to .env)
	if emailResolver != nil && user != nil {
		companyName := getEnv("COMPANY_NAME", "Microfinance")
		emailData := email.PasswordResetConfirmationData{
			UserName:    user.FullName,
			CompanyName: companyName,
		}

		// Fetch org email branding
		branding := email.DefaultBranding()
		if Organize != nil {
			if bs, bErr := Organize.GetEmailBranding(context.Background(), user.OrganizationID); bErr == nil && bs != nil {
				branding = email.EmailBranding{
					PrimaryColor:    bs.PrimaryColor,
					HeaderTextColor: bs.HeaderTextColor,
					AccentColor:     bs.AccentColor,
					FontFamily:      bs.FontFamily,
					FooterText:      bs.FooterText,
					SignOff:         bs.SignOff,
				}
			}
		}

		body, renderErr := email.RenderTemplate(email.TemplatePasswordResetConfirmation, emailData, branding)
		if renderErr != nil {
			log.Printf("Failed to render password reset confirmation template: %v", renderErr)
		}

		emailSvc := emailResolver.ForOrg(context.Background(), user.OrganizationID)
		err = emailSvc.SendEmail(
			context.Background(),
			[]string{user.Email},
			"Password Changed Successfully",
			body,
			true,
		)
		if err != nil {
			log.Printf("Failed to send confirmation email: %v", err)
			// Don't fail the request
		}
	}

	log.Printf("Password reset successful for user ID: %d", userID)

	securejson.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  http.StatusOK,
		"message": "Password reset successful. You can now log in with your new password.",
	})
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if strings.Contains(ip, ":") {
		ip = strings.Split(ip, ":")[0]
	}

	return ip
}

func getBaseURL() string {
	// FRONTEND_URL is for user-facing links (password reset, email links, etc.)
	baseURL := os.Getenv("FRONTEND_URL")
	if baseURL == "" {
		// Fall back to first CORS origin (always points to frontend)
		if cors := os.Getenv("CORS_ALLOWED_ORIGIN"); cors != "" {
			baseURL = strings.Split(cors, ",")[0]
		}
	}
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}
	return strings.TrimSuffix(baseURL, "/")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
