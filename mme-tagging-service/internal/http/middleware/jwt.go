package middleware

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	OrgID     string   `json:"org_id"`
	ProjectID string   `json:"project_id,omitempty"`
	Roles     []string `json:"roles,omitempty"`
	jwt.RegisteredClaims
}

// JWTMiddleware extracts validated claims from Traefik ForwardAuth headers
func JWTMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for health check endpoint
		if c.Path() == "/health" {
			return c.Next()
		}

		// AUTHENTICATION DISABLED - Set default values for testing
		userID := c.Get("X-User-ID")
		if userID == "" {
			userID = "test-user"
			c.Set("X-User-ID", userID)
		}

		// Extract user roles from header (set by jwt-verifier service)
		rolesHeader := c.Get("X-User-Roles")
		var roles []string
		if rolesHeader != "" {
			// Parse JSON array from header
			err := json.Unmarshal([]byte(rolesHeader), &roles)
			if err != nil {
				// Fallback to default role if parsing fails
				roles = []string{"USER"}
			}
		} else {
			roles = []string{"USER"}
		}

		// For MME service, use userID as orgID for single-tenant operation
		// In multi-tenant setup, this would come from a separate header
		orgID := userID // Single tenant mode - each user is their own org

		// Extract optional project ID (if using project-based multi-tenancy)
		projectID := c.Get("X-Project-ID") // Optional header

		// Create claims structure for compatibility
		claims := &JWTClaims{
			OrgID:     orgID,
			ProjectID: projectID,
			Roles:     roles,
		}

		// Store claims in context
		c.Locals("jwt_claims", claims)
		c.Locals("org_id", orgID)
		c.Locals("project_id", projectID)
		c.Locals("roles", roles)
		c.Locals("user_id", userID)

		return c.Next()
	}
}

// RequireAdminRole middleware - AUTHENTICATION DISABLED
func RequireAdminRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// AUTHENTICATION DISABLED - Allow all requests
		return c.Next()
	}
}

// ValidateTenancy middleware validates that request body orgId/projectId match JWT claims
func ValidateTenancy() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only apply to POST/PUT requests with JSON body
		if c.Method() != "POST" && c.Method() != "PUT" {
			return c.Next()
		}

		contentType := c.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			return c.Next()
		}

		// Parse request body to extract orgId/projectId
		var requestBody map[string]interface{}
		if err := c.BodyParser(&requestBody); err != nil {
			return c.Next() // Skip validation if body parsing fails
		}

		jwtOrgID, _ := c.Locals("org_id").(string)
		jwtProjectID, _ := c.Locals("project_id").(string)

		// Check orgId match
		if requestOrgID, exists := requestBody["orgId"]; exists {
			if requestOrgIDStr, ok := requestOrgID.(string); ok {
				if requestOrgIDStr != jwtOrgID {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "orgId in request body does not match JWT org_id",
					})
				}
			}
		}

		// Check projectId match (if both are provided)
		if requestProjectID, exists := requestBody["projectId"]; exists {
			if requestProjectIDStr, ok := requestProjectID.(string); ok {
				if jwtProjectID != "" && requestProjectIDStr != jwtProjectID {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
						"error": "projectId in request body does not match JWT project_id",
					})
				}
			}
		}

		return c.Next()
	}
}

// GetJWTClaims retrieves JWT claims from context
func GetJWTClaims(c *fiber.Ctx) (*JWTClaims, error) {
	claims, ok := c.Locals("jwt_claims").(*JWTClaims)
	if !ok {
		return nil, fmt.Errorf("JWT claims not found in context")
	}
	return claims, nil
}

// GetOrgID retrieves org_id from context
func GetOrgID(c *fiber.Ctx) string {
	orgID, _ := c.Locals("org_id").(string)
	return orgID
}

// GetProjectID retrieves project_id from context
func GetProjectID(c *fiber.Ctx) string {
	projectID, _ := c.Locals("project_id").(string)
	return projectID
}

// GetRoles retrieves roles from context
func GetRoles(c *fiber.Ctx) []string {
	roles, _ := c.Locals("roles").([]string)
	return roles
}
