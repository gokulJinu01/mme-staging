package memory

import (
	"sort"
)

const MaxTokenBudget = 2048 // tokens allowed for memory

func PromoteMemory(blocks []MemoryBlock, tags []string, goal string) []MemoryBlock {

	// Score each block
	type scoredBlock struct {
		Block MemoryBlock
		Score float64
	}

	scored := []scoredBlock{}
	for _, b := range blocks {
		// Create activation map from tags
		act := make(map[string]float64)
		for _, tag := range tags {
			act[tag] = 1.0
		}
		
		scoreParts := ScoreBlockWithDefaults(b, act, nil)
		scored = append(scored, scoredBlock{Block: b, Score: scoreParts.Total})
	}

	// Sort by score DESC
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	// Pick blocks until token budget is exhausted
	selected := []MemoryBlock{}
	usedTokens := 0

	for _, sb := range scored {
		size := EstimateTokens(sb.Block.Content)
		if usedTokens+size <= MaxTokenBudget {
			selected = append(selected, sb.Block)
			usedTokens += size
		} else if size > 256 {
			// If too big but important, compress it
			compressed := CompressBlock(sb.Block)
			size = EstimateTokens(compressed.Content)
			if usedTokens+size <= MaxTokenBudget {
				selected = append(selected, compressed)
				usedTokens += size
			}
		}
	}

	return selected
}
