package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	testTags := []string{"implemented_machine", "used_python", "meeting"}
	
	for _, tag := range testTags {
		cleaned := cleanTag(tag)
		fmt.Printf("Original: %s -> Cleaned: %s\n", tag, cleaned)
	}
}

func cleanTag(tag string) string {
	cleaned := strings.ToLower(strings.TrimSpace(tag))
	re := regexp.MustCompile(`[^\w\s]`)
	cleaned = re.ReplaceAllString(cleaned, " ")
	return cleaned
}
