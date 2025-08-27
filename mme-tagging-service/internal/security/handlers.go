package security

import (
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SecurityHandler handles security-related endpoints
type SecurityHandler struct {
	middleware *SecurityMiddleware
	metrics    *SecurityMetrics
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(middleware *SecurityMiddleware) *SecurityHandler {
	return &SecurityHandler{
		middleware: middleware,
		metrics:    NewSecurityMetrics(),
	}
}

// GetSecurityHealth returns security health status
func (sh *SecurityHandler) GetSecurityHealth(c *fiber.Ctx) error {
	health := fiber.Map{
		"status":           "UP",
		"service":          "mme-tagging-security",
		"rate_limiter":     "UP",
		"threat_detector":  "UP",
		"audit_logger":     "UP",
		"security_monitor": "UP",
		"timestamp":        time.Now(),
	}

	return c.JSON(health)
}

// GetSecurityMetrics returns security metrics (admin only)
func (sh *SecurityHandler) GetSecurityMetrics(c *fiber.Ctx) error {
	// Check if user has admin role
	userRoles := c.Get("X-User-Roles")
	if !hasRequiredRole(userRoles, "ADMIN") {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Access denied - admin role required",
			"code":  "INSUFFICIENT_PRIVILEGES",
		})
	}

	metrics := sh.metrics.GetMetrics()
	
	// Log metrics access
	userID := c.Get("X-User-ID")
	sh.middleware.auditor.LogSecurityEvent("SECURITY_METRICS_ACCESS", userID, c.IP(),
		"Security metrics accessed by admin user")

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   metrics,
	})
}

// SecurityMetrics tracks security-related metrics
type SecurityMetrics struct {
	RateLimitViolations  int64                  `json:"rate_limit_violations"`
	ThreatDetections     int64                  `json:"threat_detections"`
	UnauthorizedAccess   int64                  `json:"unauthorized_access"`
	ThreatsByType        map[string]int64       `json:"threats_by_type"`
	ViolationsByEndpoint map[string]int64       `json:"violations_by_endpoint"`
	LastUpdated          time.Time              `json:"last_updated"`
	mutex                sync.RWMutex
}

// NewSecurityMetrics creates a new security metrics instance
func NewSecurityMetrics() *SecurityMetrics {
	return &SecurityMetrics{
		ThreatsByType:        make(map[string]int64),
		ViolationsByEndpoint: make(map[string]int64),
		LastUpdated:          time.Now(),
	}
}

// RecordRateLimitViolation records a rate limit violation
func (sm *SecurityMetrics) RecordRateLimitViolation(endpoint string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.RateLimitViolations++
	sm.ViolationsByEndpoint[endpoint]++
	sm.LastUpdated = time.Now()
}

// RecordThreatDetection records a threat detection
func (sm *SecurityMetrics) RecordThreatDetection(threatType string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.ThreatDetections++
	sm.ThreatsByType[threatType]++
	sm.LastUpdated = time.Now()
}

// RecordUnauthorizedAccess records an unauthorized access attempt
func (sm *SecurityMetrics) RecordUnauthorizedAccess() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	sm.UnauthorizedAccess++
	sm.LastUpdated = time.Now()
}

// GetMetrics returns current security metrics
func (sm *SecurityMetrics) GetMetrics() SecurityMetrics {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	
	// Create a copy to avoid race conditions
	threatsByType := make(map[string]int64)
	for k, v := range sm.ThreatsByType {
		threatsByType[k] = v
	}
	
	violationsByEndpoint := make(map[string]int64)
	for k, v := range sm.ViolationsByEndpoint {
		violationsByEndpoint[k] = v
	}
	
	return SecurityMetrics{
		RateLimitViolations:  sm.RateLimitViolations,
		ThreatDetections:     sm.ThreatDetections,
		UnauthorizedAccess:   sm.UnauthorizedAccess,
		ThreatsByType:        threatsByType,
		ViolationsByEndpoint: violationsByEndpoint,
		LastUpdated:          sm.LastUpdated,
	}
}