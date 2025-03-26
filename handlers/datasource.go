// handlers/datasource.go
package handlers

import (
	"fmt"
	"sitemap-builder/models"
	"sitemap-builder/utils"

	"github.com/gofiber/fiber/v2"
)

// GetDatasources returns all datasources
func GetDatasources(c *fiber.Ctx) error {
	var datasources []models.Datasource
	DB.Find(&datasources)
	return c.JSON(datasources)
}

// GetDatasource returns a specific datasource
func GetDatasource(c *fiber.Ctx) error {
	id := c.Params("id")
	var datasource models.Datasource
	result := DB.First(&datasource, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Datasource not found"})
	}
	return c.JSON(datasource)
}

// CreateDatasource creates a new datasource
func CreateDatasource(c *fiber.Ctx) error {
	datasource := new(models.Datasource)
	if err := c.BodyParser(datasource); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Validate required fields
	if datasource.Name == "" || datasource.Type == "" || datasource.ConnectionString == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields"})
	}

	// Check for unique name
	var existingDS models.Datasource
	if result := DB.Where("name = ?", datasource.Name).First(&existingDS); result.Error == nil {
		return c.Status(400).JSON(fiber.Map{"error": "Datasource name already exists"})
	}

	// Validate datasource type
	if !isValidDatasourceType(datasource.Type) {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid datasource type",
			"valid_types": []string{"sqlite", "mysql", "postgres"},
		})
	}

	// Test connection
	if err := testDatasourceConnection(datasource); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Connection test failed",
			"details": err.Error(),
		})
	}

	DB.Create(&datasource)
	return c.JSON(datasource)
}

// UpdateDatasource updates a datasource
func UpdateDatasource(c *fiber.Ctx) error {
	id := c.Params("id")
	var datasource models.Datasource
	result := DB.First(&datasource, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Datasource not found"})
	}

	updateData := new(models.Datasource)
	if err := c.BodyParser(updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
	}

	// Check for name conflict
	if updateData.Name != "" && updateData.Name != datasource.Name {
		var existingDS models.Datasource
		if result := DB.Where("name = ?", updateData.Name).First(&existingDS); result.Error == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Datasource name already exists"})
		}
		datasource.Name = updateData.Name
	}

	// Validate type if changing
	if updateData.Type != "" && updateData.Type != datasource.Type {
		if !isValidDatasourceType(updateData.Type) {
			return c.Status(400).JSON(fiber.Map{
				"error": "Invalid datasource type",
				"valid_types": []string{"sqlite", "mysql", "postgres"},
			})
		}
		datasource.Type = updateData.Type
	}

	// Validate connection string if changing
	if updateData.ConnectionString != "" {
		datasource.ConnectionString = updateData.ConnectionString
	}

	// Test connection if any sensitive fields changed
	if updateData.Type != "" || updateData.ConnectionString != "" {
		if err := testDatasourceConnection(&datasource); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"error": "Connection test failed",
				"details": err.Error(),
			})
		}
	}

	DB.Save(&datasource)
	return c.JSON(datasource)
}

// DeleteDatasource deletes a datasource
func DeleteDatasource(c *fiber.Ctx) error {
	id := c.Params("id")
	var datasource models.Datasource
	result := DB.First(&datasource, id)
	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Datasource not found"})
	}

	// Check if any configs reference this datasource
	var configCount int64
	DB.Model(&models.SitemapConfig{}).Where("datasource_id = ?", id).Count(&configCount)
	if configCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot delete datasource - referenced by existing configurations",
			"config_count": configCount,
		})
	}

	DB.Delete(&datasource)
	return c.JSON(fiber.Map{"message": "Datasource deleted"})
}

// Helper function to validate datasource type
func isValidDatasourceType(dsType string) bool {
	validTypes := map[string]bool{
		"sqlite":   true,
		"mysql":    true,
		"pgsql": true,
	}
	return validTypes[dsType]
}

// Helper function to test datasource connection
func testDatasourceConnection(ds *models.Datasource) error {
	// Test connection using utils package
	db, err := utils.ConnectToDatasource(ds)
	if err != nil {
		return fmt.Errorf("connection failed: %v", err)
	}
	
	// Verify we can ping the database
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("connection pool error: %v", err)
	}
	defer sqlDB.Close()
	
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("ping failed: %v", err)
	}
	
	return nil
}
