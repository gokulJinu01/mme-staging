package memory

import (
	"context"
	"sort"
	"time"

	"mme-tagging-service/internal/config"
	"mme-tagging-service/internal/db"

	"go.mongodb.org/mongo-driver/bson"
)

// BoundedTagGraphConfig defines the bounded parameters for tag graph propagation
type BoundedTagGraphConfig struct {
	MaxEdgesPerTag    int     // M: Maximum edges per tag (default: 32)
	MaxDepth          int     // D: Maximum propagation depth (default: 2)
	BeamWidth         int     // B: Beam width for top activations (default: 128)
	MaxSeedTags       int     // |S|: Maximum seed tags from prompt (default: 5)
	MinScoreThreshold float64 // Minimum score threshold for relevance
}

// InjectFilters defines filters for memory injection queries
type InjectFilters struct {
	OrgID     string `json:"orgId"`
	ProjectID string `json:"projectId,omitempty"`
	Section   string `json:"section,omitempty"`
	Status    string `json:"status,omitempty"` // draft|submitted|completed
	SinceISO  string `json:"since,omitempty"`  // ISO8601; default 90d
}

// DefaultBoundedConfig returns the bounded tag graph configuration from environment
func DefaultBoundedConfig() BoundedTagGraphConfig {
	mmeConfig := config.GetMMEConfig()
	return BoundedTagGraphConfig{
		MaxEdgesPerTag:    mmeConfig.MaxEdgesPerTag,
		MaxDepth:          mmeConfig.MaxDepth,
		BeamWidth:         mmeConfig.BeamWidth,
		MaxSeedTags:       mmeConfig.MaxSeedTags,
		MinScoreThreshold: mmeConfig.MinActivation,
	}
}

// BoundedTagGraphSearch performs bounded tag graph propagation search
func BoundedTagGraphSearch(request SemanticSearchRequest, config BoundedTagGraphConfig) ([]SemanticSearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Step 1: Extract bounded seed tags (|S| parameter)
	allTags := CleanPromptToTags(request.Query)
	seedTags := boundedSeedTags(allTags, config.MaxSeedTags)

	// Step 2: Perform bounded tag graph propagation
	activatedTags := performBoundedPropagation(ctx, seedTags, config)

	// Step 3: Query memory blocks with bounded tag set
	results, err := queryWithBoundedTags(ctx, request.UserID, activatedTags, request.Limit, config, nil)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// boundedSeedTags limits the number of seed tags (|S| parameter)
func boundedSeedTags(allTags []string, maxSeedTags int) []string {
	if len(allTags) <= maxSeedTags {
		return allTags
	}

	// Sort by length (prefer longer, more specific tags)
	sort.Slice(allTags, func(i, j int) bool {
		return len(allTags[i]) > len(allTags[j])
	})

	return allTags[:maxSeedTags]
}

// performBoundedPropagation performs bounded tag graph propagation (M, D, B parameters)
func performBoundedPropagation(ctx context.Context, seedTags []string, config BoundedTagGraphConfig) []string {
	activatedTags := make(map[string]float64)

	// Initialize with seed tags
	for _, tag := range seedTags {
		activatedTags[tag] = 1.0
	}

	// Perform bounded depth propagation (D parameter)
	for depth := 0; depth < config.MaxDepth; depth++ {
		nextLevel := make(map[string]float64)

		// For each activated tag, find related tags (M parameter)
		for tag, activation := range activatedTags {
			relatedTags := findRelatedTags(ctx, tag, config.MaxEdgesPerTag)

			// Propagate activation to related tags
			for relatedTag, similarity := range relatedTags {
				newActivation := activation * similarity * 0.8 // Decay factor
				if newActivation > config.MinScoreThreshold {
					nextLevel[relatedTag] = max(nextLevel[relatedTag], newActivation)
				}
			}
		}

		// Merge with current level
		for tag, activation := range nextLevel {
			activatedTags[tag] = max(activatedTags[tag], activation)
		}
	}

	// Apply beam width limit (B parameter)
	return applyBeamWidth(activatedTags, config.BeamWidth)
}

// findRelatedTags finds related tags for a given tag using the edge store (M parameter)
func findRelatedTags(ctx context.Context, tag string, maxEdges int) map[string]float64 {
	relatedTags := make(map[string]float64)

	// Get related tags using the new edge system
	relatedTagList, err := GetRelatedTags(ctx, tag, maxEdges)
	if err != nil {
		// Fallback to aggregation-based approach
		return findRelatedTagsAggregation(ctx, tag, maxEdges)
	}

	// Convert to map with default weight
	for _, relatedTag := range relatedTagList {
		relatedTags[relatedTag] = 0.5 // Default weight
	}

	return relatedTags
}

// findRelatedTagsAggregation is the fallback method using aggregation
func findRelatedTagsAggregation(ctx context.Context, tag string, maxEdges int) map[string]float64 {
	relatedTags := make(map[string]float64)

	// Query memory blocks containing this tag
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"tags": bson.M{"$in": []string{tag}},
			},
		},
		{
			"$unwind": "$tags",
		},
		{
			"$group": bson.M{
				"_id":       "$tags",
				"count":     bson.M{"$sum": 1},
				"totalDocs": bson.M{"$sum": 1},
			},
		},
		{
			"$sort": bson.M{"count": -1},
		},
		{
			"$limit": int64(maxEdges),
		},
	}

	cursor, err := db.MemoryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return relatedTags
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			continue
		}

		relatedTag, ok := result["_id"].(string)
		if !ok || relatedTag == tag {
			continue
		}

		count, _ := result["count"].(int32)
		totalDocs, _ := result["totalDocs"].(int32)

		// Calculate similarity score
		similarity := float64(count) / float64(totalDocs)
		if similarity > 0.1 { // Minimum similarity threshold
			relatedTags[relatedTag] = similarity
		}
	}

	return relatedTags
}

