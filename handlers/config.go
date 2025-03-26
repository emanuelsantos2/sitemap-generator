// handlers/config.go
package handlers

import (
	"sitemap-builder/models"

	"github.com/gofiber/fiber/v2"
)

// GetConfigs returns all sitemap configurations
func GetConfigs(c *fiber.Ctx) error {
	var configs []models.SitemapConfig
	DB.Find(&configs)
	return c.JSON(configs)
}

// GetConfig returns a specific configuration
func GetConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var config models.SitemapConfig
	result := DB.First(&config, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Config not found"})
	}
	return c.JSON(config)
}

// CreateConfig creates a new sitemap configuration
func CreateConfig(c *fiber.Ctx) error {
	config := new(models.SitemapConfig)
	if err := c.BodyParser(config); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if config.SitemapID == 0 || config.DatasourceID == 0 || config.TableName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Check if sitemap exists
	var sitemap models.Sitemap
	if result := DB.First(&sitemap, config.SitemapID); result.Error != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid SitemapID"})
	}

	// Check if datasource exists
	var datasource models.Datasource
	if result := DB.First(&datasource, config.DatasourceID); result.Error != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid DatasourceID"})
	}

	// Check if sitemap already has a config
	var existingConfig models.SitemapConfig
	if result := DB.Where("sitemap_id = ?", config.SitemapID).First(&existingConfig); result.Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Sitemap already has a configuration"})
	}

	DB.Create(&config)
	return c.JSON(config)
}

// UpdateConfig updates a configuration
func UpdateConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var config models.SitemapConfig
	result := DB.First(&config, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Config not found"})
	}

	updateData := new(models.SitemapConfig)
	if err := c.BodyParser(updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate datasource if being updated
	if updateData.DatasourceID != 0 && updateData.DatasourceID != config.DatasourceID {
		var datasource models.Datasource
		if result := DB.First(&datasource, updateData.DatasourceID); result.Error != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid DatasourceID"})
		}
		config.DatasourceID = updateData.DatasourceID
	}

	// Validate sitemap if being updated
	if updateData.SitemapID != 0 && updateData.SitemapID != config.SitemapID {
		// Check if new sitemap exists
		var sitemap models.Sitemap
		if result := DB.First(&sitemap, updateData.SitemapID); result.Error != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid SitemapID"})
		}
		
		// Check if new sitemap already has a config
		var existingConfig models.SitemapConfig
		if result := DB.Where("sitemap_id = ?", updateData.SitemapID).First(&existingConfig); result.Error == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Target sitemap already has a configuration"})
		}
		config.SitemapID = updateData.SitemapID
	}

	// Update other fields
	if updateData.TableName != "" {
		config.TableName = updateData.TableName
	}
	if updateData.BaseURL != "" {
		config.BaseURL = updateData.BaseURL
	}
	if updateData.URLPattern != "" {
		config.URLPattern = updateData.URLPattern
	}
	if updateData.ChangeFrequency != "" {
		config.ChangeFrequency = updateData.ChangeFrequency
	}
	if updateData.Priority != 0 {
		config.Priority = updateData.Priority
	}

	DB.Save(&config)
	return c.JSON(config)
}

// DeleteConfig deletes a configuration
func DeleteConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var config models.SitemapConfig
	result := DB.First(&config, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Config not found"})
	}

	DB.Delete(&config)
	return c.JSON(fiber.Map{"message": "Config deleted"})
}
