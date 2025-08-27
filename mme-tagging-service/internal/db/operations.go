package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation  string
	Collection string
	Err        error
	UserID     string
}

func (e *DatabaseError) Error() string {
	return e.Err.Error()
}

// logDatabaseOperation logs database operations for monitoring
func logDatabaseOperation(operation, collection, userID string, err error) {
	if err != nil {
		log.Printf("Database operation failed: %s on %s for user %s: %v",
			operation, collection, userID, err)
	} else {
		log.Printf("Database operation successful: %s on %s for user %s",
			operation, collection, userID)
	}
}

// GetTotalMemoryBlocks returns the total number of memory blocks for a user
func GetTotalMemoryBlocks(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	count, err := MemoryCollection.CountDocuments(ctx, filter)
	if err != nil {
		logDatabaseOperation("count_memory_blocks", "memories", userID, err)
		return 0, &DatabaseError{
			Operation:  "count_memory_blocks",
			Collection: "memories",
			Err:        err,
			UserID:     userID,
		}
	}

	logDatabaseOperation("count_memory_blocks", "memories", userID, nil)
	return int(count), nil
}

// GetTotalTags returns the total number of unique tags for a user
func GetTotalTags(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use aggregation to count unique tags
	pipeline := []bson.M{
		{"$match": bson.M{"userId": userID}},
		{"$unwind": "$tags"},
		{"$group": bson.M{"_id": "$tags"}},
		{"$count": "unique_tags"},
	}

	cursor, err := MemoryCollection.Aggregate(ctx, pipeline)
	if err != nil {
		logDatabaseOperation("count_unique_tags", "memories", userID, err)
		return 0, &DatabaseError{
			Operation:  "count_unique_tags",
			Collection: "memories",
			Err:        err,
			UserID:     userID,
		}
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		logDatabaseOperation("count_unique_tags", "memories", userID, err)
		return 0, &DatabaseError{
			Operation:  "count_unique_tags",
			Collection: "memories",
			Err:        err,
			UserID:     userID,
		}
	}

	count := 0
	if len(results) > 0 {
		if uniqueTags, ok := results[0]["unique_tags"].(int32); ok {
			count = int(uniqueTags)
		}
	}

	logDatabaseOperation("count_unique_tags", "memories", userID, nil)
	return count, nil
}

// GetLastCleanupTime returns the timestamp of the last cleanup operation
func GetLastCleanupTime(userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})

	var cleanupLog bson.M
	err := CleanupLogsCollection.FindOne(ctx, filter, opts).Decode(&cleanupLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No cleanup logs found, return empty string
			logDatabaseOperation("get_last_cleanup", "cleanup_logs", userID, nil)
			return "", nil
		}
		logDatabaseOperation("get_last_cleanup", "cleanup_logs", userID, err)
		return "", &DatabaseError{
			Operation:  "get_last_cleanup",
			Collection: "cleanup_logs",
			Err:        err,
			UserID:     userID,
		}
	}

	if timestamp, ok := cleanupLog["timestamp"].(time.Time); ok {
		logDatabaseOperation("get_last_cleanup", "cleanup_logs", userID, nil)
		return timestamp.Format(time.RFC3339), nil
	}

	logDatabaseOperation("get_last_cleanup", "cleanup_logs", userID, nil)
	return "", nil
}

// PerformMemoryCleanup removes expired or old memory blocks
func PerformMemoryCleanup(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define cleanup criteria
	now := time.Now().UTC()
	cleanupStart := time.Now()

	// Clean up expired TTL records
	ttlFilter := bson.M{
		"userId": userID,
		"ttl": bson.M{
			"$lt": now.Unix(),
		},
	}

	// Clean up old records (older than 90 days)
	oldRecordsFilter := bson.M{
		"userId": userID,
		"createdAt": bson.M{
			"$lt": now.AddDate(0, 0, -90),
		},
	}

	// Combine filters
	filter := bson.M{
		"$or": []bson.M{ttlFilter, oldRecordsFilter},
	}

	// Delete expired/old records
	result, err := MemoryCollection.DeleteMany(ctx, filter)
	if err != nil {
		logDatabaseOperation("memory_cleanup", "memories", userID, err)
		return 0, &DatabaseError{
			Operation:  "memory_cleanup",
			Collection: "memories",
			Err:        err,
			UserID:     userID,
		}
	}

	deletedCount := int(result.DeletedCount)
	duration := time.Since(cleanupStart)

	// Log cleanup operation
	cleanupLog := bson.M{
		"operation":    "memory_cleanup",
		"userId":       userID,
		"itemsCleaned": deletedCount,
		"durationMs":   duration.Milliseconds(),
		"timestamp":    now,
		"status":       "success",
	}

	_, logErr := CleanupLogsCollection.InsertOne(ctx, cleanupLog)
	if logErr != nil {
		log.Printf("Warning: Failed to log cleanup operation: %v", logErr)
	}

	logDatabaseOperation("memory_cleanup", "memories", userID, nil)
	return deletedCount, nil
}

// PerformTagCleanup removes unused or low-value tags
func PerformTagCleanup(userID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define cleanup criteria
	now := time.Now().UTC()
	cleanupStart := time.Now()

	// Clean up tags that haven't been used in 30 days and have low hotness
	filter := bson.M{
		"userId": userID,
		"$or": []bson.M{
			{
				"lastUsed": bson.M{
					"$lt": now.AddDate(0, 0, -30),
				},
				"hotness": bson.M{
					"$lt": 0.1,
				},
			},
			{
				"lastUsed": bson.M{
					"$exists": false,
				},
				"hotness": bson.M{
					"$lt": 0.05,
				},
			},
		},
	}

	// Delete unused/low-value tags
	result, err := TagsCollection.DeleteMany(ctx, filter)
	if err != nil {
		logDatabaseOperation("tag_cleanup", "tags", userID, err)
		return 0, &DatabaseError{
			Operation:  "tag_cleanup",
			Collection: "tags",
			Err:        err,
			UserID:     userID,
		}
	}

	deletedCount := int(result.DeletedCount)
	duration := time.Since(cleanupStart)

	// Log cleanup operation
	cleanupLog := bson.M{
		"operation":    "tag_cleanup",
		"userId":       userID,
		"itemsCleaned": deletedCount,
		"durationMs":   duration.Milliseconds(),
		"timestamp":    now,
		"status":       "success",
	}

	_, logErr := CleanupLogsCollection.InsertOne(ctx, cleanupLog)
	if logErr != nil {
		log.Printf("Warning: Failed to log cleanup operation: %v", logErr)
	}

	logDatabaseOperation("tag_cleanup", "tags", userID, nil)
	return deletedCount, nil
}

// GetCleanupDuration returns the duration of the last cleanup operation
func GetCleanupDuration(userID string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userID}
	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})

	var cleanupLog bson.M
	err := CleanupLogsCollection.FindOne(ctx, filter, opts).Decode(&cleanupLog)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// No cleanup logs found, return 0
			return 0, nil
		}
		return 0, &DatabaseError{
			Operation:  "get_cleanup_duration",
			Collection: "cleanup_logs",
			Err:        err,
			UserID:     userID,
		}
	}

	if durationMs, ok := cleanupLog["durationMs"].(int64); ok {
		return durationMs, nil
	}

	return 0, nil
}
