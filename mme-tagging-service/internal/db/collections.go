package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection references
var (
	MemoryCollection      *mongo.Collection
	TagsCollection        *mongo.Collection
	CleanupLogsCollection *mongo.Collection
	UserStatsCollection   *mongo.Collection
	TagEdgesCollection    *mongo.Collection
	PackEventsCollection  *mongo.Collection
)

// InitCollections initializes all MongoDB collections and creates indexes
func InitCollections() {
	if MongoClient == nil {
		log.Fatal("MongoDB client not initialized")
	}

	db := MongoClient.Database("mme")

	// Initialize collections
	MemoryCollection = db.Collection("memories")
	TagsCollection = db.Collection("tags")
	CleanupLogsCollection = db.Collection("cleanup_logs")
	UserStatsCollection = db.Collection("user_stats")
	TagEdgesCollection = db.Collection("tag_edges")
	PackEventsCollection = db.Collection("pack_events")

	// Create indexes
	createIndexes()
}

// createIndexes creates all required indexes for optimal performance
func createIndexes() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Memory collection indexes
	memoryIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("userId_createdAt_desc"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tagsFlat", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("userId_tagsFlat_createdAt"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tags.label", Value: 1},
			},
			Options: options.Index().SetName("userId_tags_label"),
		},
		{
			Keys: bson.D{
				{Key: "tags.section", Value: 1},
			},
			Options: options.Index().SetName("tags_section"),
		},
		{
			Keys: bson.D{
				{Key: "ttl", Value: 1},
			},
			Options: options.Index().SetName("ttl_index"),
		},
		{
			Keys: bson.D{
				{Key: "orgId", Value: 1},
				{Key: "projectId", Value: 1},
				{Key: "section", Value: 1},
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
			Options: options.Index().SetName("org_project_section_status_created"),
		},
	}

	// Tags collection indexes
	tagsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "tag", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("userId_tag_unique"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "hotness", Value: -1},
			},
			Options: options.Index().SetName("userId_hotness_desc"),
		},
		{
			Keys: bson.D{
				{Key: "lastUsed", Value: 1},
			},
			Options: options.Index().SetName("lastUsed_index"),
		},
	}

	// Cleanup logs collection indexes
	cleanupIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("timestamp_desc"),
		},
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
				{Key: "timestamp", Value: -1},
			},
			Options: options.Index().SetName("userId_timestamp_desc"),
		},
	}

	// User stats collection indexes
	userStatsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "userId", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("userId_unique"),
		},
	}

	// Tag edges collection indexes
	tagEdgesIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "a", Value: 1},
				{Key: "b", Value: 1},
			},
			Options: options.Index().SetUnique(true).SetName("a_b_unique"),
		},
		{
			Keys: bson.D{
				{Key: "weight", Value: -1},
				{Key: "hits", Value: -1},
			},
			Options: options.Index().SetName("weight_hits_desc"),
		},
	}

	// Pack events collection indexes
	packEventsIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "orgId", Value: 1},
				{Key: "ts", Value: -1},
			},
			Options: options.Index().SetName("orgId_ts_desc"),
		},
		{
			Keys: bson.D{
				{Key: "orgId", Value: 1},
				{Key: "packId", Value: 1},
			},
			Options: options.Index().SetName("orgId_packId"),
		},
	}

	// Create indexes for each collection
	if _, err := MemoryCollection.Indexes().CreateMany(ctx, memoryIndexes); err != nil {
		log.Printf("Warning: Failed to create memory indexes: %v", err)
	}

	if _, err := TagsCollection.Indexes().CreateMany(ctx, tagsIndexes); err != nil {
		log.Printf("Warning: Failed to create tags indexes: %v", err)
	}

	if _, err := CleanupLogsCollection.Indexes().CreateMany(ctx, cleanupIndexes); err != nil {
		log.Printf("Warning: Failed to create cleanup logs indexes: %v", err)
	}

	if _, err := UserStatsCollection.Indexes().CreateMany(ctx, userStatsIndexes); err != nil {
		log.Printf("Warning: Failed to create user stats indexes: %v", err)
	}

	if _, err := TagEdgesCollection.Indexes().CreateMany(ctx, tagEdgesIndexes); err != nil {

		log.Printf("Warning: Failed to create tag edges indexes: %v", err)
	}

	if _, err := PackEventsCollection.Indexes().CreateMany(ctx, packEventsIndexes); err != nil {
		log.Printf("Warning: Failed to create pack events indexes: %v", err)
	}

	log.Println("MongoDB collections and indexes initialized successfully")
}
