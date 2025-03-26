// utils/helpers.go
package utils

import (
	"database/sql"
	"fmt"
	"log"
	"sitemap-builder/models"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/driver/postgres"
	"time"

	"gorm.io/gorm"
)

// BuildURL builds a URL from a pattern and data
func BuildURL(baseURL, pattern string, data map[string]interface{}) string {
	result := pattern
	
	// Replace placeholders in pattern with actual values
	// Example pattern: "/{language}/{slug}"
	for key, value := range data {
		placeholder := fmt.Sprintf("{%s}", key)
		if strings.Contains(result, placeholder) {
			result = strings.Replace(result, placeholder, fmt.Sprintf("%v", value), -1)
		}
	}
	
	// Construct full URL
	if !strings.HasPrefix(result, "http") {
		if !strings.HasPrefix(result, "/") {
			result = "/" + result
		}
		result = "https://" + baseURL + result
	}
	
	return result
}

// ConnectToDatasource establishes a connection to the specified datasource
func ConnectToDatasource(datasource *models.Datasource) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	
	switch datasource.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(datasource.ConnectionString), &gorm.Config{})
	// Add support for other database types as needed
	case "pgsql":
		db, err = gorm.Open(postgres.Open(datasource.ConnectionString), &gorm.Config{})
	// Add support for other database types as needed
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", datasource.Type)
	}
	
	return db, err
}

// ScanRowToMap converts a database row to a map
func ScanRowToMap(rows *sql.Rows) (map[string]interface{}, error) {
	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	
	// Prepare scan destination
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	
	// Scan row into destination
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}
	
	// Convert to map
	rowData := make(map[string]interface{})
	for i, col := range columns {
		var v interface{}
		val := values[i]
		b, ok := val.([]byte)
		if ok {
			v = string(b)
		} else {
			v = val
		}
		rowData[col] = v
	}
	
	return rowData, nil
}

func FormatNewsDate(date interface{}) string {
    dateStr := fmt.Sprintf("%v", date)
    
    // Try different formats
    layouts := []string{
        time.RFC3339,        // Full timestamp with timezone
        "2006-01-02",        // Date-only format
        "2006-01-02T15:04", // Date and time without seconds
		"2006-01-02 15:04:05.999999999 -0700 MST",  // Full format with nanoseconds
        "2006-01-02 15:04:05.999999 -0700 MST", 
    }

    for _, layout := range layouts {
        t, err := time.Parse(layout, dateStr)
        if err == nil {
            return t.Format("2006-01-02")
        }
    }

    // Fallback to current date if parsing fails
    log.Printf("Failed to parse date: %v, using current date", dateStr)
    return time.Now().UTC().Format("2006-01-02")
}

func GetValueOrDefault(value interface{}, def string) string {
    if value == nil || fmt.Sprintf("%v", value) == "" {
        return def
    }
    return fmt.Sprintf("%v", value)
}