package memory

import (
	"context"

	"mme-tagging-service/internal/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindBlocks finds memory blocks with various filters
func FindBlocks(userId string, tags []string, section string, status string, limit int64) ([]MemoryBlock, error) {
	ctx := context.Background()
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{"userId": userId}

	// Add section filter if provided
	if section != "" {
		filter["section"] = section
	}

	// Add status filter if provided
	if status != "" {
		filter["status"] = status
	}

	// Add tag filter if tags provided
	if len(tags) > 0 {
		// Normalize tags for consistent matching
		normalizedTags := make([]string, len(tags))
		for i, tag := range tags {
			normalizedTags[i] = NormalizeLabel(tag)
		}
		// Use tagsFlat for backward compatibility
		filter["tagsFlat"] = bson.M{"$in": normalizedTags}
	}

	cursor, err := db.MemoryCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blocks []MemoryBlock
	if err = cursor.All(ctx, &blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}

// FindBlocksStructured finds memory blocks using structured tag fields
func FindBlocksStructured(userId string, tagLabel, tagSection, tagType, tagScope, section, status string, limit int64) ([]MemoryBlock, error) {
	ctx := context.Background()
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	filter := bson.M{"userId": userId}

	// Add section filter if provided
	if section != "" {
		filter["section"] = section
	}

	// Add status filter if provided
	if status != "" {
		filter["status"] = status
	}

	// Build structured tag filters
	tagFilters := bson.M{}
	if tagLabel != "" {
		tagFilters["label"] = NormalizeLabel(tagLabel)
	}
	if tagSection != "" {
		tagFilters["section"] = tagSection
	}
	if tagType != "" {
		tagFilters["type"] = tagType
	}
	if tagScope != "" {
		tagFilters["scope"] = tagScope
	}

	// Add structured tag filter if any tag fields provided
	if len(tagFilters) > 0 {
		filter["tags"] = bson.M{"$elemMatch": tagFilters}
	}

	cursor, err := db.MemoryCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blocks []MemoryBlock
	if err = cursor.All(ctx, &blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}

// FindBlocksWithFilter finds memory blocks with additional quality filters
func FindBlocksWithFilter(userId string, tags []string, section string, status string, limit int64, qualityFilter bson.M) ([]MemoryBlock, error) {
	ctx := context.Background()
	opts := options.Find().SetLimit(limit).SetSort(bson.D{{Key: "createdAt", Value: -1}})

	// Start with quality filter
	filter := qualityFilter

	// Add section filter if provided
	if section != "" {
		filter["section"] = section
	}

	// Add status filter if provided
	if status != "" {
		filter["status"] = status
	}

	// Add tag filter if tags provided
	if len(tags) > 0 {
		// For tag queries, use tagsFlat for backward compatibility
		filter["tagsFlat"] = bson.M{"$in": tags}
	}

	// Remove $expr if it's causing issues and no tags are provided
	if len(tags) == 0 && filter["$expr"] != nil {
		delete(filter, "$expr")
	}

	cursor, err := db.MemoryCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var blocks []MemoryBlock
	if err = cursor.All(ctx, &blocks); err != nil {
		return nil, err
	}

	return blocks, nil
}
