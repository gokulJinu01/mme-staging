//go:build integration

package testutil

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"mme-tagging-service/internal/db"
	"mme-tagging-service/internal/memory"
)

var (
	mongoContainer testcontainers.Container
	mongoURI       string
)

// TestMain sets up and tears down the MongoDB container for all tests
func TestMain(m *testing.M) {
	ctx := context.Background()

	// Start MongoDB container
	req := testcontainers.ContainerRequest{
		Image:        "mongo:7.0",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections"),
		Env: map[string]string{
			"MONGO_INITDB_ROOT_USERNAME": "test",
			"MONGO_INITDB_ROOT_PASSWORD": "test",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatalf("Failed to start MongoDB container: %v", err)
	}
	mongoContainer = container

	// Get container host and port
	host, err := container.Host(ctx)
	if err != nil {
		log.Fatalf("Failed to get container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "27017")
	if err != nil {
		log.Fatalf("Failed to get container port: %v", err)
	}

	// Set up MongoDB URI
	mongoURI = fmt.Sprintf("mongodb://test:test@%s:%s", host, port.Port())

	// Wait for MongoDB to be ready
	time.Sleep(5 * time.Second)

	// Initialize database connection and indexes
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Ensure all required indexes are created
	err = EnsureIndexes(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Close the connection
	client.Disconnect(ctx)

	// Run tests
	code := m.Run()

	// Clean up
	if err := container.Terminate(ctx); err != nil {
		log.Printf("Failed to terminate MongoDB container: %v", err)
	}

	os.Exit(code)
}

// GetMongoURI returns the MongoDB URI for the test container
func GetMongoURI() string {
	return mongoURI
}

// SetupTestDB initializes the database connection for tests
func SetupTestDB(t *testing.T) *mongo.Client {
	t.Helper()

	// Set environment variable for MongoDB URI
	t.Setenv("MONGODB_URI", mongoURI)

	// Initialize database connection
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	require.NoError(t, err)

	// Test the connection
	err = client.Ping(context.Background(), nil)
	require.NoError(t, err)

	// Initialize global database client and collections
	db.MongoClient = client
	db.InitCollections()
	
	// Ensure indexes are created
	err = EnsureIndexes(context.Background(), client)
	require.NoError(t, err)

	return client
}

// CleanupTestDB cleans up test data
func CleanupTestDB(t *testing.T, client *mongo.Client) {
	t.Helper()

	// Drop all collections
	collections := []string{"memories", "tags", "cleanup_logs", "user_stats", "tag_edges", "pack_events"}
	for _, collectionName := range collections {
		collection := client.Database("mme").Collection(collectionName)
		err := collection.Drop(context.Background())
		if err != nil {
			t.Logf("Warning: Failed to drop collection %s: %v", collectionName, err)
		}
	}
}

// SeedTagEdges seeds tag edges for testing
func SeedTagEdges(t *testing.T, orgID string, edges map[string][]memory.Edge) {
	t.Helper()

	client := SetupTestDB(t)
	defer CleanupTestDB(t, client)

	collection := client.Database("mme").Collection("tag_edges")

	for tag, tagEdges := range edges {
		// Convert edges to documents
		var edgeDocs []bson.M
		for _, edge := range tagEdges {
			edgeDocs = append(edgeDocs, bson.M{
				"to":        edge.To,
				"w":         edge.W,
				"facetSrc":  edge.FacetSrc,
				"facetDst":  edge.FacetDst,
				"ts":        edge.Ts,
			})
		}

		doc := bson.M{
			"orgId": orgID,
			"tag":   tag,
			"edges": edgeDocs,
		}

		_, err := collection.InsertOne(context.Background(), doc)
		require.NoError(t, err)
	}
}

// SeedMemories seeds memory blocks for testing
func SeedMemories(t *testing.T, orgID string, blocks []memory.MemoryBlock) {
	t.Helper()

	client := SetupTestDB(t)
	defer CleanupTestDB(t, client)

	collection := client.Database("mme").Collection("memories")

	for _, block := range blocks {
		// Convert memory block to document
		doc := bson.M{
			"userId":    block.UserID,
			"tags":      block.Tags,
			"section":   block.Section,
			"status":    block.Status,
			"content":   block.Content,
			"source":    block.Source,
			"createdAt": block.CreatedAt,
			"hash":      block.Hash,
			"priority":  block.Priority,
		}

		if block.Confidence > 0 {
			doc["confidence"] = block.Confidence
		}

		if block.TTL > 0 {
			doc["ttl"] = block.TTL
		}

		_, err := collection.InsertOne(context.Background(), doc)
		require.NoError(t, err)
	}
}

// SeedPackEvents seeds pack events for testing
func SeedPackEvents(t *testing.T, orgID string, events []memory.PackEvent) {
	t.Helper()

	client := SetupTestDB(t)
	defer CleanupTestDB(t, client)

	collection := client.Database("mme").Collection("pack_events")

	for _, event := range events {
		doc := bson.M{
			"orgId":     event.OrgID,
			"projectId": event.ProjectID,
			"packId":    event.PackID,
			"accepted":  event.Accepted,
			"tags":      event.Tags,
			"itemIds":   event.ItemIDs,
			"ts":        event.Timestamp,
		}

		_, err := collection.InsertOne(context.Background(), doc)
		require.NoError(t, err)
	}
}

// GetTagEdges retrieves tag edges for verification
func GetTagEdges(t *testing.T, orgID, tag string) []memory.Edge {
	t.Helper()

	client := SetupTestDB(t)
	defer CleanupTestDB(t, client)

	collection := client.Database("mme").Collection("tag_edges")

	var doc bson.M
	err := collection.FindOne(context.Background(), bson.M{"orgId": orgID, "tag": tag}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		require.NoError(t, err)
	}

	edgesRaw := doc["edges"].(bson.A)
	var edges []memory.Edge
	for _, edgeRaw := range edgesRaw {
		edgeDoc := edgeRaw.(bson.M)
		edge := memory.Edge{
			To:       edgeDoc["to"].(string),
			W:        edgeDoc["w"].(float64),
			FacetSrc: edgeDoc["facetSrc"].(string),
			FacetDst: edgeDoc["facetDst"].(string),
			Ts:       edgeDoc["ts"].(int64),
		}
		edges = append(edges, edge)
	}

	return edges
}

// GetMemories retrieves memory blocks for verification
func GetMemories(t *testing.T, orgID string) []memory.MemoryBlock {
	t.Helper()

	client := SetupTestDB(t)
	defer CleanupTestDB(t, client)

	collection := client.Database("mme").Collection("memories")

	cursor, err := collection.Find(context.Background(), bson.M{})
	require.NoError(t, err)
	defer cursor.Close(context.Background())

	var blocks []memory.MemoryBlock
	err = cursor.All(context.Background(), &blocks)
	require.NoError(t, err)

	return blocks
}
