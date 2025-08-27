package security

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

// SecurityConfig holds security configuration
type SecurityConfig struct {
	RateLimitEnabled    bool
	ThreatDetection     bool
	AuditLogging        bool
	SecurityHeaders     bool
	InputValidation     bool
	MaxRequestSize      int64
	RateLimitPerMinute  int
}

// DefaultSecurityConfig returns default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		RateLimitEnabled:    true,
		ThreatDetection:     true,
		AuditLogging:        true,
		SecurityHeaders:     true,
		InputValidation:     true,
		MaxRequestSize:      10 * 1024 * 1024, // 10MB
		RateLimitPerMinute:  100,
	}
}

// SecurityMiddleware provides comprehensive security for MME Tagging Service
type SecurityMiddleware struct {
	config      SecurityConfig
	rateLimiter *RateLimiter
	auditor     *SecurityAuditor
	detector    *ThreatDetector
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		config:      config,
		rateLimiter: NewRateLimiter(config.RateLimitPerMinute),
		auditor:     NewSecurityAuditor(),
		detector:    NewThreatDetector(),
	}
}

// SecurityHeaders middleware adds security headers
func (sm *SecurityMiddleware) SecurityHeaders() fiber.Handler {
	if !sm.config.SecurityHeaders {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		// Security headers
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Content-Security-Policy", "default-src 'self'")
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		return c.Next()
	}
}

// RateLimit middleware implements rate limiting
func (sm *SecurityMiddleware) RateLimit() fiber.Handler {
	if !sm.config.RateLimitEnabled {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		clientIP := c.IP()
		
		if !sm.rateLimiter.IsAllowed(userID, clientIP) {
			if sm.config.AuditLogging {
				sm.auditor.LogSecurityEvent("RATE_LIMIT_VIOLATION", userID, clientIP, 
					fmt.Sprintf("Rate limit exceeded for endpoint %s", c.Path()))
			}

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"code":  "RATE_LIMIT_EXCEEDED",
				"retry_after": 60,
			})
		}

		return c.Next()
	}
}

// ThreatDetection middleware detects malicious requests
func (sm *SecurityMiddleware) ThreatDetection() fiber.Handler {
	if !sm.config.ThreatDetection {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		clientIP := c.IP()
		
		// Check for threats
		if threat := sm.detector.AnalyzeRequest(c); threat != nil {
			if sm.config.AuditLogging {
				sm.auditor.LogSecurityEvent("THREAT_DETECTED", userID, clientIP, 
					fmt.Sprintf("Threat: %s - %s", threat.Type, threat.Description))
			}

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Request blocked by security policy",
				"code":  "SECURITY_VIOLATION",
			})
		}

		return c.Next()
	}
}

// InputValidation middleware validates request input
func (sm *SecurityMiddleware) InputValidation() fiber.Handler {
	if !sm.config.InputValidation {
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	return func(c *fiber.Ctx) error {
		userID := c.Get("X-User-ID")
		clientIP := c.IP()

		// Check request size
		if c.Context().Request.Header.ContentLength() > int(sm.config.MaxRequestSize) {
			if sm.config.AuditLogging {
				sm.auditor.LogSecurityEvent("OVERSIZED_REQUEST", userID, clientIP, 
					fmt.Sprintf("Request size: %d bytes", c.Context().Request.Header.ContentLength()))
			}

			return c.Status(fiber.StatusRequestEntityTooLarge).JSON(fiber.Map{
				"error": "Request too large",
				"code":  "REQUEST_TOO_LARGE",
			})
		}

		// Check for path traversal
		path := c.Path()
		if strings.Contains(path, "..") || strings.Contains(path, "~") {
			if sm.config.AuditLogging {
				sm.auditor.LogSecurityEvent("PATH_TRAVERSAL_ATTEMPT", userID, clientIP, 
					fmt.Sprintf("Suspicious path: %s", path))
			}

			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid request path",
				"code":  "INVALID_PATH",
			})
		}

		return c.Next()
	}
}

// AuthenticationMiddleware validates user authentication
func (sm *SecurityMiddleware) AuthenticationMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// JWT authentication is handled by Traefik gateway
		// Extract user information from gateway headers
		userID := c.Get("X-User-ID")
		userRoles := c.Get("X-User-Roles")

		// Set user context for downstream handlers
		if userID != "" {
			c.Locals("userID", userID)
		}
		if userRoles != "" {
			c.Locals("userRoles", userRoles)
		}

		return c.Next()
	}
}

// RoleBasedAccessControl creates role-based access control middleware
func (sm *SecurityMiddleware) RoleBasedAccessControl(requiredRole string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRoles := c.Get("X-User-Roles")
		userID := c.Get("X-User-ID")
		
		if !hasRequiredRole(userRoles, requiredRole) {
			if sm.config.AuditLogging {
				sm.auditor.LogSecurityEvent("UNAUTHORIZED_ACCESS", userID, c.IP(), 
					fmt.Sprintf("Access denied to %s - required role: %s", c.Path(), requiredRole))
			}

			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Access denied - insufficient privileges",
				"code":  "INSUFFICIENT_PRIVILEGES",
				"required_role": requiredRole,
			})
		}

		return c.Next()
	}
}

