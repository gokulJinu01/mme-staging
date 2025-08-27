package memory

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mme-tagging-service/internal/config"
	"mme-tagging-service/internal/feature"
	"mme-tagging-service/internal/http/middleware"
	"mme-tagging-service/internal/metrics"
)

// InjectRequest represents the request for memory injection
type InjectRequest struct {
	OrgID       string         `json:"orgId"`
	ProjectID   string         `json:"projectId,omitempty"`
	Prompt      string         `json:"prompt"`
	Limit       int            `json:"limit,omitempty"`
	Filters     *InjectFilters `json:"filters,omitempty"`
	TokenBudget int            `json:"tokenBudget,omitempty"`
}

// InjectResponse represents the response from memory injection
type InjectResponse struct {
	PackID      string                 `json:"packId"`
	SeedTags    []string               `json:"seedTags"`
	Bounds      map[string]interface{} `json:"bounds"`
	Filters     *InjectFilters         `json:"filters"`
	TokenBudget int                    `json:"tokenBudget"`
	TotalTokens int                    `json:"totalTokens"`
	Items       []InjectItem           `json:"items"`
	Rationale   InjectRationale        `json:"rationale"`
}

// InjectItem represents an item in the injection pack
type InjectItem struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Tags      []string   `json:"tags"`
	Excerpt   string     `json:"excerpt"`
	TokenCost int        `json:"tokenCost"`
	Score     ScoreParts `json:"score"`
}

// InjectRationale represents the rationale for the injection pack
type InjectRationale struct {
	Paths []RationalePath `json:"paths"`
	Notes []string        `json:"notes"`
}

// RationalePath represents a path in the tag graph
type RationalePath struct {
	To     string  `json:"to"`
	From   string  `json:"from"`
	Weight float64 `json:"weight"`
	Depth  int     `json:"depth"`
}

// HandleMemoryInject handles the POST /memory/inject endpoint
func HandleMemoryInject(c *fiber.Ctx) error {
	startTime := time.Now()

	var request InjectRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get orgID from JWT if not provided in request
	if request.OrgID == "" {
		request.OrgID = middleware.GetOrgID(c)
	}

	// Validate required fields
	if request.OrgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "orgId is required",
		})
	}

	if request.Prompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prompt is required",
		})
	}

	// Set defaults
	if request.Limit == 0 {
		request.Limit = 20
	}
	if request.TokenBudget == 0 {
		mmeConfig := config.GetMMEConfig()
		request.TokenBudget = mmeConfig.TokenBudget
	}

	// Check if propagation is enabled
	propagationEnabled := feature.IsPropagationEnabled(request.OrgID)

	// Build injection pack
	response, err := buildInjectionPack(c.Context(), request, propagationEnabled)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to build injection pack: %v", err),
		})
	}

	// Record latency for SLO monitoring
	duration := time.Since(startTime)
	if feature.IsSLOGuardEnabled(request.OrgID) {
		feature.GlobalSLOGuard.RecordLatency(request.OrgID, duration)
	}

	// Record metrics
	metrics.RecordInjectDuration(duration, "success")
	metrics.RecordPackItems(len(response.Items), "default", "success")

	return c.JSON(response)
}

// buildInjectionPack builds a token-budgeted injection pack
func buildInjectionPack(ctx context.Context, request InjectRequest, propagationEnabled bool) (*InjectResponse, error) {
	// Step 1: Extract seed tags from prompt
	allTags := CleanPromptToTags(request.Prompt)
	mmeConfig := config.GetMMEConfig()
	seedTags := boundedSeedTags(allTags, mmeConfig.MaxSeedTags)

	// Step 2: Perform bounded propagation to get activated tags (if enabled)
	var activatedTags []string
	var boundedConfig BoundedTagGraphConfig
	if propagationEnabled {
		boundedConfig = DefaultBoundedConfig()
		activatedTags = performBoundedPropagation(ctx, seedTags, boundedConfig)
	} else {
		// Fallback to exact tag matching only
		activatedTags = seedTags
		boundedConfig = DefaultBoundedConfig() // Still need config for querying
	}

	// Step 3: Query memory blocks with filters
	results, err := queryWithBoundedTags(ctx, request.OrgID, activatedTags, request.Limit*2, boundedConfig, request.Filters)
	if err != nil {
		return nil, err
	}

	// Step 4: Score and rank memory blocks
	// Convert activatedTags slice to map for scoring
	activatedTagsMap := make(map[string]float64)
	for _, tag := range activatedTags {
		activatedTagsMap[tag] = 1.0 // Default activation score
	}
	scoredBlocks := scoreAndRankBlocks(results, activatedTagsMap)

	// Step 5: Build token-budgeted pack with deduplication
	packItems, totalTokens := buildTokenBudgetedPack(scoredBlocks, request.TokenBudget)

	// Step 6: Build rationale
	rationale := buildRationale(seedTags, boundedConfig, packItems)

	// Step 7: Build response
	response := &InjectResponse{
		PackID:   uuid.New().String(),
		SeedTags: seedTags,
		Bounds: map[string]interface{}{
			"M":     boundedConfig.MaxEdgesPerTag,
			"D":     boundedConfig.MaxDepth,
			"B":     boundedConfig.BeamWidth,
			"alpha": mmeConfig.DecayAlpha,
			"theta": mmeConfig.MinActivation,
		},
		Filters:     request.Filters,
		TokenBudget: request.TokenBudget,
		TotalTokens: totalTokens,
		Items:       packItems,
		Rationale:   rationale,
	}

	return response, nil
}

