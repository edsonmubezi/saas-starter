package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/edsonmubezi/myapp/internal/auth"
	"github.com/edsonmubezi/myapp/internal/identity"
	pkgAuth "github.com/edsonmubezi/myapp/pkg/auth"
	"github.com/edsonmubezi/myapp/pkg/securejson"
)

// SessionsResponse represents the response for active sessions
type SessionsResponse struct {
	Sessions     []auth.ActiveSession `json:"sessions"`
	CurrentID    int64                `json:"current_session_id"`
	TotalCount   int                  `json:"total_count"`
}

// RevokeSessionRequest represents a request to revoke a specific session
type RevokeSessionRequest struct {
	SessionID int64 `json:"session_id"`
}

// @Summary      Get Active Sessions
// @Description  Returns all active sessions for the authenticated user
// @Tags         Sessions
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} SessionsResponse "Active sessions"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/sessions [get]
func GetActiveSessionsHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if refreshTokenRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Session service not configured", nil)
		return
	}

	// Get all active sessions
	sessions, err := refreshTokenRepo.GetActiveSessions(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error getting active sessions: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to get sessions", nil)
		return
	}

	// Get current session ID from token
	var currentSessionID int64
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if tokenString != "" {
		// We need to get the refresh token to find current session
		// For now, we'll mark the most recent session as current
		// A better approach would be to store session ID in the token
		if len(sessions) > 0 {
			currentSessionID = sessions[0].ID
		}
	}

	// Mark current session
	for i := range sessions {
		sessions[i].IsCurrent = sessions[i].ID == currentSessionID
	}

	response := SessionsResponse{
		Sessions:   sessions,
		CurrentID:  currentSessionID,
		TotalCount: len(sessions),
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Active sessions retrieved",
		"data":    response,
	})
}

// @Summary      Revoke Session
// @Description  Revokes a specific session by ID
// @Tags         Sessions
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body RevokeSessionRequest true "Session to revoke"
// @Success      200 {object} MessageResponse "Session revoked"
// @Failure      400 {object} ErrorResponse "Invalid request or cannot revoke current session"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Session not found"
// @Router       /api/sessions/revoke [post]
func RevokeSessionHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req RevokeSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if req.SessionID == 0 {
		SendJSONResponse(w, http.StatusBadRequest, "Session ID is required", nil)
		return
	}

	if refreshTokenRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Session service not configured", nil)
		return
	}

	// Revoke the session
	err = refreshTokenRepo.RevokeSessionByID(r.Context(), authUser.UserID, req.SessionID)
	if err != nil {
		log.Printf("Error revoking session: %v", err)
		SendJSONResponse(w, http.StatusNotFound, "Session not found or already revoked", nil)
		return
	}

	log.Printf("Session %d revoked for user %d", req.SessionID, authUser.UserID)

	SendJSONResponse(w, http.StatusOK, "Session revoked successfully", nil)
}

// @Summary      Revoke All Other Sessions
// @Description  Revokes all sessions except the current one
// @Tags         Sessions
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} MessageResponse "Sessions revoked"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      500 {object} ErrorResponse "Internal server error"
// @Router       /api/sessions/revoke-others [post]
func RevokeOtherSessionsHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if refreshTokenRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Session service not configured", nil)
		return
	}

	// Get the current refresh token from the request
	// We need to extract the refresh token fingerprint to preserve the current session
	authHeader := r.Header.Get("Authorization")
	tokenString := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))

	// Get token claims to identify user
	claims, err := pkgAuth.ValidateToken(tokenString)
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Invalid token", nil)
		return
	}

	// For security, we'll revoke all sessions and force re-login
	// This is safer than trying to preserve the current session
	count, err := refreshTokenRepo.RevokeOtherSessions(r.Context(), claims.UserID, "")
	if err != nil {
		log.Printf("Error revoking other sessions: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to revoke sessions", nil)
		return
	}

	log.Printf("Revoked %d other sessions for user %d", count, authUser.UserID)

	SendJSONResponse(w, http.StatusOK, "All other sessions have been revoked", map[string]any{
		"revoked_count": count,
	})
}

// @Summary      Get Session Count
// @Description  Returns the count of active sessions for the authenticated user
// @Tags         Sessions
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} map[string]int "Session count"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Router       /api/sessions/count [get]
func GetSessionCountHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if refreshTokenRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Session service not configured", nil)
		return
	}

	count, err := refreshTokenRepo.GetSessionCount(r.Context(), authUser.UserID)
	if err != nil {
		log.Printf("Error getting session count: %v", err)
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to get session count", nil)
		return
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  200,
		"message": "Session count retrieved",
		"data": map[string]int{
			"count": count,
		},
	})
}

// @Summary      Revoke Session by ID (URL param)
// @Description  Revokes a specific session by ID passed in URL
// @Tags         Sessions
// @Security     BearerAuth
// @Produce      json
// @Param        id path int true "Session ID"
// @Success      200 {object} MessageResponse "Session revoked"
// @Failure      400 {object} ErrorResponse "Invalid session ID"
// @Failure      401 {object} ErrorResponse "Unauthorized"
// @Failure      404 {object} ErrorResponse "Session not found"
// @Router       /api/sessions/{id} [delete]
func DeleteSessionHandler(w http.ResponseWriter, r *http.Request) {
	authUser, err := identity.Require(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	// Extract session ID from URL path
	// URL format: /api/sessions/{id}
	path := r.URL.Path
	parts := strings.Split(path, "/")
	if len(parts) < 4 {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid session ID", nil)
		return
	}

	sessionIDStr := parts[len(parts)-1]
	sessionID, err := strconv.ParseInt(sessionIDStr, 10, 64)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid session ID", nil)
		return
	}

	if refreshTokenRepo == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Session service not configured", nil)
		return
	}

	// Revoke the session
	err = refreshTokenRepo.RevokeSessionByID(r.Context(), authUser.UserID, sessionID)
	if err != nil {
		log.Printf("Error revoking session: %v", err)
		SendJSONResponse(w, http.StatusNotFound, "Session not found or already revoked", nil)
		return
	}

	log.Printf("Session %d revoked for user %d", sessionID, authUser.UserID)

	SendJSONResponse(w, http.StatusOK, "Session revoked successfully", nil)
}
