package memory

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"mme-tagging-service/internal/db"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SaveBlock saves a new memory block
func SaveBlock(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Parse request into block
	var requestBody map[string]interface{}
	if err := c.BodyParser(&requestBody); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid payload"})
	}

	// Create memory block
	block := MemoryBlock{
		UserID:    userID,
		CreatedAt: time.Now().UTC(),
	}

	// Parse content
	if content, ok := requestBody["content"].(string); ok {
		block.Content = content
	}

	// Parse section, status, source
	if section, ok := requestBody["section"].(string); ok {
		block.Section = section
	}
	if status, ok := requestBody["status"].(string); ok {
		block.Status = status
	}
	if source, ok := requestBody["source"].(string); ok {
		block.Source = source
	}

	// Handle tags - support both structured []Tag and legacy []string
	if tagsInterface, exists := requestBody["tags"]; exists {
		switch tags := tagsInterface.(type) {
		case []interface{}:
			// Check if it's structured tags (objects) or legacy tags (strings)
			if len(tags) > 0 {
				if _, isObject := tags[0].(map[string]interface{}); isObject {
					// Structured tags - convert to Tag objects
					structuredTags := make([]Tag, 0, len(tags))
					for _, tagInterface := range tags {
						if tagMap, ok := tagInterface.(map[string]interface{}); ok {
							tag := Tag{}
							if label, ok := tagMap["label"].(string); ok {
								tag.Label = NormalizeLabel(label)
							}
							if section, ok := tagMap["section"].(string); ok {
								tag.Section = section
							}
							if origin, ok := tagMap["origin"].(string); ok {
								tag.Origin = origin
							}
							if scope, ok := tagMap["scope"].(string); ok {
								tag.Scope = scope
							}
							if tagType, ok := tagMap["type"].(string); ok {
								tag.Type = tagType
							}
							if confidence, ok := tagMap["confidence"].(float64); ok {
								tag.Confidence = confidence
							}
							if links, ok := tagMap["links"].([]interface{}); ok {
								tag.Links = make([]string, 0, len(links))
								for _, link := range links {
									if linkStr, ok := link.(string); ok {
										tag.Links = append(tag.Links, linkStr)
									}
								}
							}
							if usageCount, ok := tagMap["usageCount"].(float64); ok {
								tag.UsageCount = int(usageCount)
							}
							if lastUsed, ok := tagMap["lastUsed"].(string); ok {
								if parsed, err := time.Parse(time.RFC3339, lastUsed); err == nil {
									tag.LastUsed = parsed
								}
							}
							structuredTags = append(structuredTags, tag)
						}
					}
					block.Tags = structuredTags
				} else {
					// Legacy string tags - convert to structured tags
					stringTags := make([]string, 0, len(tags))
					for _, tagInterface := range tags {
						if tagStr, ok := tagInterface.(string); ok {
							stringTags = append(stringTags, tagStr)
						}
					}
					block.Tags = FromStringsToTags(stringTags)
				}
			}
		}
	}

	// Auto-generate tags if content exists and no tags provided
	if len(block.Tags) == 0 && strings.TrimSpace(block.Content) != "" {
		stringTags, err := extractTagsFromTagmaker(block.Content, userID, c.Get("Authorization"))
		if err == nil && len(stringTags) > 0 {
			block.Tags = FromStringsToTags(stringTags)
		}
	}

	// Ensure tagsFlat computed BEFORE insert
	block.TagsFlat = ToFlatTags(block.Tags)

	// Insert into database
	ctx := context.Background()
	res, err := db.MemoryCollection.InsertOne(ctx, block)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "db insert failed"})
	}

	// Create tag edges (non-blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := UpsertEdges(ctx, block.Tags); err != nil {
			fmt.Printf("Failed to create tag edges: %v\n", err)
		}
	}()

	// Return explicit JSON (never serialize the whole struct)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id":       res.InsertedID,
		"message":  "Memory saved",
		"userId":   block.UserID,
		"tags":     block.Tags,
		"tagsFlat": block.TagsFlat,
		"status":   block.Status,
	})
}