// applyBeamWidth applies beam width limit to activated tags (B parameter)
func applyBeamWidth(activatedTags map[string]float64, beamWidth int) []string {
	if len(activatedTags) <= beamWidth {
		// Convert map to slice
		result := make([]string, 0, len(activatedTags))
		for tag := range activatedTags {
			result = append(result, tag)
		}
		return result
	}

	// Sort by activation score and take top B
	type tagScore struct {
		tag   string
		score float64
	}

	var tagScores []tagScore
	for tag, score := range activatedTags {
		tagScores = append(tagScores, tagScore{tag, score})
	}

	sort.Slice(tagScores, func(i, j int) bool {
		return tagScores[i].score > tagScores[j].score
	})

	result := make([]string, beamWidth)
	for i := 0; i < beamWidth && i < len(tagScores); i++ {
		result[i] = tagScores[i].tag
	}

	return result
}

// queryWithBoundedTags queries memory blocks using bounded tag set with filters
func queryWithBoundedTags(ctx context.Context, userID string, boundedTags []string, limit int, config BoundedTagGraphConfig, filters *InjectFilters) ([]SemanticSearchResult, error) {
	if len(boundedTags) == 0 {
		return []SemanticSearchResult{}, nil
	}

	// Build match filter - use tagsFlat for string matching
	matchFilter := bson.M{
		"userId":   userID,
		"tagsFlat": bson.M{"$in": boundedTags},
	}

	// Apply additional filters if provided
	if filters != nil {
		if filters.OrgID != "" {
			matchFilter["orgId"] = filters.OrgID
		}
		if filters.ProjectID != "" {
			matchFilter["projectId"] = filters.ProjectID
		}
		if filters.Section != "" {
			matchFilter["section"] = filters.Section
		}
		if filters.Status != "" {
			matchFilter["status"] = filters.Status
		}
		if filters.SinceISO != "" {
			// Parse ISO8601 date and add to filter
			if sinceTime, err := time.Parse(time.RFC3339, filters.SinceISO); err == nil {
				matchFilter["createdAt"] = bson.M{"$gte": sinceTime}
			}
		}
	}

	// Build aggregation pipeline for bounded tag search
	pipeline := []bson.M{
		{
			"$match": matchFilter,
		},
		{
			"$addFields": bson.M{
				"tagOverlap": bson.M{
					"$size": bson.M{
						"$setIntersection": []interface{}{"$tagsFlat", boundedTags},
					},
				},
				"totalTags": bson.M{"$size": "$tagsFlat"},
			},
		},
		{
			"$addFields": bson.M{
				"score": bson.M{
					"$divide": []interface{}{
						"$tagOverlap",
						bson.M{
							"$max": []interface{}{
								"$totalTags",
								len(boundedTags),
								1,
							},
						},
					},
				},
			},
		},
		{
			"$match": bson.M{
				"score": bson.M{"$gte": config.MinScoreThreshold},
			},
		},
		{
			"$sort": bson.M{"score": -1},
		},
		{
			"$limit": int64(limit),
		},
	}

	cursor, err := db.MemoryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []SemanticSearchResult
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			continue
		}

		score, _ := doc["score"].(float64)

		var block MemoryBlock
		blockData, _ := bson.Marshal(doc)
		if err := bson.Unmarshal(blockData, &block); err != nil {
			continue
		}

		relevance := "low"
		if score >= 0.7 {
			relevance = "high"
		} else if score >= 0.4 {
			relevance = "medium"
		}

		results = append(results, SemanticSearchResult{
			MemoryBlock: block,
			Score:       score,
			Relevance:   relevance,
		})
	}

	return results, nil
}

// max returns the maximum of two float64 values
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// GetMMEConfig returns the current MME configuration
func GetMMEConfig() map[string]interface{} {
	config := DefaultBoundedConfig()
	return map[string]interface{}{
		"maxEdgesPerTag":    config.MaxEdgesPerTag,
		"maxDepth":          config.MaxDepth,
		"beamWidth":         config.BeamWidth,
		"maxSeedTags":       config.MaxSeedTags,
		"minScoreThreshold": config.MinScoreThreshold,
	}
}
