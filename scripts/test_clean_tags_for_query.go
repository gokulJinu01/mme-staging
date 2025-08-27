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
}

func main() {
	testTags := []string{"implemented_machine", "used_python", "meeting", "project"}
	
	for _, tag := range testTags {
		result := cleanTagsForQuery([]string{tag})
		fmt.Printf("Original: [%s] -> Cleaned: %v\n", tag, result)
	}
}

func cleanTagsForQuery(rawTags []string) []string {
	if len(rawTags) == 0 {
		return rawTags
	}

	cleanedTags := make([]string, 0, len(rawTags))
	tagSet := make(map[string]bool)

	for _, tag := range rawTags {
		cleaned := strings.ToLower(strings.TrimSpace(tag))

		if cleaned == "" {
			continue
		}

		// Skip stopwords and very short words, but allow meaningful numbers and keep useful tags
		if !stopwords[cleaned] && (len(cleaned) > 1) && !tagSet[cleaned] {
			cleanedTags = append(cleanedTags, cleaned)
			tagSet[cleaned] = true
		}
	}

	return cleanedTags
}