// QueryBlocks queries memory blocks with various filters
func QueryBlocks(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Parse query parameters
	limit := int64(c.QueryInt("limit", 10))
	if limit > 100 {
		limit = 100
	}

	tags := c.Query("tags")
	tagLabel := c.Query("tagLabel")
	tagSection := c.Query("tagSection")
	tagType := c.Query("tagType")
	tagScope := c.Query("tagScope")
	section := c.Query("section")
	status := c.Query("status")

	// Use structured query if any structured parameters are provided
	if tagLabel != "" || tagSection != "" || tagType != "" || tagScope != "" {
		results, err := FindBlocksStructured(userID, tagLabel, tagSection, tagType, tagScope, section, status, limit)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to query memory blocks",
			})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"results": results,
			"count":   len(results),
		})
	}

	// Legacy tag query (backward compatibility)
	tagsList := []string{}
	if tags != "" {
		rawTagsList := strings.Split(tags, ",")
		tagsList = CleanTagsForQuery(rawTagsList) // Apply tag cleaning
	}

	results, err := FindBlocks(userID, tagsList, section, status, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query memory blocks",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"results": results,
		"count":   len(results),
	})
}

// GetRecent gets recent memory blocks
func GetRecent(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	limit := c.QueryInt("limit", 10)

	results, err := GetRecentBlocks(userID, int64(limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch recent blocks",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"userId":  userID,
		"count":   len(results),
	})
}

// DeleteBlock deletes a memory block by ID
func DeleteBlock(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Parse memory ID from path parameter
	idHex := c.Params("id")
	if idHex == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Memory ID is required",
		})
	}

	// Convert to ObjectID
	oid, err := primitive.ObjectIDFromHex(idHex)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid memory ID format",
		})
	}

	// Delete with user scope
	ctx := context.Background()
	result, err := db.MemoryCollection.DeleteOne(ctx, bson.M{
		"_id":    oid,
		"userId": userID,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete memory",
		})
	}

	if result.DeletedCount == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Memory not found or access denied",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Memory deleted successfully",
		"id":      idHex,
	})
}

// PromoteBlocks promotes memory blocks based on tags and goal
func PromoteBlocks(c *fiber.Ctx) error {
	// Get user ID from Traefik ForwardAuth header
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Parse query parameters
	tags := c.Query("tags")
	goal := c.Query("goal")
	mode := c.Query("mode", "normal")

	// Log spike-trace for observability
	fmt.Printf("spike_trace seed=[%s] tier=direct\n", tags)

	// Validate required parameters
	if tags == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Tags parameter is required",
		})
	}

	// Parse tags
	tagsList := strings.Split(tags, ",")
	for i, tag := range tagsList {
		tagsList[i] = strings.TrimSpace(tag)
	}

	// Try direct tag matching first
	allBlocks, err := FindBlocks(userID, tagsList, "", "", 50)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch memory"})
	}

	// Debug: log the number of blocks found
	fmt.Printf("Found %d blocks for tags: %v\n", len(allBlocks), tagsList)

	// Empty-safe fallback: only for standard mode
	if len(allBlocks) == 0 && mode == "standard" {
		// Temporarily disabled neighbor fallback for debugging
		// neighborTags := getNeighborTags(c.Context(), tags)
		// if len(neighborTags) > 0 {
		// 	allBlocks, err = FindBlocks(userID, neighborTags, "", "", 50)
		// 	if err != nil {
		// 		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch neighbor memory"})
		// 	}
		// }
	}

	// Final fallback: get recent memories for user (only in standard mode)
	if len(allBlocks) == 0 && mode == "standard" {
		allBlocks, err = FindBlocks(userID, []string{}, "", "", 10)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch recent memory"})
		}
	}

	final := PromoteMemory(allBlocks, tagsList, goal)

	// determine fallback tier wherever you choose direct/neighbor/recent
	tier, _ := c.Locals("fallbackTier").(string)
	if tier == "" {
		tier = "direct"
	}

	// normalized input tags string (best-effort)
	seed := c.Query("tags")

	// REQUIRED log format:
	logLine := "spike_trace seed=[" + seed + "] tier=" + tier
	fmt.Printf("%s\n", logLine) // Use fmt.Printf for better visibility

	return c.JSON(fiber.Map{
		"results": final,
		"count":   len(final),
	})
}

