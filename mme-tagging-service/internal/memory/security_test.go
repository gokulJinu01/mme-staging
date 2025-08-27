package memory

import (
	"strings"
	"testing"
	"time"
)

// TestSecurityEdgeCases tests potential security vulnerabilities and edge cases
func TestSecurityEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		prompt      string
		shouldPanic bool
		expectTags  bool
	}{
		{
			name:       "MongoDB injection attempt",
			prompt:     `{"$ne": null} OR 1==1 OR {"$where": "function() { return true; }"}`,
			expectTags: true, // Should be treated as normal text
		},
		{
			name:       "JavaScript injection",
			prompt:     `<script>alert('xss')</script> AND function() { return db.users.find(); }`,
			expectTags: true,
		},
		{
			name:       "SQL injection patterns",
			prompt:     `'; DROP TABLE users; -- OR 1=1 UNION SELECT * FROM passwords`,
			expectTags: true,
		},
		{
			name:       "Unicode normalization attack",
			prompt:     "＜script＞alert('unicode')＜/script＞",
			expectTags: true,
		},
		{
			name:       "Null bytes and control chars",
			prompt:     "test\x00\x01\x02\x03\x04\x05",
			expectTags: true,
		},
		{
			name:       "Very long single word",
			prompt:     strings.Repeat("a", 10000),
			expectTags: true,
		},
		{
			name:       "Emoji and special unicode",
			prompt:     "🧠 AI model deployment 🚀 with ñoño characters",
			expectTags: true,
		},
		{
			name:       "Mixed HTML and code",
			prompt:     `<div>Deploy model</div> with function() { return "test"; } and #include <stdio.h>`,
			expectTags: true,
		},
		{
			name:       "Regex DoS patterns",
			prompt:     strings.Repeat("a", 1000) + strings.Repeat("b", 1000),
			expectTags: true,
		},
		{
			name:       "BSON injection patterns",
			prompt:     `$where: function() { return true; } $ne: null $gt: ""`,
			expectTags: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldPanic {
					t.Errorf("CleanPromptToTags() panicked unexpectedly: %v", r)
				}
			}()

			tags := CleanPromptToTags(tt.prompt)

			if tt.expectTags && len(tags) == 0 {
				t.Logf("Note: %s produced no tags (might be expected)", tt.name)
			}

			// Ensure no dangerous patterns leak through
			for _, tag := range tags {
				if strings.Contains(tag, "$") {
					t.Errorf("Dangerous MongoDB operator found in tag: %s", tag)
				}
				if strings.Contains(tag, "<") || strings.Contains(tag, ">") {
					t.Errorf("HTML/XML characters found in tag: %s", tag)
				}
				if strings.Contains(tag, "script") {
					t.Errorf("Script pattern found in tag: %s", tag)
				}
				if len(tag) == 0 {
					t.Errorf("Empty tag produced")
				}
				if len(tag) > 100 {
					t.Errorf("Suspiciously long tag: %d characters", len(tag))
				}
			}
		})
	}
}

// TestMultiLanguagePrompts tests non-English content
func TestMultiLanguagePrompts(t *testing.T) {
	tests := []struct {
		name    string
		prompt  string
		minTags int
	}{
		{
			name:    "Spanish text",
			prompt:  "Configurar la base de datos con autenticación y conexión segura",
			minTags: 3,
		},
		{
			name:    "French text",
			prompt:  "Déployer le nouveau service d'authentification avec MongoDB",
			minTags: 3,
		},
		{
			name:    "German text",
			prompt:  "Neue Authentifizierungsdienst mit MongoDB-Integration konfigurieren",
			minTags: 2,
		},
		{
			name:    "Mixed languages",
			prompt:  "Deploy el nuevo authentication service avec MongoDB integración",
			minTags: 4,
		},
		{
			name:    "Chinese characters",
			prompt:  "部署新的身份验证服务 deploy authentication service",
			minTags: 2, // Should at least get English words
		},
		{
			name:    "Arabic text",
			prompt:  "نشر خدمة المصادقة الجديدة deploy authentication service",
			minTags: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tags := CleanPromptToTags(tt.prompt)

			if len(tags) < tt.minTags {
				t.Logf("Warning: %s produced only %d tags, expected at least %d",
					tt.name, len(tags), tt.minTags)
			}

			// Check for reasonable tag lengths
			for _, tag := range tags {
				if len(tag) < 2 {
					t.Errorf("Tag too short: '%s'", tag)
				}
			}
		})
	}
}

