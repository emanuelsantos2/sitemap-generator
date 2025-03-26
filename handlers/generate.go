// handlers/generate.go
package handlers

import (
	"sitemap-builder/services"

	"github.com/gofiber/fiber/v2"
)

func GenerateSitemaps(c *fiber.Ctx) error {
	// Start sitemap generation in a goroutine
	go services.GenerateAllSitemaps(DB)
	
	return c.JSON(fiber.Map{"message": "Sitemap generation started"})
}