// getNeighborTags gets one-hop neighbor tags from tag edges
func getNeighborTags(ctx context.Context, seedTags []string) []string {
	neighborSet := make(map[string]bool)

	for _, seedTag := range seedTags {
		// Get related tags using the edge system
		relatedTags, err := GetRelatedTags(ctx, seedTag, 10)
		if err == nil {
			for _, relatedTag := range relatedTags {
				neighborSet[relatedTag] = true
			}
		}
	}

	// Convert set to slice
	neighbors := make([]string, 0, len(neighborSet))
	for neighbor := range neighborSet {
		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}

// ExtractTagsFromPrompt extracts clean tags from a raw prompt
func ExtractTagsFromPrompt(c *fiber.Ctx) error {
	var input PromptInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	// Get authenticated user ID from context
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Override user ID with authenticated user
	input.UserID = userID

	if input.Prompt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Prompt cannot be empty",
		})
	}

	// Limit prompt length for security
	if len(input.Prompt) > 5000 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Prompt too long (max 5000 characters)",
		})
	}

	tags := CleanPromptToTags(input.Prompt)

	response := TagExtractionResponse{
		UserID: userID,
		Tags:   tags,
		Count:  len(tags),
	}

	return c.Status(http.StatusOK).JSON(response)
}

// QueryBlocksByPrompt queries memory blocks using a natural language prompt
func QueryBlocksByPrompt(c *fiber.Ctx) error {
	var input PromptQueryInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	// Get authenticated user ID from context
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Override user ID with authenticated user
	input.UserID = userID

	if input.Prompt == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Prompt cannot be empty",
		})
	}

	// Limit prompt length for security
	if len(input.Prompt) > 5000 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Prompt too long (max 5000 characters)",
		})
	}

	// Extract tags from prompt
	tags := CleanPromptToTags(input.Prompt)

	// Set default limit if not specified
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100 // Max limit for performance
	}

	// Query memory blocks using extracted tags
	results, err := FindBlocks(input.UserID, tags, input.Section, input.Status, int64(limit))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query memory blocks",
		})
	}

	// Return results with metadata
	response := fiber.Map{
		"prompt":        input.Prompt,
		"extractedTags": tags,
		"tagCount":      len(tags),
		"results":       results,
		"resultCount":   len(results),
		"userId":        userID,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// QueryBlocksDebug is a debug version of QueryBlocks that returns debug information
func QueryBlocksDebug(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID := c.Get("X-User-ID")
	if userID == "" {
		userID = "test-user" // Debug mode fallback
	}

	rawTags := c.Query("tags", "")
	tags := []string{}
	if rawTags != "" {
		rawTagsList := strings.Split(rawTags, ",")
		tags = CleanTagsForQuery(rawTagsList)
	}

	section := c.Query("section", "")
	status := c.Query("status", "")
	limit := c.QueryInt("limit", 10)

	// Get filter as would be used by MongoDB
	filter := map[string]interface{}{
		"userId": userID,
	}
	if len(tags) > 0 {
		filter["tags"] = map[string]interface{}{"$in": tags}
	}
	if section != "" {
		filter["section"] = section
	}
	if status != "" {
		filter["status"] = status
	}

	results, err := FindBlocks(userID, tags, section, status, int64(limit))

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"debug": fiber.Map{
			"userID":       userID,
			"rawTags":      rawTags,
			"rawTagsArray": strings.Split(rawTags, ","),
			"cleanedTags":  tags,
			"section":      section,
			"status":       status,
			"limit":        limit,
			"mongoFilter":  filter,
		},
		"results": results,
		"count":   len(results),
		"error":   err,
	})
}