// TestLargePromptHandling tests performance with large inputs
func TestLargePromptHandling(t *testing.T) {
	tests := []struct {
		name       string
		promptSize int
		maxTime    int // milliseconds
	}{
		{
			name:       "1KB prompt",
			promptSize: 1024,
			maxTime:    10,
		},
		{
			name:       "10KB prompt",
			promptSize: 10240,
			maxTime:    50,
		},
		{
			name:       "100KB prompt",
			promptSize: 102400,
			maxTime:    500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create large prompt with realistic content
			baseText := "Deploy the new authentication service with MongoDB integration and Redis caching. "
			prompt := ""
			for len(prompt) < tt.promptSize {
				prompt += baseText
			}
			prompt = prompt[:tt.promptSize] // Truncate to exact size

			// Measure processing time
			start := time.Now()
			tags := CleanPromptToTags(prompt)
			elapsed := time.Since(start)
			elapsedMs := elapsed.Milliseconds()

			if elapsedMs > int64(tt.maxTime) {
				t.Errorf("Processing took %dms, expected max %dms", elapsedMs, tt.maxTime)
			}

			if len(tags) == 0 {
				t.Error("Large prompt produced no tags")
			}

			t.Logf("Processed %d chars in %dms, produced %d tags",
				tt.promptSize, elapsedMs, len(tags))
		})
	}
}

// TestTagNormalization tests various tag normalization edge cases
func TestTagNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "Case variations",
			input:    []string{"API", "api", "Api", "aPI"},
			expected: []string{"api"},
		},
		{
			name:     "Punctuation variations",
			input:    []string{"user-data", "user_data", "user.data", "userdata"},
			expected: []string{"userdata"},
		},
		{
			name:     "Special characters",
			input:    []string{"model@v1", "model#v1", "model$v1"},
			expected: []string{"modelv1"},
		},
		{
			name:     "Whitespace variations",
			input:    []string{" mongodb ", "\tmongodb\n", "  mongodb  "},
			expected: []string{"mongodb"},
		},
		{
			name:     "Empty and short inputs",
			input:    []string{"", "a", "ab", "abc", "abcd"},
			expected: []string{"abc", "abcd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanTagsForQuery(tt.input)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d tags, got %d", len(tt.expected), len(result))
			}

			// Check each expected tag is present
			for _, expectedTag := range tt.expected {
				found := false
				for _, resultTag := range result {
					if resultTag == expectedTag {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected tag '%s' not found in result: %v", expectedTag, result)
				}
			}
		})
	}
}

// BenchmarkLargePrompt benchmarks performance with realistic large prompts
func BenchmarkLargePrompt(b *testing.B) {
	// Create a 5KB realistic prompt
	prompt := `
	Deploy the new authentication service with MongoDB integration and Redis caching.
	The service should handle user registration, login, password reset, and session management.
	We need to ensure proper error handling, logging, and monitoring capabilities.
	The authentication flow should include JWT token generation and validation.
	Database connections must be pooled and properly configured for high availability.
	Redis will be used for session storage and caching frequently accessed user data.
	The service should support rate limiting to prevent brute force attacks.
	All endpoints must be properly documented with OpenAPI specifications.
	Security headers should be implemented including CORS, CSRF protection.
	The deployment should use Docker containers with health checks and proper resource limits.
	Logging should be structured and sent to centralized monitoring systems.
	Performance metrics should be exposed via Prometheus endpoints.
	The service needs to integrate with the existing API gateway and load balancer.
	Configuration should be externalized using environment variables and config files.
	Database migrations should be handled automatically during deployments.
	` + strings.Repeat("Additional context data for the authentication service deployment. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanPromptToTags(prompt)
	}
}
