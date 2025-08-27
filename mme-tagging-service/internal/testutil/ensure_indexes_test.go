//go:build integration

package testutil

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates all required indexes for the test database
func EnsureIndexes(ctx context.Context, client *mongo.Client) error {
	db := client.Database("mme")

	// Memories collection indexes
	memoriesCollection := db.Collection("memories")
	
	// Index for orgId, projectId, section, status, createdAt
	memoriesIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "orgId", Value: 1},
			{Key: "projectId", Value: 1},
			{Key: "section", Value: 1},
			{Key: "status", Value: 1},
			{Key: "createdAt", Value: -1},
		},
		Options: options.Index().SetName("memories_org_project_section_status_created"),
	}
	
	_, err := memoriesCollection.Indexes().CreateOne(ctx, memoriesIndex)
	if err != nil {
		log.Printf("Failed to create memories index: %v", err)
		return err
	}

	// Tag edges collection indexes
	tagEdgesCollection := db.Collection("tag_edges")
	
	// Unique index for orgId + tag
	tagEdgesUniqueIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "orgId", Value: 1},
			{Key: "tag", Value: 1},
		},
		Options: options.Index().SetName("tag_edges_org_tag_unique").SetUnique(true),
	}
	
	_, err = tagEdgesCollection.Indexes().CreateOne(ctx, tagEdgesUniqueIndex)
	if err != nil {
		log.Printf("Failed to create tag_edges unique index: %v", err)
		return err
	}

	// Optional index for orgId + edges.to
	tagEdgesToIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "orgId", Value: 1},
			{Key: "edges.to", Value: 1},
		},
		Options: options.Index().SetName("tag_edges_org_edges_to"),
	}
	
	_, err = tagEdgesCollection.Indexes().CreateOne(ctx, tagEdgesToIndex)
	if err != nil {
		log.Printf("Failed to create tag_edges edges.to index: %v", err)
		return err
	}

	// Pack events collection indexes
	packEventsCollection := db.Collection("pack_events")
	
	// Index for orgId + timestamp
	packEventsTimeIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "orgId", Value: 1},
			{Key: "ts", Value: -1},
		},
		Options: options.Index().SetName("pack_events_org_ts"),
	}
	
	_, err = packEventsCollection.Indexes().CreateOne(ctx, packEventsTimeIndex)
	if err != nil {
		log.Printf("Failed to create pack_events time index: %v", err)
		return err
	}

	// Index for orgId + packId
	packEventsPackIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "orgId", Value: 1},
			{Key: "packId", Value: 1},
		},
		Options: options.Index().SetName("pack_events_org_pack"),
	}
	
	_, err = packEventsCollection.Indexes().CreateOne(ctx, packEventsPackIndex)
	if err != nil {
		log.Printf("Failed to create pack_events pack index: %v", err)
		return err
	}

	log.Println("✅ All test database indexes created successfully")
	return nil
}