// UpdateTagDelta handles delta operations from tagmaker service
func UpdateTagDelta(c *fiber.Ctx) error {
	// Parse the delta operation
	var delta map[string]interface{}
	if err := c.BodyParser(&delta); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid delta format",
		})
	}

	// Try to get user ID from authentication first
	userID := c.Get("X-User-ID")
	if userID == "" {
		// Fallback to header for service-to-service communication
		userID = c.Get("X-USER-ID")
		if userID == "" {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "User ID required (from auth or X-USER-ID header)",
			})
		}
	}

	// Process delta operation
	tag, ok := delta["tag"].(string)
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Delta must contain 'tag' field",
		})
	}

	ops, ok := delta["ops"].(map[string]interface{})
	if !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Delta must contain 'ops' field",
		})
	}

	// Apply the delta operation to the memory collection
	err := ApplyTagDelta(userID, tag, ops)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to apply delta",
			"details": err.Error(),
		})
	}

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Delta applied successfully",
		"tag":     tag,
		"userID":  userID,
	})
}

// ApplyTagDelta applies a delta operation to the memory collection
func ApplyTagDelta(userID, tag string, ops map[string]interface{}) error {
	ctx := context.Background()

	// Build MongoDB update document from delta operations
	updateDoc := bson.M{}

	// Handle $inc operations
	if incOps, ok := ops["$inc"].(map[string]interface{}); ok {
		updateDoc["$inc"] = incOps
	}

	// Handle $set operations
	if setOps, ok := ops["$set"].(map[string]interface{}); ok {
		updateDoc["$set"] = setOps
	}

	// Handle $addToSet operations
	if addToSetOps, ok := ops["$addToSet"].(map[string]interface{}); ok {
		updateDoc["$addToSet"] = addToSetOps
	}

	// Find or create a memory block for this tag
	filter := bson.M{
		"userId":  userID,
		"tags":    bson.M{"$in": []string{tag}},
		"section": "tagmaker",
	}

	// If no document matches, create one
	updateDoc["$setOnInsert"] = bson.M{
		"userId":    userID,
		"tags":      []string{tag},
		"section":   "tagmaker",
		"status":    "active",
		"content":   "Auto-generated tag data",
		"source":    "mme-tagmaker",
		"createdAt": time.Now().UTC(),
	}

	// Apply the update with upsert
	opts := options.Update().SetUpsert(true)

	_, err := db.MemoryCollection.UpdateOne(ctx, filter, updateDoc, opts)
	return err
}

// TagRequest represents the request structure for tagmaker service
type TagRequest struct {
	Content string `json:"content"`
	UserID  string `json:"userId"`
}

// TagResponse represents the response from tagmaker service
type TagResponse struct {
	Saved      bool     `json:"saved"`
	Cues       []string `json:"cues"`
	Confidence float64  `json:"confidence"`
	PrimaryTag string   `json:"primary_tag,omitempty"`
}

