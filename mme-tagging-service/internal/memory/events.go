package memory

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"mme-tagging-service/internal/db"
	"mme-tagging-service/internal/metrics"
	"go.mongodb.org/mongo-driver/bson"
)

// PackEvent represents a pack acceptance/rejection event
type PackEvent struct {
	OrgID     string   `json:"orgId" bson:"orgId"`
	ProjectID string   `json:"projectId" bson:"projectId"`
	PackID    string   `json:"packId" bson:"packId"`
	Accepted  bool     `json:"accepted" bson:"accepted"`
	Tags      []string `json:"tags" bson:"tags"`
	ItemIDs   []string `json:"itemIds" bson:"itemIds"`
	Timestamp int64    `json:"ts" bson:"ts"`
}

// PackEventRequest represents the request body for pack events
type PackEventRequest struct {
	OrgID     string   `json:"orgId"`
	ProjectID string   `json:"projectId"`
	PackID    string   `json:"packId"`
	Accepted  bool     `json:"accepted"`
	Tags      []string `json:"tags"`
	ItemIDs   []string `json:"itemIds"`
	Timestamp *int64   `json:"ts,omitempty"`
}

// SavePackEvent saves a pack event to the database
func SavePackEvent(event PackEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use current timestamp if not provided
	if event.Timestamp == 0 {
		event.Timestamp = time.Now().Unix()
	}

	_, err := db.PackEventsCollection.InsertOne(ctx, event)
	return err
}

// GetRecentPackEvents retrieves recent pack events for an organization
func GetRecentPackEvents(orgID string, hours int) ([]PackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Calculate the cutoff time
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour).Unix()

	filter := bson.M{
		"orgId": orgID,
		"ts":    bson.M{"$gte": cutoffTime},
	}

	cursor, err := db.PackEventsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []PackEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, err
	}

	return events, nil
}

// GetPackEventByID retrieves a specific pack event by ID
func GetPackEventByID(orgID, packID string) (*PackEvent, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"orgId":  orgID,
		"packId": packID,
	}

	var event PackEvent
	err := db.PackEventsCollection.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		return nil, err
	}

	return &event, nil
}

// ValidatePackEventRequest validates a pack event request
func ValidatePackEventRequest(req PackEventRequest) error {
	if req.OrgID == "" {
		return &ValidationError{Field: "orgId", Message: "orgId is required"}
	}
	if req.PackID == "" {
		return &ValidationError{Field: "packId", Message: "packId is required"}
	}
	if len(req.Tags) == 0 {
		return &ValidationError{Field: "tags", Message: "tags cannot be empty"}
	}
	if len(req.ItemIDs) == 0 {
		return &ValidationError{Field: "itemIds", Message: "itemIds cannot be empty"}
	}
	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// HandlePackEvent handles the POST /events/pack endpoint
func HandlePackEvent(c *fiber.Ctx) error {
	var request PackEventRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := ValidatePackEventRequest(request); err != nil {
		if validationErr, ok := err.(*ValidationError); ok {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": validationErr.Message,
				"field": validationErr.Field,
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get JWT claims for tenancy validation (temporarily disabled)
	// claims, err := middleware.GetJWTClaims(c)
	// if err != nil {
	// 	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
	// 		"error": "JWT claims not found",
	// 	})
	// }

	// Validate orgId matches JWT (temporarily disabled)
	// if request.OrgID != claims.OrgID {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"error": "orgId in request does not match JWT org_id",
	// 	})
	// }

	// Validate projectId if provided (temporarily disabled)
	// if request.ProjectID != "" && claims.ProjectID != "" && request.ProjectID != claims.ProjectID {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"error": "projectId in request does not match JWT project_id",
	// 	})
	// }

	// Create pack event
	event := PackEvent{
		OrgID:     request.OrgID,
		ProjectID: request.ProjectID,
		PackID:    request.PackID,
		Accepted:  request.Accepted,
		Tags:      request.Tags,
		ItemIDs:   request.ItemIDs,
	}

	// Set timestamp if provided, otherwise use current time
	if request.Timestamp != nil {
		event.Timestamp = *request.Timestamp
	} else {
		event.Timestamp = time.Now().Unix()
	}

	// Save event to database
	if err := SavePackEvent(event); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to save pack event",
		})
	}

	// Record metrics
	if request.Accepted {
		metrics.RecordPackAccept(request.OrgID, request.ProjectID)
	} else {
		metrics.RecordPackReject(request.OrgID, request.ProjectID)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"event":  event,
	})
}
