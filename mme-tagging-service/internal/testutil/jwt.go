//go:build integration

package testutil

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// TestJWTSecret is a deterministic secret for testing
	TestJWTSecret = "mme-test-secret-key-for-jwt-validation"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	OrgID     string   `json:"org_id"`
	ProjectID string   `json:"project_id,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// MakeToken creates a JWT token for testing
func MakeToken(orgID, projectID string, roles []string) (string, error) {
	claims := JWTClaims{
		OrgID:     orgID,
		ProjectID: projectID,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// MakeTokenWithExpiry creates a JWT token with custom expiry
func MakeTokenWithExpiry(orgID, projectID string, roles []string, expiry time.Duration) (string, error) {
	claims := JWTClaims{
		OrgID:     orgID,
		ProjectID: projectID,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// MakeExpiredToken creates an expired JWT token for testing
func MakeExpiredToken(orgID, projectID string, roles []string) (string, error) {
	claims := JWTClaims{
		OrgID:     orgID,
		ProjectID: projectID,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(TestJWTSecret))
}

// MakeInvalidToken creates an invalid JWT token for testing
func MakeInvalidToken() string {
	return "invalid.jwt.token"
}

// ParseToken parses and validates a JWT token
func ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(TestJWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