// scoreAndRankBlocks scores and ranks memory blocks
func scoreAndRankBlocks(results []SemanticSearchResult, activatedTags map[string]float64) []ScoredBlock {
	var scoredBlocks []ScoredBlock

	for _, result := range results {
		// Convert activated tags map to the format expected by scoring
		act := make(map[string]float64)
		for tag, activation := range activatedTags {
			act[tag] = activation
		}

		// Score the block
		score := ScoreBlockWithDefaults(result.MemoryBlock, act, nil)

		// Estimate token cost
		tokenCost := estimateTokens(result.MemoryBlock.Content)

		scoredBlocks = append(scoredBlocks, ScoredBlock{
			MemoryBlock: result.MemoryBlock,
			Score:       score,
			TokenCost:   tokenCost,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scoredBlocks, func(i, j int) bool {
		return scoredBlocks[i].Score.Total > scoredBlocks[j].Score.Total
	})

	return scoredBlocks
}

// ScoredBlock represents a memory block with its score and token cost
type ScoredBlock struct {
	MemoryBlock MemoryBlock `json:"memoryBlock"`
	Score       ScoreParts  `json:"score"`
	TokenCost   int         `json:"tokenCost"`
}

// buildTokenBudgetedPack builds a token-budgeted pack with deduplication
func buildTokenBudgetedPack(scoredBlocks []ScoredBlock, tokenBudget int) ([]InjectItem, int) {
	var packItems []InjectItem
	var totalTokens int
	seen := make(map[string]bool)

	for _, block := range scoredBlocks {
		// Check if we can add this item within budget
		if totalTokens+block.TokenCost > tokenBudget {
			continue
		}

		// Check for duplicates
		dedupKey := generateDedupKey(block.MemoryBlock)
		if seen[dedupKey] {
			continue
		}
		seen[dedupKey] = true

		// Create excerpt (first 500 chars or summary)
		excerpt := block.MemoryBlock.Content
		if len(excerpt) > 500 {
			excerpt = excerpt[:500] + "..."
		}

		// Create title (use first line or generate from tags)
		title := generateTitle(block.MemoryBlock)

		// Convert structured tags to flat tags for InjectItem
		flatTags := ToFlatTags(block.MemoryBlock.Tags)
		
		item := InjectItem{
			ID:        block.MemoryBlock.ID,
			Title:     title,
			Tags:      flatTags,
			Excerpt:   excerpt,
			TokenCost: block.TokenCost,
			Score:     block.Score,
		}

		packItems = append(packItems, item)
		totalTokens += block.TokenCost
	}

	return packItems, totalTokens
}

// generateDedupKey generates a deduplication key for a memory block
func generateDedupKey(block MemoryBlock) string {
	// Use normalized title + content hash for deduplication
	normalizedTitle := strings.ToLower(strings.TrimSpace(block.Content))
	hash := md5.Sum([]byte(normalizedTitle))
	return hex.EncodeToString(hash[:])
}

// generateTitle generates a title for a memory block
func generateTitle(block MemoryBlock) string {
	// Use first line if it's short enough
	lines := strings.Split(block.Content, "\n")
	if len(lines) > 0 && len(lines[0]) <= 100 {
		return strings.TrimSpace(lines[0])
	}

	// Generate from tags
	if len(block.Tags) > 0 {
		flatTags := ToFlatTags(block.Tags)
		return strings.Join(flatTags[:min(3, len(flatTags))], " - ")
	}

	// Fallback
	return "Memory Block"
}

// buildRationale builds the rationale for the injection pack
func buildRationale(seedTags []string, config BoundedTagGraphConfig, items []InjectItem) InjectRationale {
	var paths []RationalePath
	var notes []string

	// Add seed tags note
	if len(seedTags) > 0 {
		notes = append(notes, fmt.Sprintf("selected due to high activation on tags: %s", strings.Join(seedTags, ", ")))
	}

	// Add dominant edges (simplified for now)
	if len(items) > 0 {
		// Find the highest scoring item and add its dominant tags
		topItem := items[0]
		if len(topItem.Tags) > 0 {
			dominantTags := topItem.Tags[:min(3, len(topItem.Tags))]
			notes = append(notes, fmt.Sprintf("top result activated by tags: %s", strings.Join(dominantTags, ", ")))
		}
	}

	// Add sample paths (simplified)
	for i, tag := range seedTags {
		if i < 3 { // Limit to 3 paths
			paths = append(paths, RationalePath{
				To:     tag,
				From:   "prompt",
				Weight: 1.0,
				Depth:  0,
			})
		}
	}

	return InjectRationale{
		Paths: paths,
		Notes: notes,
	}
}

// estimateTokens estimates the token count for a string
func estimateTokens(text string) int {
	// Simple heuristic: ~4 characters per token
	// This is a rough approximation; in production, use a proper tokenizer
	return len(text) / 4
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
