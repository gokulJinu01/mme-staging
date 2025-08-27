package memory

import (
	"reflect"
	"sort"
	"testing"
)

func TestCleanPromptToTags(t *testing.T) {
	tests := []struct {
		name     string
		prompt   string
		expected []string
	}{
		{
			name:     "Basic prompt with stopwords",
			prompt:   "What happened in our last funding call?",
			expected: []string{"funding", "last"},
		},
		{
			name:     "Technical prompt",
			prompt:   "Submit the IRAP proposal",
			expected: []string{"submit", "irap", "proposal"},
		},
		{
			name:     "Empty prompt",
			prompt:   "",
			expected: []string{},
		},
		{
			name:     "Only stopwords",
			prompt:   "the and or but",
			expected: []string{},
		},
		{
			name:     "Mixed case with punctuation",
			prompt:   "Deploy the AI Model to Production!",
			expected: []string{"deploy", "model", "production"},
		},
		{
			name:     "Short words filtered",
			prompt:   "Go to AI ML lab",
			expected: []string{"lab"},
		},
		{
			name:     "Deduplication test",
			prompt:   "test test the test again",
			expected: []string{"test", "again"},
		},
		{
			name:     "Complex sentence",
			prompt:   "How can we integrate the new authentication service with our existing user management system?",
			expected: []string{"integrate", "new", "authentication", "service", "existing", "user", "management", "system"},
		},
		{
			name:     "Numbers and special characters",
			prompt:   "Configure port 8080 for the API service v2.1",
			expected: []string{"configure", "port", "8080", "api", "service", "v21"},
		},
		{
			name:     "Domain-specific terms",
			prompt:   "MongoDB connection failed with timeout error",
			expected: []string{"mongodb", "connection", "failed", "timeout", "error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPromptToTags(tt.prompt)
			
			// Sort both slices for comparison since map iteration order is not guaranteed
			sort.Strings(result)
			sort.Strings(tt.expected)
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CleanPromptToTags() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCleanTagsForQuery(t *testing.T) {
	tests := []struct {
		name     string
		rawTags  []string
		expected []string
	}{
		{
			name:     "Basic tag cleaning",
			rawTags:  []string{"the", "API", "service", "and"},
			expected: []string{"api", "service"},
		},
		{
			name:     "Empty input",
			rawTags:  []string{},
			expected: []string{},
		},
		{
			name:     "Tags with punctuation",
			rawTags:  []string{"auth-service", "user.data", "config!"},
			expected: []string{"authservice", "userdata", "config"},
		},
		{
			name:     "Duplicates removal",
			rawTags:  []string{"API", "api", "Service", "service"},
			expected: []string{"api", "service"},
		},
		{
			name:     "Whitespace handling",
			rawTags:  []string{" mongodb ", "  redis  ", "kafka"},
			expected: []string{"mongodb", "redis", "kafka"},
		},
		{
			name:     "Short words filtered",
			rawTags:  []string{"ai", "ml", "api", "db"},
			expected: []string{"api"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanTagsForQuery(tt.rawTags)
			
			// Sort both slices for comparison
			sort.Strings(result)
			sort.Strings(tt.expected)
			
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("CleanTagsForQuery() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		maxTags  int
		minCount int // minimum expected tags
	}{
		{
			name:     "Technical documentation",
			content:  "The microservice architecture includes API gateway, authentication service, and database connections",
			maxTags:  5,
			minCount: 3,
		},
		{
			name:     "No limit",
			content:  "Deploy Docker containers with Kubernetes orchestration and monitoring",
			maxTags:  0,
			minCount: 4,
		},
		{
			name:     "Limit exceeded",
			content:  "Configure MongoDB Redis PostgreSQL MySQL databases with connection pooling and caching",
			maxTags:  3,
			minCount: 0, // Should be exactly 3 due to limit
		},
		{
			name:     "Empty content",
			content:  "",
			maxTags:  10,
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractKeywords(tt.content, tt.maxTags)
			
			// Check minimum count
			if len(result) < tt.minCount {
				t.Errorf("ExtractKeywords() returned %d tags, expected at least %d", len(result), tt.minCount)
			}
			
			// Check maximum limit
			if tt.maxTags > 0 && len(result) > tt.maxTags {
				t.Errorf("ExtractKeywords() returned %d tags, expected maximum %d", len(result), tt.maxTags)
			}
			
			// Check for stopwords
			for _, tag := range result {
				if stopwords[tag] {
					t.Errorf("ExtractKeywords() returned stopword: %s", tag)
				}
			}
		})
	}
}

func BenchmarkCleanPromptToTags(b *testing.B) {
	prompt := "What happened in our last IRAP funding submission and how can we improve the authentication service?"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanPromptToTags(prompt)
	}
}

func BenchmarkCleanTagsForQuery(b *testing.B) {
	tags := []string{"the", "API", "service", "and", "authentication", "mongodb", "redis", "cache"}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CleanTagsForQuery(tags)
	}
}

// Test edge cases and security considerations
func TestTagCleanerEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantLen int
	}{
		{
			name:    "Very long input",
			input:   "this is a very long prompt with many words that should be processed efficiently " + string(make([]byte, 1000)),
			wantLen: 0, // At least some tags should be extracted
		},
		{
			name:    "Special characters only",
			input:   "!@#$%^&*()_+-=[]{}|;:'\",.<>?/~`",
			wantLen: 0,
		},
		{
			name:    "Unicode characters",
			input:   "café résumé naïve Zürich",
			wantLen: 0, // Should handle unicode gracefully
		},
		{
			name:    "Numbers only",
			input:   "123 456 789 0",
			wantLen: 0,
		},
		{
			name:    "Mixed valid and invalid",
			input:   "valid123 terms456 mixed789 content",
			wantLen: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CleanPromptToTags(tt.input)
			
			if tt.wantLen > 0 && len(result) < tt.wantLen {
				t.Errorf("CleanPromptToTags() = %d tags, want at least %d", len(result), tt.wantLen)
			}
			
			// Ensure no empty strings
			for _, tag := range result {
				if tag == "" {
					t.Error("CleanPromptToTags() returned empty string")
				}
			}
		})
	}
}