// hasRequiredRole checks if user has the required role
func hasRequiredRole(userRoles, requiredRole string) bool {
	if userRoles == "" {
		return false
	}

	// Check for admin role (has all permissions)
	if strings.Contains(userRoles, "ADMIN") {
		return true
	}

	// Check for specific role
	return strings.Contains(userRoles, requiredRole)
}

// RateLimiter implements in-memory rate limiting
type RateLimiter struct {
	requests map[string]*RateLimitEntry
	limit    int
	mutex    sync.RWMutex
}

type RateLimitEntry struct {
	Count     int
	ResetTime time.Time
}

func NewRateLimiter(limit int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*RateLimitEntry),
		limit:    limit,
	}
	
	// Start cleanup goroutine
	go rl.cleanup()
	
	return rl
}

func (rl *RateLimiter) IsAllowed(userID, clientIP string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	key := fmt.Sprintf("%s:%s", userID, clientIP)
	now := time.Now()

	entry, exists := rl.requests[key]
	if !exists {
		rl.requests[key] = &RateLimitEntry{
			Count:     1,
			ResetTime: now.Add(time.Minute),
		}
		return true
	}

	// Reset if time window expired
	if now.After(entry.ResetTime) {
		entry.Count = 1
		entry.ResetTime = now.Add(time.Minute)
		return true
	}

	if entry.Count >= rl.limit {
		return false
	}

	entry.Count++
	return true
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			if now.After(entry.ResetTime.Add(5 * time.Minute)) {
				delete(rl.requests, key)
			}
		}
		rl.mutex.Unlock()
	}
}

// SecurityAuditor handles security event logging
type SecurityAuditor struct{}

func NewSecurityAuditor() *SecurityAuditor {
	return &SecurityAuditor{}
}

func (sa *SecurityAuditor) LogSecurityEvent(eventType, userID, clientIP, description string) {
	log.Printf("[SECURITY EVENT] Type: %s, User: %s, IP: %s, Description: %s, Service: mme-tagging-service, Timestamp: %s",
		eventType, userID, clientIP, description, time.Now().Format(time.RFC3339))
}

// ThreatDetector analyzes requests for security threats
type ThreatDetector struct{}

type Threat struct {
	Type        string
	Severity    string
	Description string
}

func NewThreatDetector() *ThreatDetector {
	return &ThreatDetector{}
}

func (td *ThreatDetector) AnalyzeRequest(c *fiber.Ctx) *Threat {
	// Check query parameters for SQL injection
	queryString := string(c.Context().QueryArgs().QueryString())
	if td.containsSQLInjection(queryString) {
		return &Threat{
			Type:        "SQL_INJECTION",
			Severity:    "HIGH",
			Description: "SQL injection attempt detected in query parameters",
		}
	}

	// Check for XSS attempts
	if td.containsXSS(queryString) {
		return &Threat{
			Type:        "XSS_ATTEMPT",
			Severity:    "MEDIUM",
			Description: "Cross-site scripting attempt detected",
		}
	}

	// Check user agent for suspicious bots
	userAgent := c.Get("User-Agent")
	if td.isSuspiciousUserAgent(userAgent) {
		return &Threat{
			Type:        "SUSPICIOUS_BOT",
			Severity:    "LOW",
			Description: "Suspicious bot or scanner detected",
		}
	}

	return nil
}

func (td *ThreatDetector) containsSQLInjection(input string) bool {
	sqlKeywords := []string{
		"union", "select", "insert", "update", "delete", "drop", "create", "alter",
		"exec", "execute", "sp_", "xp_", "script", "javascript", "vbscript",
	}

	inputLower := strings.ToLower(input)
	for _, keyword := range sqlKeywords {
		if strings.Contains(inputLower, keyword) {
			return true
		}
	}
	return false
}

func (td *ThreatDetector) containsXSS(input string) bool {
	xssPatterns := []string{
		"<script", "</script>", "javascript:", "onload=", "onerror=", "onclick=",
		"alert(", "document.cookie", "window.location", "eval(",
	}

	inputLower := strings.ToLower(input)
	for _, pattern := range xssPatterns {
		if strings.Contains(inputLower, pattern) {
			return true
		}
	}
	return false
}

func (td *ThreatDetector) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousAgents := []string{
		"nmap", "sqlmap", "nikto", "burp", "metasploit", "scanner", "crawler",
		"bot", "spider", "scraper", "harvest", "extract",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, agent := range suspiciousAgents {
		if strings.Contains(userAgentLower, agent) {
			return true
		}
	}
	return false
}