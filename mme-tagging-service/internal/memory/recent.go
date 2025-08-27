package memory

import (
	"context"
	"time"

	"mme-tagging-service/internal/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetRecentBlocks(userId string, limit int64) ([]MemoryBlock, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"userId": userId}
	opts := options.Find().SetSort(bson.M{"createdAt": -1}).SetLimit(limit)

	cursor, err := db.MemoryCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}

	var results []MemoryBlock
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}
