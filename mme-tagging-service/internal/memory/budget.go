package memory

func EstimateTokens(content string) int {
	// Approximate: 1 token ≈ 4 chars
	return len(content) / 4
}

func CalculateMemorySize(blocks []MemoryBlock) int {
	total := 0
	for _, b := range blocks {
		total += EstimateTokens(b.Content)
	}
	return total
}
