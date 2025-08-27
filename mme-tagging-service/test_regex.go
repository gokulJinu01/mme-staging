package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	// Test the regex pattern from CleanPromptToTags
	testStrings := []string{
		"implemented_machine",
		"used_python", 
		"meeting",
		"project",
		"timeline_updates",
		"test-tag",
		"test.tag",
		"test@tag",
	}
	
	re := regexp.MustCompile(`[^\w\s]`)
	
	for _, str := range testStrings {
		cleaned := re.ReplaceAllString(strings.ToLower(str), " ")
		fmt.Printf("Original: '%s' -> Cleaned: '%s'\n", str, cleaned)
	}
}
