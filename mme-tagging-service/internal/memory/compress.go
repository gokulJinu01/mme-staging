package memory

import (
	"strings"
)

func CompressBlock(block MemoryBlock) MemoryBlock {
	if len(block.Content) <= 400 {
		return block
	}

	// Extract key sentences and important content
	content := block.Content
	
	// Simple summarization: extract first and last sentences, plus any sentences with keywords
	sentences := splitIntoSentences(content)
	if len(sentences) <= 2 {
		return block
	}

	var summary []string
	
	// Always include first sentence
	summary = append(summary, sentences[0])
	
	// Include sentences with important keywords
	keywords := []string{"error", "failed", "success", "completed", "important", "warning", "critical"}
	for _, sentence := range sentences[1:len(sentences)-1] {
		if containsKeywords(sentence, keywords) && len(strings.Join(summary, " ")) < 300 {
			summary = append(summary, sentence)
		}
	}
	
	// Always include last sentence if we have room
	lastSentence := sentences[len(sentences)-1]
	if len(strings.Join(summary, " ")+lastSentence) <= 350 {
		summary = append(summary, lastSentence)
	}
	
	block.Content = "[SUMMARY] " + strings.Join(summary, " ")
	return block
}

func splitIntoSentences(text string) []string {
	// Simple sentence splitting on periods, exclamation marks, and question marks
	text = strings.ReplaceAll(text, "! ", "!|")
	text = strings.ReplaceAll(text, "? ", "?|")
	text = strings.ReplaceAll(text, ". ", ".|")
	
	sentences := strings.Split(text, "|")
	var result []string
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 0 {
			result = append(result, sentence)
		}
	}
	return result
}

func containsKeywords(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}
