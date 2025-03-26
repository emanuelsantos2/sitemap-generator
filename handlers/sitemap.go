// handlers/sitemap.go
package handlers

import (
	"sitemap-builder/models"

	"github.com/gofiber/fiber/v2"
)

// GetSitemaps returns all sitemaps
func GetSitemaps(c *fiber.Ctx) error {
	var sitemaps []models.Sitemap
	DB.Preload("Config").Find(&sitemaps)
	return c.JSON(sitemaps)
}

// GetSitemap returns a specific sitemap
func GetSitemap(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemap models.Sitemap
	result := DB.Preload("Config").First(&sitemap, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Sitemap not found"})
	}
	return c.JSON(sitemap)
}

// CreateSitemap creates a new sitemap
func CreateSitemap(c *fiber.Ctx) error {
	sitemap := new(models.Sitemap)
	if err := c.BodyParser(sitemap); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Verify the associated sitemap index exists
	var index models.SitemapIndex
	if result := DB.First(&index, sitemap.SitemapIndexID); result.Error != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid SitemapIndexID"})
	}

	DB.Create(&sitemap)
	return c.JSON(sitemap)
}

// UpdateSitemap updates a sitemap
func UpdateSitemap(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemap models.Sitemap
	result := DB.First(&sitemap, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Sitemap not found"})
	}

	// Preserve existing relationships
	currentSitemapIndexID := sitemap.SitemapIndexID
	
	if err := c.BodyParser(&sitemap); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Maintain original index ID unless explicitly changed
	if sitemap.SitemapIndexID == 0 {
		sitemap.SitemapIndexID = currentSitemapIndexID
	}

	DB.Save(&sitemap)
	return c.JSON(sitemap)
}

// DeleteSitemap deletes a sitemap
func DeleteSitemap(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemap models.Sitemap
	result := DB.First(&sitemap, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Sitemap not found"})
	}

	// Delete associated config first
	if sitemap.Config.ID != 0 {
		DB.Delete(&sitemap.Config)
	}

	DB.Delete(&sitemap)
	return c.JSON(fiber.Map{"message": "Sitemap deleted"})
}
