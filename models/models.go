package models

import (
	"encoding/xml"
	"time"

	"gorm.io/gorm"
)

// SitemapIndex model
type SitemapIndex struct {
	gorm.Model
	Name           string    `json:"name"`
	LastGeneration time.Time `json:"last_generation"`
	Sitemaps       []Sitemap `json:"sitemaps" gorm:"foreignKey:SitemapIndexID"`
	StorageConfig  StorageConfig `json:"storage_config" gorm:"foreignKey:SitemapIndexID"`
}

// Sitemap model
type Sitemap struct {
	gorm.Model
	Name           string        `json:"name"`
	SitemapIndexID uint          `json:"sitemap_index_id"`
	Config         SitemapConfig `json:"config" gorm:"foreignKey:SitemapID"`
	LastGeneration time.Time     `json:"last_generation"`
	FilePath       string        `json:"file_path"`
	Type		   string  		 `json:"type"`
}

// SitemapConfig model
type SitemapConfig struct {
	gorm.Model
	SitemapID       uint    `json:"sitemap_id"`
	DatasourceID    uint    `json:"datasource_id"`
	TableName       string  `json:"table_name"`
	BaseURL         string  `json:"base_url"`
	URLPattern      string  `json:"url_pattern"` // e.g., "/{language}/{slug}"
	ChangeFrequency string  `json:"change_frequency"`
	Priority        float64 `json:"priority"`

	PublicationName  string  `json:"publication_name" gorm:"default:'Default News Publication'"`
	DefaultLanguage  string  `json:"default_language" gorm:"default:'en'"`
}

// Datasource model
type Datasource struct {
	gorm.Model
	Name             string `json:"name"`
	Type             string `json:"type"` // e.g., "sqlite", "mysql", "postgres"
	ConnectionString string `json:"connection_string"`
}

type StorageConfig struct {
    gorm.Model
    SitemapIndexID uint   `json:"sitemap_index_id"`
    Mode           string `json:"mode" gorm:"default:'local'"` // "local" or "s3"
    Bucket         string `json:"bucket"`
    Region         string `json:"region"`
    Endpoint       string `json:"endpoint"`
    Path           string `json:"path" gorm:"default:'sitemaps/'"`
}

// XML structures for sitemap generation
type XMLSitemapIndex struct {
	XMLName  xml.Name     `xml:"sitemapindex"`
	XMLNS    string       `xml:"xmlns,attr"`
	Sitemaps []XMLSitemap `xml:"sitemap"`
}

type XMLSitemap struct {
	Loc     string `xml:"loc"`
	LastMod string `xml:"lastmod,omitempty"`
}

type XMLURLSet struct {
	XMLName    xml.Name  `xml:"urlset"`
	XMLNS      string    `xml:"xmlns,attr"`
	XMLNSNews  string    `xml:"xmlns:news,attr,omitempty"`
	URLs       []XMLURL  `xml:"url"`
}

type XMLURL struct {
	Loc        string    `xml:"loc"`
	LastMod    string    `xml:"lastmod,omitempty"`
	ChangeFreq string    `xml:"changefreq,omitempty"`
	Priority   float64   `xml:"priority,omitempty"`
	News       *XMLNews  `xml:"news:news,omitempty"`
}

type XMLNews struct {
	Publication      XMLPublication `xml:"news:publication"`
	PublicationDate  string         `xml:"news:publication_date"`
	Title            string         `xml:"news:title"`
	Genres           string         `xml:"news:genres,omitempty"`
	Keywords         string         `xml:"news:keywords,omitempty"`
	StockTickers     string         `xml:"news:stock_tickers,omitempty"`
}

type XMLPublication struct {
	Name     string `xml:"news:name"`
	Language string `xml:"news:language"`
}
