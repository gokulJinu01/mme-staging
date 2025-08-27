//go:build integration

package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mme-tagging-service/internal/testutil"
)

func TestJWTMiddlewareIntegration(t *testing.T) {
	app := fiber.New()
	
	// Set test JWT secret
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	
	// Add JWT middleware
	app.Use(JWTMiddleware())
	
	// Test endpoint
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrgID(c)
		return c.JSON(fiber.Map{"orgId": orgID})
	})

	// Test with valid JWT
	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	assert.Equal(t, "test-org", result["orgId"])
}

func TestJWTMiddlewareMissingTokenIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test without Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddlewareInvalidTokenIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestJWTMiddlewareExpiredTokenIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test with expired token
	expiredToken, err := testutil.MakeExpiredToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAdminRoleIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	
	// Add JWT middleware first
	app.Use(JWTMiddleware())
	app.Use(RequireAdminRole())
	
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "admin"})
	})

	// Test with admin role
	adminToken, err := testutil.MakeToken("test-org", "test-project", []string{"admin"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestRequireAdminRoleNonAdminIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	
	app.Use(JWTMiddleware())
	app.Use(RequireAdminRole())
	
	app.Get("/admin", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "admin"})
	})

	// Test with non-admin role
	userToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestValidateTenancyIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	
	app.Use(JWTMiddleware())
	app.Use(ValidateTenancy())
	
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test with matching orgId
	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	requestBody := map[string]interface{}{
		"orgId":     "test-org",
		"projectId": "test-project",
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestValidateTenancyMismatchIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	
	app.Use(JWTMiddleware())
	app.Use(ValidateTenancy())
	
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	// Test with mismatched orgId
	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	requestBody := map[string]interface{}{
		"orgId":     "different-org",
		"projectId": "test-project",
	}
	
	bodyBytes, _ := json.Marshal(requestBody)
	
	req := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGetJWTClaimsIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		claims, err := GetJWTClaims(c)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{
			"orgId":     claims.OrgID,
			"projectId": claims.ProjectID,
			"roles":     claims.Roles,
		})
	})

	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user", "admin"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	assert.Equal(t, "test-org", result["orgId"])
	assert.Equal(t, "test-project", result["projectId"])
	
	roles, ok := result["roles"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, roles, "user")
	assert.Contains(t, roles, "admin")
}

func TestGetOrgIDIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		orgID := GetOrgID(c)
		return c.JSON(fiber.Map{"orgId": orgID})
	})

	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	assert.Equal(t, "test-org", result["orgId"])
}

func TestGetProjectIDIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		projectID := GetProjectID(c)
		return c.JSON(fiber.Map{"projectId": projectID})
	})

	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	assert.Equal(t, "test-project", result["projectId"])
}

func TestGetRolesIntegration(t *testing.T) {
	app := fiber.New()
	t.Setenv("JWT_SECRET", testutil.TestJWTSecret)
	app.Use(JWTMiddleware())
	
	app.Get("/test", func(c *fiber.Ctx) error {
		roles := GetRoles(c)
		return c.JSON(fiber.Map{"roles": roles})
	})

	validToken, err := testutil.MakeToken("test-org", "test-project", []string{"user", "admin"})
	require.NoError(t, err)
	
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)
	
	roles, ok := result["roles"].([]interface{})
	assert.True(t, ok)
	assert.Contains(t, roles, "user")
	assert.Contains(t, roles, "admin")
}