// extractTagsFromTagmaker calls the tagmaker service to extract tags from content
func extractTagsFromTagmaker(content, userID, authHeader string) ([]string, error) {
	// Get tagmaker service URL from environment
	tagmakerURL := os.Getenv("MME_TAGMAKER_SERVICE_URL")
	if tagmakerURL == "" {
		tagmakerURL = "http://mme-tagmaker-service:8000"
	}

	// Prepare request payload
	payload := TagRequest{
		Content: content,
		UserID:  userID,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", tagmakerURL+"/extract-tags", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Set timeout for the request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tagmaker service request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tagmaker service returned status %d", resp.StatusCode)
	}

	// Parse response
	var tagResponse TagResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract tags from cues
	var tags []string
	if tagResponse.PrimaryTag != "" {
		tags = append(tags, tagResponse.PrimaryTag)
	}

	// Add other meaningful tags from cues
	for _, cue := range tagResponse.Cues {
		// Extract the concept part before the colon
		if parts := strings.Split(cue, ":"); len(parts) > 0 {
			concept := strings.TrimSpace(parts[0])
			if concept != "" && concept != tagResponse.PrimaryTag {
				tags = append(tags, concept)
			}
		}
	}

	// Limit to first 10 tags to prevent bloat
	if len(tags) > 10 {
		tags = tags[:10]
	}

	return tags, nil
}

// BackfillTags processes existing memories without tagsFlat
func BackfillTags(c *fiber.Ctx) error {
	// Get authenticated user ID from context
	userID := c.Get("X-User-ID")
	if userID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User authentication required",
		})
	}

	// Parse limit
	limitStr := c.Query("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// Find memories that have tags but no tagsFlat (or empty)
	ctx := context.Background()
	filter := bson.M{
		"userId": userID,
		"$and": []bson.M{
			{"tags.0": bson.M{"$exists": true}}, // at least one tag
			{"$or": []bson.M{
				{"tagsFlat": bson.M{"$exists": false}},
				{"tagsFlat": bson.A{}},
				{"tagsFlat": nil},
			}},
		},
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := db.MemoryCollection.Find(ctx, filter, opts)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to query memories",
		})
	}
	defer cursor.Close(ctx)

	processed := 0
	for cursor.Next(ctx) {
		var block MemoryBlock
		if err := cursor.Decode(&block); err != nil {
			continue
		}

		// Generate tagsFlat from existing tags
		flat := ToFlatTags(block.Tags)
		if len(flat) == 0 {
			continue // Skip if no valid tags
		}

		// Update the document
		_, err := db.MemoryCollection.UpdateByID(ctx, block.ID, bson.M{
			"$set": bson.M{"tagsFlat": flat},
		})
		if err == nil {
			processed++
		}
	}

	return c.JSON(fiber.Map{
		"processed": processed,
		"userId":    userID,
		"limit":     limit,
	})
}

// DeleteBlockByIDAndUser deletes a memory block by ID and user ID
func DeleteBlockByIDAndUser(id, userID string) (*mongo.DeleteResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// propagate a typed error so handler can return 400
		return nil, fmt.Errorf("invalid_object_id: %w", err)
	}

	filter := bson.M{
		"_id":    objectID, // ✅ use ObjectID, not string
		"userId": userID,
	}
	return db.MemoryCollection.DeleteOne(ctx, filter)
}

// SemanticSearch performs semantic search on memory blocks
func SemanticSearch(c *fiber.Ctx) error {
	var request SemanticSearchRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JSON format",
		})
	}

	// AUTHENTICATION DISABLED - Use default user for testing
	userID := c.Get("X-User-ID")
	if userID == "" {
		userID = "test-user"
	}

	// Override user ID with authenticated user
	request.UserID = userID

	if request.Query == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Query cannot be empty",
		})
	}

	// Limit query length for security
	if len(request.Query) > 5000 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Query too long (max 5000 characters)",
		})
	}

	// Set default limit if not specified
	if request.Limit <= 0 {
		request.Limit = 10
	}
	if request.Limit > 100 {
		request.Limit = 100 // Max limit for performance
	}

	// Set default minimum score if not specified
	if request.MinScore <= 0 {
		request.MinScore = 0.1
	}

	// Perform semantic search
	results, err := PerformSemanticSearch(request)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to perform semantic search",
			"details": err.Error(),
		})
	}

	// Return results
	response := SemanticSearchResponse{
		Query:   request.Query,
		Results: results,
		Count:   len(results),
		UserID:  userID,
	}

	return c.Status(http.StatusOK).JSON(response)
}
