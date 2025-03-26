// handlers/sitemap_index.go
package handlers

import (
	"sitemap-builder/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

var DB *gorm.DB

func SetDB(db *gorm.DB) {
	DB = db
}

// GetSitemapIndexes returns all sitemap indexes
func GetSitemapIndexes(c *fiber.Ctx) error {
	var sitemapIndexes []models.SitemapIndex
	DB.Find(&sitemapIndexes)
	return c.JSON(sitemapIndexes)
}

// GetSitemapIndex returns a specific sitemap index
func GetSitemapIndex(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemapIndex models.SitemapIndex
	result := DB.First(&sitemapIndex, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "SitemapIndex not found"})
	}
	return c.JSON(sitemapIndex)
}

// CreateSitemapIndex creates a new sitemap index
func CreateSitemapIndex(c *fiber.Ctx) error {
	sitemapIndex := new(models.SitemapIndex)
	if err := c.BodyParser(sitemapIndex); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	DB.Create(&sitemapIndex)
	return c.JSON(sitemapIndex)
}

// UpdateSitemapIndex updates a sitemap index
func UpdateSitemapIndex(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemapIndex models.SitemapIndex
	result := DB.First(&sitemapIndex, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "SitemapIndex not found"})
	}
	
	if err := c.BodyParser(&sitemapIndex); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}
	
	DB.Save(&sitemapIndex)
	return c.JSON(sitemapIndex)
}

// DeleteSitemapIndex deletes a sitemap index
func DeleteSitemapIndex(c *fiber.Ctx) error {
	id := c.Params("id")
	var sitemapIndex models.SitemapIndex
	result := DB.First(&sitemapIndex, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "SitemapIndex not found"})
	}
	
	DB.Delete(&sitemapIndex)
	return c.JSON(fiber.Map{"message": "SitemapIndex deleted"})
}
