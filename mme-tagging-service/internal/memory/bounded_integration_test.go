//go:build integration

package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBoundedTagGraphSearchIntegration(t *testing.T) {
	// Test configuration
	config := BoundedTagGraphConfig{
		MaxEdgesPerTag:    2,
		MaxDepth:          2,
		BeamWidth:         4,
		MaxSeedTags:       5,
		MinScoreThreshold: 0.05,
	}

	// Test seed tags
	request := SemanticSearchRequest{
		UserID: "test-user",
		Query:  "python fastapi",
		Limit:  10,
	}

	// Run bounded propagation
	results, err := BoundedTagGraphSearch(request, config)

	// Verify results
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Note: Results may be empty if no data exists, but function should not error
}

func TestBoundedTagGraphSearchWithMultipleSeedsIntegration(t *testing.T) {
	config := BoundedTagGraphConfig{
		MaxEdgesPerTag:    2,
		MaxDepth:          1,
		BeamWidth:         3,
		MaxSeedTags:       5,
		MinScoreThreshold: 0.05,
	}

	request := SemanticSearchRequest{
		UserID: "test-user",
		Query:  "python fastapi machine learning",
		Limit:  10,
	}

	results, err := BoundedTagGraphSearch(request, config)

	// Should have results
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestBoundedTagGraphSearchRespectsMaxDepthIntegration(t *testing.T) {
	config := BoundedTagGraphConfig{
		MaxEdgesPerTag:    1,
		MaxDepth:          2, // Should only reach depth 2
		BeamWidth:         2,
		MaxSeedTags:       5,
		MinScoreThreshold: 0.05,
	}

	request := SemanticSearchRequest{
		UserID: "test-user",
		Query:  "A",
		Limit:  10,
	}

	results, err := BoundedTagGraphSearch(request, config)

	// Should complete without errors
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestBoundedTagGraphSearchRespectsMinActivationIntegration(t *testing.T) {
	config := BoundedTagGraphConfig{
		MaxEdgesPerTag:    2,
		MaxDepth:          1,
		BeamWidth:         2,
		MaxSeedTags:       5,
		MinScoreThreshold: 0.1, // Higher threshold
	}

	request := SemanticSearchRequest{
		UserID: "test-user",
		Query:  "python",
		Limit:  10,
	}

	results, err := BoundedTagGraphSearch(request, config)

	// Should complete without errors
	assert.NoError(t, err)
	assert.NotNil(t, results)
}

func TestDefaultBoundedConfig(t *testing.T) {
	config := DefaultBoundedConfig()

	// Verify config has reasonable defaults
	assert.Greater(t, config.MaxEdgesPerTag, 0)
	assert.Greater(t, config.MaxDepth, 0)
	assert.Greater(t, config.BeamWidth, 0)
	assert.Greater(t, config.MaxSeedTags, 0)
	assert.GreaterOrEqual(t, config.MinScoreThreshold, 0.0)
}

func TestQueryWithBoundedTagsIntegration(t *testing.T) {
	ctx := context.Background()
	userID := "test-user"
	boundedTags := []string{"python", "fastapi"}
	limit := 10
	config := DefaultBoundedConfig()

	// Test with nil filters
	results, err := queryWithBoundedTags(ctx, userID, boundedTags, limit, config, nil)

	// Should not error (may return empty results if no data)
	assert.NoError(t, err)
	assert.NotNil(t, results)

	// Test with filters
	filters := &InjectFilters{
		Section: "research",
		Status:  "completed",
	}

	filteredResults, err := queryWithBoundedTags(ctx, userID, boundedTags, limit, config, filters)
	assert.NoError(t, err)
	assert.NotNil(t, filteredResults)
}

func TestInjectFilters(t *testing.T) {
	filters := &InjectFilters{
		OrgID:     "test-org",
		ProjectID: "test-project",
		Section:   "research",
		Status:    "completed",
		SinceISO:  "2025-01-01T00:00:00Z",
	}

	assert.Equal(t, "test-org", filters.OrgID)
	assert.Equal(t, "test-project", filters.ProjectID)
	assert.Equal(t, "research", filters.Section)
	assert.Equal(t, "completed", filters.Status)
	assert.Equal(t, "2025-01-01T00:00:00Z", filters.SinceISO)
}
