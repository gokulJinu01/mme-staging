package memory

import (
	"context"
	"strconv"
	"time"

	"mme-tagging-service/internal/db"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BackfillTagsFlat processes existing memories that have tags but no tagsFlat
func BackfillTagsFlat(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	limitStr := c.Query("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Filter: docs for this user that have tags but no tagsFlat (or empty)
	filter := bson.M{
		"userId": userID,
		"tags.0": bson.M{"$exists": true}, // at least one tag
		"$or": []bson.M{
			{"tagsFlat": bson.M{"$exists": false}},
			{"tagsFlat": bson.A{}},
			{"tagsFlat": nil},
		},
	}

	cur, err := db.MemoryCollection.Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "db find failed"})
	}
	defer cur.Close(ctx)

	processed := 0
	for cur.Next(ctx) {
		var m MemoryBlock
		if err := cur.Decode(&m); err != nil {
			continue
		}
		flat := ToFlatTags(m.Tags)
		if len(flat) == 0 {
			continue
		}
		_, _ = db.MemoryCollection.UpdateByID(ctx, m.ID, bson.M{"$set": bson.M{"tagsFlat": flat}})
		processed++
	}

	return c.JSON(fiber.Map{"processed": processed})
}
