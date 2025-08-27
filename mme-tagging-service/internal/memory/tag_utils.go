package memory

import (
	"strings"
	"time"
)

// NormalizeLabel normalizes a tag label (lowercase, trim, collapse spaces)
func NormalizeLabel(s string) string {
	// Convert to lowercase and trim whitespace
	normalized := strings.ToLower(strings.TrimSpace(s))
	// Replace multiple spaces with single space
	normalized = strings.Join(strings.Fields(normalized), " ")
	return normalized
}

// ToFlatTags converts structured tags to flat string array for backward compatibility
func ToFlatTags(tags []Tag) []string {
	if len(tags) == 0 {
		return []string{}
	}
	
	flatTags := make([]string, 0, len(tags))
	seen := make(map[string]bool)
	
	for _, tag := range tags {
		normalized := NormalizeLabel(tag.Label)
		if normalized != "" && !seen[normalized] {
			flatTags = append(flatTags, normalized)
			seen[normalized] = true
		}
	}
	
	return flatTags
}

// FromStringsToTags converts string array to structured tags with defaults
func FromStringsToTags(stringTags []string) []Tag {
	if len(stringTags) == 0 {
		return []Tag{}
	}
	
	tags := make([]Tag, 0, len(stringTags))
	now := time.Now()
	
	for _, label := range stringTags {
		normalized := NormalizeLabel(label)
		if normalized == "" {
			continue
		}
		
		tag := Tag{
			Label:      normalized,
			Origin:     "unknown",
			Scope:      "shared",
			Type:       "concept",
			Confidence: 0.6,
			UsageCount: 1,
			LastUsed:   now,
		}
		
		// Simple heuristics for type detection
		if containsActionWords(normalized) {
			tag.Type = "action"
		} else if containsObjectWords(normalized) {
			tag.Type = "object"
		} else if containsErrorWords(normalized) {
			tag.Type = "error"
		} else if containsStatusWords(normalized) {
			tag.Type = "status"
		}
		
		tags = append(tags, tag)
	}
	
	return tags
}

// containsActionWords checks if label contains action-related words
func containsActionWords(label string) bool {
	actionWords := []string{
		"submit", "create", "build", "implement", "deploy", "test", "review",
		"complete", "finish", "start", "begin", "launch", "release", "update",
		"fix", "resolve", "solve", "process", "handle", "manage", "execute",
		"run", "perform", "conduct", "carry", "out", "deliver", "provide",
	}
	
	labelLower := strings.ToLower(label)
	for _, word := range actionWords {
		if strings.Contains(labelLower, word) {
			return true
		}
	}
	return false
}

// containsObjectWords checks if label contains object-related words
func containsObjectWords(label string) bool {
	objectWords := []string{
		"form", "document", "file", "report", "proposal", "budget", "plan",
		"system", "application", "database", "api", "service", "module",
		"component", "interface", "model", "framework", "library", "tool",
		"platform", "environment", "configuration", "setting", "parameter",
	}
	
	labelLower := strings.ToLower(label)
	for _, word := range objectWords {
		if strings.Contains(labelLower, word) {
			return true
		}
	}
	return false
}

// containsErrorWords checks if label contains error-related words
func containsErrorWords(label string) bool {
	errorWords := []string{
		"error", "fail", "failure", "exception", "bug", "issue", "problem",
		"crash", "timeout", "invalid", "missing", "broken", "corrupt",
		"unavailable", "denied", "rejected", "cancelled", "aborted",
	}
	
	labelLower := strings.ToLower(label)
	for _, word := range errorWords {
		if strings.Contains(labelLower, word) {
			return true
		}
	}
	return false
}

// containsStatusWords checks if label contains status-related words
func containsStatusWords(label string) bool {
	statusWords := []string{
		"status", "state", "condition", "ready", "pending", "active",
		"inactive", "enabled", "disabled", "running", "stopped", "completed",
		"failed", "success", "approved", "rejected", "draft", "final",
		"published", "archived", "expired", "valid", "invalid",
	}
	
	labelLower := strings.ToLower(label)
	for _, word := range statusWords {
		if strings.Contains(labelLower, word) {
			return true
		}
	}
	return false
}
