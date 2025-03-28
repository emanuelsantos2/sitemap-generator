package main

import (
	"log"
	"sitemap-builder/handlers"
	"sitemap-builder/models"
	"sitemap-builder/middleware"
	"github.com/joho/godotenv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"encoding/json"
	"os"
	"path/filepath"
)

// Database connection
var DB *gorm.DB

func initDatabase() {
	godotenv.Load()

	dbPath := "data/sitemap_builder.db"
	dir := filepath.Dir(dbPath)

	// Ensure the directory exists
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Printf("Error creating directory:", err)
		return
	}

	DB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	DB.AutoMigrate(&models.User{},&models.StorageConfig{}, &models.SitemapIndex{}, &models.Sitemap{}, &models.SitemapConfig{}, &models.Datasource{})

	if os.Getenv("INIT_DB") == "true" {
		seedDatabase(DB)
	}else{
		var admin models.User
		if result := DB.Where("role = ?", "admin").First(&admin); result.Error != nil {
			adminUser := models.User{
				Username: "admin",
				Password: "admin123",
				Role:     "admin",
			}
			adminUser.HashPassword()
			DB.Create(&adminUser)
		}
	}
	handlers.SetDB(DB)
}

func setupRoutes(app *fiber.App) {
	godotenv.Load()

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "default_secret_key" // Fallback default key
	}
	// Middleware for protected routes
	app.Use(middleware.NewAuthMiddleware(secretKey))

	// API group
	api := app.Group("/api")
	api.Post("/auth/login", handlers.Login)


	// SitemapIndex routes
	sitemapIndex := api.Group("/sitemap-index")
	sitemapIndex.Use(middleware.AdminOnly)
	sitemapIndex.Get("/", handlers.GetSitemapIndexes)
	sitemapIndex.Get("/:id", handlers.GetSitemapIndex)
	sitemapIndex.Post("/", handlers.CreateSitemapIndex)
	sitemapIndex.Put("/:id", handlers.UpdateSitemapIndex)
	sitemapIndex.Delete("/:id", handlers.DeleteSitemapIndex)

	// Sitemap routes
	sitemap := api.Group("/sitemap")
	sitemap.Use(middleware.AdminOnly)
	sitemap.Get("/", handlers.GetSitemaps)
	sitemap.Get("/:id", handlers.GetSitemap)
	sitemap.Post("/", handlers.CreateSitemap)
	sitemap.Put("/:id", handlers.UpdateSitemap)
	sitemap.Delete("/:id", handlers.DeleteSitemap)

	// SitemapConfig routes
	config := api.Group("/config")
	config.Use(middleware.AdminOnly)
	config.Get("/", handlers.GetConfigs)
	config.Get("/:id", handlers.GetConfig)
	config.Post("/", handlers.CreateConfig)
	config.Put("/:id", handlers.UpdateConfig)
	config.Delete("/:id", handlers.DeleteConfig)

	// Datasource routes
	datasource := api.Group("/datasource")
	datasource.Use(middleware.AdminOnly)
	datasource.Get("/", handlers.GetDatasources)
	datasource.Get("/:id", handlers.GetDatasource)
	datasource.Post("/", handlers.CreateDatasource)
	datasource.Put("/:id", handlers.UpdateDatasource)
	datasource.Delete("/:id", handlers.DeleteDatasource)

	validation := api.Group("/validation")
	validation.Use(middleware.AdminOnly)
	validation.Post("/start", handlers.StartValidation)
	validation.Get("/results/:id", handlers.GetValidationResults)


	// Protected route to trigger sitemap regeneration
	api.Post("/generate", middleware.AdminOnly, handlers.GenerateSitemaps)
}

func main() {
	app := fiber.New()

	// Initialize database
	initDatabase()

	// Logger middleware
	app.Use(logger.New())

	// Setup routes
	setupRoutes(app)

	// Start server
	log.Fatal(app.Listen(":3000"))
}

type InitConfig struct {
	Admin          AdminConfig     `json:"admin"`
	Datasources    []models.Datasource `json:"datasources"`
	StorageConfigs []models.StorageConfig `json:"storage_configs"`
	SitemapIndexes []struct {
		Name     string `json:"name"`
		StorageConfigID uint   `json:"storage_config_id"`
		Sitemaps []struct {
			Name   string `json:"name"`
			Type   string `json:"type"`
			Config struct {
				Datasource      string  `json:"datasource"`
				TableName       string  `json:"table_name"`
				BaseURL         string  `json:"base_url"`
				URLPattern      string  `json:"url_pattern"`
				ChangeFrequency string  `json:"change_frequency"`
				Priority        float64 `json:"priority"`
				PublicationName string  `json:"publication_name"`
				DefaultLanguage string  `json:"default_language"`
			} `json:"config"`
		} `json:"sitemaps"`
	} `json:"sitemap_indexes"`
}

type AdminConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func seedDatabase(db *gorm.DB) {
	// Check if admin user already exists
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount > 0 {
		return
	}

	// Load init config
	file, err := os.Open("init.json")
	if err != nil {
		log.Fatal("Init config file missing: ", err)
	}
	defer file.Close()

	var config InitConfig
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		log.Fatal("Error decoding init config: ", err)
	}

	// Create admin user
	adminUser := models.User{
		Username: config.Admin.Username,
		Password: config.Admin.Password,
		Role:     "admin",
	}
	adminUser.HashPassword()
	db.Create(&adminUser)

	// Create datasources
	datasourceMap := make(map[string]uint)
	for _, ds := range config.Datasources {
		newDS := models.Datasource{
			Name:             ds.Name,
			Type:             ds.Type,
			ConnectionString: ds.ConnectionString,
		}
		db.Create(&newDS)
		datasourceMap[ds.Name] = newDS.ID
	}



	// Create sitemap indexes and sitemaps
	for _, index := range config.SitemapIndexes {
		sitemapIndex := models.SitemapIndex{
            Name: index.Name,
        }
		db.Create(&sitemapIndex)

		for i, sc := range config.StorageConfigs {
			if uint(i) == index.StorageConfigID {
				sc.SitemapIndexID = sitemapIndex.ID
				db.Create(&sc)
			}
		}

		for _, sitemap := range index.Sitemaps {
			newSitemap := models.Sitemap{
				Name:           sitemap.Name,
				Type:		    sitemap.Type,
				SitemapIndexID: sitemapIndex.ID,
			}
			db.Create(&newSitemap)

			config := models.SitemapConfig{
				SitemapID:       newSitemap.ID,
				DatasourceID:    datasourceMap[sitemap.Config.Datasource],
				TableName:       sitemap.Config.TableName,
				BaseURL:         sitemap.Config.BaseURL,
				URLPattern:      sitemap.Config.URLPattern,
				ChangeFrequency: sitemap.Config.ChangeFrequency,
				Priority:        sitemap.Config.Priority,
				PublicationName: sitemap.Config.PublicationName,
				DefaultLanguage: sitemap.Config.DefaultLanguage,
			}
			db.Create(&config)
		}
	}
}