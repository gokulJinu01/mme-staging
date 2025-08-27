package memory

import (
	"strings"
)

// PerformSemanticSearch performs semantic search on memory blocks using bounded tag graph propagation
func PerformSemanticSearch(request SemanticSearchRequest) ([]SemanticSearchResult, error) {
	// Use bounded tag graph search for better performance and controlled propagation
	config := DefaultBoundedConfig()

	// Set minimum score from request if provided
	if request.MinScore > 0 {
		config.MinScoreThreshold = request.MinScore
	}

	return BoundedTagGraphSearch(request, config)
}

// extractKeywords extracts important keywords from a query string
func extractKeywords(query string) []string {
	// Simple keyword extraction - split by spaces and remove common words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "this": true, "that": true,
		"these": true, "those": true, "i": true, "you": true, "he": true, "she": true,
		"it": true, "we": true, "they": true, "me": true, "him": true, "her": true,
		"us": true, "them": true, "my": true, "your": true, "his": true,
		"its": true, "our": true, "their": true,
	}

	words := strings.Fields(strings.ToLower(query))
	var keywords []string

	for _, word := range words {
		// Clean word (remove punctuation)
		cleaned := strings.Trim(word, ".,!?;:")
		if len(cleaned) > 2 && !stopWords[cleaned] {
			keywords = append(keywords, cleaned)
		}
	}

	return keywords
}
