package memory

import (
	"context"
	"math"
	"time"

	"mme-tagging-service/internal/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// TagEdge represents a directed edge between tags
type TagEdge struct {
	A         string    `bson:"a" json:"a"`                 // normalized label
	B         string    `bson:"b" json:"b"`                 // normalized label
	Weight    float64   `bson:"weight" json:"weight"`       // learned strength
	Hits      int       `bson:"hits" json:"hits"`           // usage count
	LastUsed  time.Time `bson:"lastUsed" json:"lastUsed"`   // last usage timestamp
	Scope     string    `bson:"scope" json:"scope"`         // most permissive of the pair
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"` // creation timestamp
}

// UpsertEdges creates/updates edges for all tag pairs within a memory block
func UpsertEdges(ctx context.Context, tags []Tag) error {
	if len(tags) < 2 {
		return nil // Need at least 2 tags to create edges
	}

	// Create all unordered pairs (a,b) where a != b
	for i := 0; i < len(tags); i++ {
		for j := i + 1; j < len(tags); j++ {
			// Normalize labels for consistent edge creation
			a := NormalizeLabel(tags[i].Label)
			b := NormalizeLabel(tags[j].Label)
			
			if a == b {
				continue // Skip self-edges
			}

			// Ensure consistent ordering (a < b)
			if a > b {
				a, b = b, a
			}

			err := upsertSingleEdge(ctx, TagEdge{
				A:         a,
				B:         b,
				Weight:    0.1, // Initial weight
				Hits:      0,
				LastUsed:  time.Now(),
				Scope:     getMostPermissiveScope(tags[i].Scope, tags[j].Scope),
				CreatedAt: time.Now(),
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// upsertSingleEdge performs atomic upsert for a single edge
func upsertSingleEdge(ctx context.Context, edge TagEdge) error {
	filter := bson.M{"a": edge.A, "b": edge.B}
	update := bson.M{
		"$inc": bson.M{"hits": 1},
		"$max": bson.M{"lastUsed": edge.LastUsed},
		"$setOnInsert": bson.M{
			"weight":    edge.Weight,
			"scope":     edge.Scope,
			"createdAt": edge.CreatedAt,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := db.TagEdgesCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

// getMostPermissiveScope returns the most permissive scope from two tags
func getMostPermissiveScope(scope1, scope2 string) string {
	// Order: global > shared > local
	if scope1 == "global" || scope2 == "global" {
		return "global"
	}
	if scope1 == "shared" || scope2 == "shared" {
		return "shared"
	}
	return "local"
}

// updateEdgeWeight updates the weight of an edge based on hits and recency
func updateEdgeWeight(ctx context.Context, a, b string) error {
	// Get current edge to calculate new weight
	var edge TagEdge
	err := db.TagEdgesCollection.FindOne(ctx, bson.M{"a": a, "b": b}).Decode(&edge)
	if err != nil {
		return err
	}

	// Calculate new weight: bounded function of hits & recency
	// weight = min(1.0, 0.1*log(hits+1)) * recencyBoost
	hits := float64(edge.Hits)
	recencyBoost := calculateRecencyBoost(edge.LastUsed)

	newWeight := math.Min(1.0, 0.1*math.Log(hits+1)) * recencyBoost

	// Update weight
	_, err = db.TagEdgesCollection.UpdateOne(
		ctx,
		bson.M{"a": a, "b": b},
		bson.M{"$set": bson.M{"weight": newWeight}},
	)

	return err
}

// calculateRecencyBoost calculates a recency boost factor (0.5 to 1.0)
func calculateRecencyBoost(lastUsed time.Time) float64 {
	now := time.Now()
	hoursSince := now.Sub(lastUsed).Hours()

	// Decay over 30 days (720 hours)
	if hoursSince > 720 {
		return 0.5
	}

	// Linear decay from 1.0 to 0.5 over 30 days
	decay := hoursSince / 720.0
	return 1.0 - (0.5 * decay)
}

// GetRelatedTags gets related tags for a given tag using edge weights
func GetRelatedTags(ctx context.Context, tag string, limit int) ([]string, error) {
	normalizedTag := NormalizeLabel(tag)
	if normalizedTag == "" {
		return []string{}, nil
	}

	// Find edges where this tag is either A or B
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"$or": []bson.M{
					{"a": normalizedTag},
					{"b": normalizedTag},
				},
			},
		},
		{
			"$sort": bson.M{"weight": -1, "hits": -1},
		},
		{
			"$limit": int64(limit),
		},
	}

	cursor, err := db.TagEdgesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return []string{}, err
	}
	defer cursor.Close(ctx)

	var relatedTags []string
	for cursor.Next(ctx) {
		var edge TagEdge
		if err := cursor.Decode(&edge); err != nil {
			continue
		}

		// Get the other tag (not the input tag)
		if edge.A == normalizedTag {
			relatedTags = append(relatedTags, edge.B)
		} else {
			relatedTags = append(relatedTags, edge.A)
		}
	}

	return relatedTags, nil
}
