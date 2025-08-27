package main

import (
	"log"

	"mme-tagging-service/api"
	"mme-tagging-service/internal/cache"
	"mme-tagging-service/internal/config"
	"mme-tagging-service/internal/db"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Load environment variables and MME configuration
	config.LoadEnv()
	config.LoadMMEConfig()

	// Initialize database
	db.InitMongo()

	// Initialize caches
	cache.InitCaches(config.MME.CacheTTLSeconds)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New())
	// Auth middleware removed - now handled by Traefik ForwardAuth

	// Setup routes
	api.SetupRoutes(app)

	// Start server
	log.Fatal(app.Listen(":8080"))
}
