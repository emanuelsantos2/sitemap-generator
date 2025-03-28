package services

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sitemap-builder/models"
	"sitemap-builder/utils"
	"strings"
	"time"

	"gorm.io/gorm"
)

// GenerateAllSitemaps generates all sitemap indexes and their sitemaps
func GenerateAllSitemaps(db *gorm.DB) {
	var sitemapIndexes []models.SitemapIndex
	db.Preload("Sitemaps.Config").Preload("StorageConfig").Find(&sitemapIndexes)
	
	for _, sitemapIndex := range sitemapIndexes {
		err := GenerateSitemapIndex(db, &sitemapIndex)
		if err != nil {
			log.Printf("Error generating sitemap index %s: %v", sitemapIndex.Name, err)
		}
	}
}

// GenerateSitemapIndex generates a specific sitemap index and all its sitemaps
func GenerateSitemapIndex(db *gorm.DB, sitemapIndex *models.SitemapIndex) error {
	xmlIndex := models.XMLSitemapIndex{
		XMLNS:    "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: []models.XMLSitemap{},
	}
	
	outputDir := "sitemaps"
	os.MkdirAll(outputDir, os.ModePerm)

	for _, sitemap := range sitemapIndex.Sitemaps {
		baseFilename := fmt.Sprintf("%s/%s", outputDir, sitemap.Name)
		chunkFiles, err := GenerateSitemap(db, &sitemap, baseFilename, sitemapIndex)
		if err != nil {
			log.Printf("Error generating sitemap %s: %v", sitemap.Name, err)
			continue
		}

		for _, chunkFile := range chunkFiles {
			xmlIndex.Sitemaps = append(xmlIndex.Sitemaps, models.XMLSitemap{
				Loc:     fmt.Sprintf("https://%s/%s", sitemap.Config.BaseURL, chunkFile),
				LastMod: time.Now().Format("2006-01-02"),
			})
		}

		sitemap.LastGeneration = time.Now()
		if len(chunkFiles) > 0 {
			sitemap.FilePath = chunkFiles[0]
		}
		db.Save(&sitemap)
	}

	indexFilename := fmt.Sprintf("%s/%s.xml", outputDir, sitemapIndex.Name)
	if err := writeXMLFile(xmlIndex, indexFilename, sitemapIndex.StorageConfig); err != nil {
		return err
	}

	sitemapIndex.LastGeneration = time.Now()
	db.Save(sitemapIndex)
	return nil
}

// GenerateSitemap generates a sitemap with chunking
func GenerateSitemap(db *gorm.DB, sitemap *models.Sitemap, baseFilename string, sitemapIndex *models.SitemapIndex) ([]string, error) {
	var generatedFiles []string
	var datasource models.Datasource
	if result := db.First(&datasource, sitemap.Config.DatasourceID); result.Error != nil {
		return nil, result.Error
	}


	externalDB, err := utils.ConnectToDatasource(&datasource)
	if err != nil {
		return nil, err
	}

	switch sitemap.Type {
	case "news":
		if strings.ToLower(sitemap.Type) == "news" {
	
			// Build a news sitemap without chunking
			urlSet := models.XMLURLSet{
				XMLNS:     "http://www.sitemaps.org/schemas/sitemap/0.9",
				XMLNSNews: "http://www.google.com/schemas/sitemap-news/0.9",
				URLs:      []models.XMLURL{},
			}
	
			// Query all rows (adjust the query if needed)
			query := fmt.Sprintf("%s", sitemap.Config.TableName)
			rows, err := externalDB.Raw(query).Rows()
			if err != nil {
				return nil, err
			}

			for rows.Next() {
				rowData, err := utils.ScanRowToMap(rows)
				if err != nil {
					rows.Close()
					return nil, err
				}
	
				url := utils.BuildURL(sitemap.Config.BaseURL, sitemap.Config.URLPattern, rowData)
				newsEntry := models.XMLNews{
					Publication: models.XMLPublication{
						Name:     sitemap.Config.PublicationName,
						Language: rowData["language"].(string),
					},
					PublicationDate: utils.FormatNewsDate(rowData["publication_date"]),
					Title:           rowData["title"].(string),
				}
				urlSet.URLs = append(urlSet.URLs, models.XMLURL{
					Loc:  url,
					News: &newsEntry,
				})
			}
			rows.Close()

			// Build single file for news sitemap
			newsFilename := fmt.Sprintf("%s.xml", strings.TrimSuffix(baseFilename, ".xml"))
			if err := writeXMLFile(urlSet, newsFilename, sitemapIndex.StorageConfig); err != nil {
				return nil, err
			}
			
			generatedFiles = append(generatedFiles, newsFilename)
			
			return generatedFiles, nil
		}
	default:
		const chunkSize = 1000
		offset := 0
		chunkNumber := 1

		for {
			query := fmt.Sprintf("%s LIMIT %d OFFSET %d", 
				sitemap.Config.TableName, chunkSize, offset)
			rows, err := externalDB.Raw(query).Rows()
			if err != nil {
				return generatedFiles, err
			}

			urlSet := models.XMLURLSet{
				XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
				URLs:  []models.XMLURL{},
			}

			rowCount := 0
			for rows.Next() {
				rowData, err := utils.ScanRowToMap(rows)
				if err != nil {
					rows.Close()
					return generatedFiles, err
				}

				url := utils.BuildURL(sitemap.Config.BaseURL, 
					sitemap.Config.URLPattern, rowData)
				
				urlSet.URLs = append(urlSet.URLs, models.XMLURL{
					Loc:        url,
					// LastMod:    time.Now().Format("2006-01-02"),
					Priority:   sitemap.Config.Priority,
				})
				rowCount++
			}
			rows.Close()

			if rowCount == 0 {
				break
			}

			chunkFilename := fmt.Sprintf("%s-%04d.xml", 
				strings.TrimSuffix(baseFilename, ".xml"), 
				chunkNumber)

			if err := writeXMLFile(urlSet, chunkFilename, sitemapIndex.StorageConfig); err != nil {
				return generatedFiles, err
			}

			generatedFiles = append(generatedFiles, chunkFilename)
			offset += rowCount
			chunkNumber++
		}

	}
	return generatedFiles, nil
}

// Helper function to write XML files
func writeXMLFile(data interface{}, filename string, storage models.StorageConfig) error {
    xmlData, err := xml.MarshalIndent(data, "", "  ")
    if err != nil {
        return err
    }
    xmlData = append([]byte(xml.Header), xmlData...)

    if storage.Mode == "s3" {
		log.Printf("Uploading to S3: %s", filename)
        key := storage.Path + filename
        return utils.UploadToS3(xmlData, storage.Bucket, key, storage.Region, storage.Endpoint, "application/xml")
    }
	log.Printf("Writing to local file: %s", filename)

    // Local mode
    return ioutil.WriteFile(filename, xmlData, 0644)
}

