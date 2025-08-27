//go:build debug_real
// +build debug_real

package main

import (
	"fmt"
	"strings"
)

var stopwords = map[string]bool{
	"what": true, "is": true, "the": true, "and": true, "in": true, "of": true,
	"on": true, "to": true, "was": true, "we": true, "you": true, "i": true,
	"at": true, "for": true, "a": true, "an": true, "this": true, "that": true,
	"have": true, "has": true, "with": true, "from": true, "by": true,
	"been": true, "as": true, "it": true, "be": true, "are": true,
	"how": true, "our": true, "your": true, "their": true, "its": true,
	"or": true, "but": true, "not": true, "so": true, "if": true, "can": true,
	"will": true, "would": true, "could": true, "should": true, "may": true,
	"might": true, "do": true, "does": true, "did": true, "get": true,
	"got": true, "go": true, "went": true, "come": true, "came": true,
	"see": true, "saw": true, "know": true, "knew": true, "think": true,
	"thought": true, "say": true, "said": true, "tell": true, "told": true,
	"make": true, "made": true, "take": true, "took": true, "give": true,
	"gave": true, "put": true, "set": true, "let": true, "use": true,
	"used": true, "find": true, "found": true, "work": true, "worked": true,
}

func main() {
	// Test the exact tags we're having issues with
	testInput := "implemented_machine"
	rawTagsList := strings.Split(testInput, ",")
	tags := CleanTagsForQuery(rawTagsList)
	
	fmt.Printf("Input: %s\n", testInput)
	fmt.Printf("RawTagsList: %v\n", rawTagsList)
	fmt.Printf("CleanedTags: %v\n", tags)
	
	// Test multiple tags
	testInput2 := "implemented_machine,simple,test_underscore"
	rawTagsList2 := strings.Split(testInput2, ",")
	tags2 := CleanTagsForQuery(rawTagsList2)
	
	fmt.Printf("\nInput2: %s\n", testInput2)
	fmt.Printf("RawTagsList2: %v\n", rawTagsList2)
	fmt.Printf("CleanedTags2: %v\n", tags2)
}

func containsDangerousPatterns(word string) bool {
	dangerousPatterns := []string{
		"script", "javascript", "eval", "function",
		"alert", "document", "window", "location",
		"onload", "onerror", "onclick", "onmouse",
		"where", "eval", "mapreduce",
	}

	wordLower := strings.ToLower(word)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(wordLower, pattern) {
			return true
		}
	}
	return false
}

func isMeaningfulNumber(s string) bool {
	// Simple version for testing
	return false
}

func CleanTagsForQuery(rawTags []string) []string {
	if len(rawTags) == 0 {
		return rawTags
	}

	cleanedTags := make([]string, 0, len(rawTags))
	tagSet := make(map[string]bool)

	for _, tag := range rawTags {
		// Clean each tag
		cleaned := strings.ToLower(strings.TrimSpace(tag))

		// Skip if empty after trimming
		if cleaned == "" {
			continue
		}

		// For queries, we want more lenient cleaning to match stored tags
		// Only remove dangerous/problematic characters, keep underscores, hyphens, etc.
		if containsDangerousPatterns(cleaned) {
			continue
		}

		// Skip stopwords and very short words, but allow meaningful numbers and keep useful tags
		if !stopwords[cleaned] && (len(cleaned) > 1 || isMeaningfulNumber(cleaned)) && !tagSet[cleaned] {
			cleanedTags = append(cleanedTags, cleaned)
			tagSet[cleaned] = true
		}
	}

	return cleanedTags
}
