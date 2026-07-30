package auth

import (
	"github.com/golang-jwt/jwt/v4"
)

type TokenClaims struct {
	UserID         int64  `json:"user_id"`
	OrganizationID int64  `json:"organization_id"`
	Email          string `json:"email"`
	FullName       string `json:"full_name,omitempty"`
	Role           string `json:"role,omitempty"`
	jwt.RegisteredClaims
}
