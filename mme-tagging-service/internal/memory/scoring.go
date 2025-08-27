package memory

import (
	"math"
	"time"

	"mme-tagging-service/internal/config"
)

// ScoreParts represents the breakdown of a memory block's score
type ScoreParts struct {
	Activation       float64 `json:"activation"`
	Recency          float64 `json:"recency"`
	Importance       float64 `json:"importance"`
	StatusBonus      float64 `json:"statusBonus"`
	DiversityPenalty float64 `json:"diversityPenalty"`
	Total            float64 `json:"total"`
}

// ScoreBlock calculates the score for a memory block
func ScoreBlock(block MemoryBlock, act map[string]float64, now time.Time, tauDays int, lambda float64, already []MemoryBlock) ScoreParts {
	parts := ScoreParts{}

	// Activation = max_{t in block.Tags} act[t]
	parts.Activation = 0.0
	for _, tag := range block.Tags {
		// Normalize tag label to match activation map keys
		normalizedLabel := NormalizeLabel(tag.Label)
		if activation, exists := act[normalizedLabel]; exists && activation > parts.Activation {
			parts.Activation = activation
		}
	}

	// Recency = exp(-(now - block.CreatedAt) / (tauDays * 24h))
	timeDiff := now.Sub(block.CreatedAt)
	tauSeconds := float64(tauDays) * 24 * 60 * 60
	parts.Recency = math.Exp(-timeDiff.Seconds() / tauSeconds)

	// Importance = float64(block.Priority) or Confidence if present
	if block.Priority > 0 {
		parts.Importance = float64(block.Priority)
	} else if block.Confidence > 0 {
		parts.Importance = block.Confidence
	} else {
		parts.Importance = 1.0 // Default importance
	}

	// StatusBonus = map[draft:0, submitted:0.5, completed:1.0]
	switch block.Status {
	case "completed":
		parts.StatusBonus = 1.0
	case "submitted":
		parts.StatusBonus = 0.5
	case "draft":
		parts.StatusBonus = 0.0
	default:
		parts.StatusBonus = 0.0
	}

	// DiversityPenalty: lambda * max_{m' in already} Jaccard(block.Tags, m'.Tags)
	parts.DiversityPenalty = 0.0
	if len(already) > 0 {
		maxJaccard := 0.0
		blockFlatTags := ToFlatTags(block.Tags)
		for _, existing := range already {
			existingFlatTags := ToFlatTags(existing.Tags)
			jaccard := calculateJaccardSimilarity(blockFlatTags, existingFlatTags)
			if jaccard > maxJaccard {
				maxJaccard = jaccard
			}
		}
		parts.DiversityPenalty = lambda * maxJaccard
	}

	// Total = β1*Activation + β2*Recency + β3*Importance + β4*StatusBonus - DiversityPenalty
	// Using β1=1.0, β2=0.5, β3=0.25, β4=0.25 as constants
	beta1 := 1.0
	beta2 := 0.5
	beta3 := 0.25
	beta4 := 0.25

	parts.Total = beta1*parts.Activation +
		beta2*parts.Recency +
		beta3*parts.Importance +
		beta4*parts.StatusBonus -
		parts.DiversityPenalty

	// Ensure total is non-negative
	if parts.Total < 0 {
		parts.Total = 0
	}

	return parts
}

// calculateJaccardSimilarity calculates the Jaccard similarity between two tag sets
func calculateJaccardSimilarity(tags1, tags2 []string) float64 {
	if len(tags1) == 0 && len(tags2) == 0 {
		return 0.0
	}

	// Convert to sets for easier intersection/union calculation
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, tag := range tags1 {
		set1[tag] = true
	}
	for _, tag := range tags2 {
		set2[tag] = true
	}

	// Calculate intersection
	intersection := 0
	for tag := range set1 {
		if set2[tag] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// ScoreBlockWithDefaults calculates the score using default configuration values
func ScoreBlockWithDefaults(block MemoryBlock, act map[string]float64, already []MemoryBlock) ScoreParts {
	mmeConfig := config.GetMMEConfig()
	now := time.Now()

	return ScoreBlock(block, act, now, mmeConfig.RecencyTauDays, mmeConfig.DiversityLambda, already)
}
