package auth

import "github.com/gofiber/fiber/v2"

// RequireUser middleware ensures X-User-ID header is present
func RequireUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		uid := c.Get("X-User-ID")
		if uid == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "auth required",
			})
		}
		c.Locals("userId", uid)
		return c.Next()
	}
}
