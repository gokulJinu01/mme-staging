package memory

import (
	"regexp"
	"strings"
)

// Stopword list — keep it lean for high performance
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
	"look": true, "looked": true, "seem": true, "seemed": true, "feel": true,
	"felt": true, "try": true, "tried": true, "ask": true, "asked": true,
	"need": true, "needed": true, "want": true, "wanted": true, "like": true,
	"liked": true, "help": true, "helped": true, "show": true, "showed": true,
	"turn": true, "turned": true, "start": true, "started": true, "call": true,
	"called": true, "move": true, "moved": true, "live": true, "lived": true,
	"believe": true, "believed": true, "hold": true, "held": true, "bring": true,
	"brought": true, "happen": true, "happened": true, "write": true, "wrote": true,
	"sit": true, "sat": true, "stand": true, "stood": true, "lose": true,
	"lost": true, "pay": true, "paid": true, "meet": true, "met": true,
	"include": true, "included": true, "continue": true, "continued": true,
	"learn": true, "learned": true, "change": true, "changed": true,
	"lead": true, "led": true, "understand": true, "understood": true,
	"watch": true, "watched": true, "follow": true, "followed": true,
	"stop": true, "stopped": true, "create": true, "created": true,
	"speak": true, "spoke": true, "read": true, "allow": true, "allowed": true,
	"add": true, "added": true, "spend": true, "spent": true, "grow": true,
	"grew": true, "open": true, "opened": true, "walk": true, "walked": true,
	"win": true, "won": true, "offer": true, "offered": true, "remember": true,
	"remembered": true, "love": true, "loved": true, "consider": true,
	"considered": true, "appear": true, "appeared": true, "buy": true,
	"bought": true, "wait": true, "waited": true, "serve": true, "served": true,
	"die": true, "died": true, "send": true, "sent": true, "expect": true,
	"expected": true, "build": true, "built": true, "stay": true, "stayed": true,
	"fall": true, "fell": true, "cut": true, "reach": true, "reached": true,
	"kill": true, "killed": true, "remain": true, "remained": true,
}

// CleanPromptToTags processes a raw prompt into deduplicated clean tags
func CleanPromptToTags(prompt string) []string {
	if prompt == "" {
		return []string{}
	}

	// Security: Limit input length to prevent DoS
	if len(prompt) > 50000 {
		prompt = prompt[:50000]
	}

	// Remove HTML/XML tags first
	htmlRe := regexp.MustCompile(`<[^>]*>`)
	prompt = htmlRe.ReplaceAllString(prompt, " ")

	// Preserve version numbers before general punctuation removal
	versionRe := regexp.MustCompile(`v\d+\.\d+`)
	versions := versionRe.FindAllString(prompt, -1)

	// Remove punctuation and normalize to lowercase, but preserve numbers
	re := regexp.MustCompile(`[^\w\s]`)
	cleaned := re.ReplaceAllString(strings.ToLower(prompt), " ")

	// Tokenize into words
	words := strings.Fields(cleaned)

	// Use map for deduplication
	tagSet := make(map[string]bool)

	// Add version numbers first
	for _, version := range versions {
		cleanVersion := strings.ToLower(version)
		// Remove dots from version numbers for consistency
		cleanVersion = strings.ReplaceAll(cleanVersion, ".", "")
		if len(cleanVersion) > 0 {
			tagSet[cleanVersion] = true
		}
	}

	for _, word := range words {
		// Security filters
		if len(word) > 50 { // Prevent extremely long words
			continue
		}
		if containsDangerousPatterns(word) {
			continue
		}

		// Filter out stopwords and very short words, but allow meaningful numbers
		if !stopwords[word] && (len(word) > 2 || isMeaningfulNumber(word)) {
			tagSet[word] = true
		}
	}

	// Convert map to slice
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags
}

// isNumeric checks if a string contains only digits
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// containsDangerousPatterns checks for dangerous content in words
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

// isMeaningfulNumber checks if a string is a meaningful number (not just a single digit)
func isMeaningfulNumber(s string) bool {
	if !isNumeric(s) {
		return false
	}
	// Allow multi-digit numbers and common meaningful single digits
	if len(s) > 1 {
		return true
	}
	// Filter out most single digits except common meaningful ones
	meaningfulSingles := map[string]bool{
		"0": true, "1": false, "2": false, "3": false, "4": false,
		"5": false, "6": false, "7": false, "8": false, "9": false,
	}
	return meaningfulSingles[s]
}

// CleanTagsForQuery enhances the existing query system to clean input tags
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

// ExtractKeywords extracts meaningful keywords from content for automatic tagging
func ExtractKeywords(content string, maxTags int) []string {
	tags := CleanPromptToTags(content)

	// Limit the number of tags if specified
	if maxTags > 0 && len(tags) > maxTags {
		tags = tags[:maxTags]
	}

	return tags
}
