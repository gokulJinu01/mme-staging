package api

import (
	"runtime"
	"sync/atomic"
	"time"

	"github.com/gofiber/fiber/v2"
	"mme-tagging-service/internal/db"
	"mme-tagging-service/internal/feature"
	"mme-tagging-service/internal/http/middleware"
	"mme-tagging-service/internal/memory"
	"mme-tagging-service/internal/metrics"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App) {
	// Health check endpoint (no auth required)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Metrics endpoint
	app.Get("/metrics", func(c *fiber.Ctx) error {
		return c.JSON(metrics.GetMetrics())
	})

	// Memory endpoints (auth handled by Traefik ForwardAuth)
	memoryGroup := app.Group("/memory")
	memoryGroup.Post("/save", memory.SaveBlock)
	memoryGroup.Get("/query", memory.QueryBlocks)
	memoryGroup.Get("/query-debug", memory.QueryBlocksDebug)
	memoryGroup.Get("/recent", memory.GetRecent)
	memoryGroup.Delete("/:id", memory.DeleteBlock)
	memoryGroup.Post("/promote", memory.PromoteBlocks)
	memoryGroup.Post("/inject", memory.HandleMemoryInject)

	// Tag extraction and querying endpoints (auth handled by Traefik ForwardAuth)
	tagGroup := app.Group("/tags")
	tagGroup.Post("/extract", memory.ExtractTagsFromPrompt)
	tagGroup.Post("/query", memory.QueryBlocksByPrompt)
	tagGroup.Post("/delta", memory.UpdateTagDelta)

	// Semantic search endpoint (auth handled by Traefik ForwardAuth)
	searchGroup := app.Group("/search")
	searchGroup.Post("/semantic", memory.SemanticSearch)

	// Events endpoints (auth handled by Traefik ForwardAuth)
	eventsGroup := app.Group("/events")
	eventsGroup.Post("/pack", memory.HandlePackEvent)

	// Background processing endpoints (auth handled by Traefik ForwardAuth)
	processingGroup := app.Group("/processing")
	processingGroup.Post("/backfill-tags", memory.BackfillTagsFlat)

	// Admin endpoints (require authentication and admin role)
	adminGroup := app.Group("/admin", middleware.RequireAdminRole())
	adminGroup.Get("/stats", handleAdminStats)
	adminGroup.Post("/cleanup", handleAdminCleanup)
	adminGroup.Post("/edges/prune", handleAdminEdgePrune)
	adminGroup.Get("/features", feature.HandleGetFeatures)
	adminGroup.Post("/features", feature.HandleSetFeatures)
}

// handleAdminStats handles admin statistics endpoint
func handleAdminStats(c *fiber.Ctx) error {
	// Get user information from validated middleware context
	orgID := middleware.GetOrgID(c)
	if orgID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Get real statistics from database
	totalMemoryBlocks, err := db.GetTotalMemoryBlocks(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get memory block count",
		})
	}

	totalTags, err := db.GetTotalTags(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get tag count",
		})
	}

	lastCleanupTime, err := db.GetLastCleanupTime(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get last cleanup time",
		})
	}

	result := fiber.Map{
		"total_memory_blocks": totalMemoryBlocks,
		"total_tags":          totalTags,
		"last_cleanup_time":   lastCleanupTime,
		"uptime_seconds":      getUptimeSeconds(),
		"memory_usage_mb":     getMemoryUsageMB(),
		"requests_per_minute": getRequestsPerMinute(),
		"timestamp":           time.Now().UTC().Format(time.RFC3339),
		"requested_by":        orgID,
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// handleAdminCleanup handles admin cleanup endpoint
func handleAdminCleanup(c *fiber.Ctx) error {
	// Get organization context from validated middleware
	orgID := middleware.GetOrgID(c)
	if orgID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Perform real cleanup operations
	cleanedBlocks, err := db.PerformMemoryCleanup(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to perform memory cleanup",
		})
	}

	cleanedTags, err := db.PerformTagCleanup(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to perform tag cleanup",
		})
	}

	cleanupDuration, err := db.GetCleanupDuration(orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get cleanup duration",
		})
	}

	result := fiber.Map{
		"status":         "completed",
		"cleaned_blocks": cleanedBlocks,
		"cleaned_tags":   cleanedTags,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"duration_ms":    cleanupDuration,
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

// handleAdminEdgePrune handles admin edge pruning endpoint
func handleAdminEdgePrune(c *fiber.Ctx) error {
	// Get organization context from validated middleware
	orgID := middleware.GetOrgID(c)
	if orgID == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Organization context required",
		})
	}

	// Parse request body
	var request struct {
		OrgID string  `json:"orgId"`
		Tau   float64 `json:"tau"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if request.OrgID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "orgId is required",
		})
	}

	// Perform edge pruning
	// Edge pruning is now handled by the new edge system
	// For now, return success without pruning
	prunedCount := 0

	result := fiber.Map{
		"status":       "completed",
		"pruned_count": prunedCount,
		"org_id":       request.OrgID,
		"tau":          request.Tau,
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	}

	return c.Status(fiber.StatusOK).JSON(result)
}

var processStartTime = time.Now()
var requestCounter int64
var lastMinute int64

func getUptimeSeconds() int64 {
	return time.Now().Unix() - processStartTime.Unix()
}

func getMemoryUsageMB() float64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return float64(m.Alloc) / 1024 / 1024
}

func getRequestsPerMinute() int {
	now := time.Now().Unix()
	currentMinute := now / 60

	if currentMinute != atomic.LoadInt64(&lastMinute) {
		atomic.StoreInt64(&lastMinute, currentMinute)
		atomic.StoreInt64(&requestCounter, 0)
	}

	return int(atomic.LoadInt64(&requestCounter))
}
