package handlers

import (
	"encoding/csv"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"os"
	"path/filepath"
	"sitemap-builder/models"
	"sync"
	"time"
	"encoding/xml"
	"io/ioutil"
)

// StartValidation initiates sitemap validation and returns a link to results
func StartValidation(c *fiber.Ctx) error {
	type ValidateRequest struct {
		IndexURL string `query:"index_url"`
	}
	
	req := new(ValidateRequest)
	if err := c.QueryParser(req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}
	
	// Generate unique ID for this validation job
	jobID := fmt.Sprintf("validation_%d", time.Now().Unix())
	resultsFile := filepath.Join("data/validations", jobID + ".csv")
	
	// Create directory if it doesn't exist
	os.MkdirAll("data/validations", os.ModePerm)
	
	// Create CSV file with headers
	file, err := os.Create(resultsFile)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create results file"})
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	writer.Write([]string{"URL", "Status", "StatusCode", "Error"})
	writer.Flush()
	
	// Start validation in background
	go validateSitemapIndexAsync(req.IndexURL, resultsFile)
	
	// Return link to results file
	return c.JSON(fiber.Map{
		"message": "Validation started",
		"job_id": jobID,
		"results_url": fmt.Sprintf("/api/validation/results/%s", jobID),
	})
}

// GetValidationResults returns the current validation results
func GetValidationResults(c *fiber.Ctx) error {
	jobID := c.Params("id")
	resultsFile := filepath.Join("data/validations", jobID + ".csv")
	
	// Check if file exists
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{"error": "Validation job not found"})
	}
	
	// Return file as download
	return c.Download(resultsFile)
}

// Background validation process
func validateSitemapIndexAsync(sitemapURL, resultsFile string) {
	// Open CSV file for appending
	file, err := os.OpenFile(resultsFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Create mutex for safe writing to CSV
	var mutex sync.Mutex

	// Fetch sitemap content
	content, err := fetchURL(sitemapURL)
	if err != nil {
		mutex.Lock()
		writer.Write([]string{sitemapURL, "ERROR", "0", err.Error()})
		writer.Flush()
		mutex.Unlock()
		return
	}

	// First, try to parse as sitemap index
	var index models.XMLSitemapIndex
	if err := xml.Unmarshal(content, &index); err == nil && len(index.Sitemaps) > 0 {
		// Record success for the index
		mutex.Lock()
		writer.Write([]string{sitemapURL, "OK", "200", ""})
		writer.Flush()
		mutex.Unlock()

		// Process each sitemap within the index with limited concurrency
		var wg sync.WaitGroup
		semaphore := make(chan struct{}, 5) // Limit to 5 concurrent requests

		for _, sitemap := range index.Sitemaps {
			wg.Add(1)
			semaphore <- struct{}{} // Acquire semaphore

			go func(sitemapURL string) {
				defer wg.Done()
				defer func() { <-semaphore }() // Release semaphore

				validateSitemap(sitemapURL, writer, &mutex)
			}(sitemap.Loc)
		}

		wg.Wait()
		return
	}

	// If not an index, try to parse as a single sitemap (URL set)
	var urlset models.XMLURLSet
	if err := xml.Unmarshal(content, &urlset); err == nil && len(urlset.URLs) > 0 {
		// Record success for the single sitemap
		mutex.Lock()
		writer.Write([]string{sitemapURL, "OK", "200", ""})
		writer.Flush()
		mutex.Unlock()

		// Validate each URL in the sitemap with limited concurrency
		var wg sync.WaitGroup
		urlSemaphore := make(chan struct{}, 10) // Limit to 10 concurrent URL checks

		for _, url := range urlset.URLs {
			wg.Add(1)
			urlSemaphore <- struct{}{} // Acquire semaphore

			go func(loc string) {
				defer wg.Done()
				defer func() { <-urlSemaphore }() // Release semaphore

				status, statusCode, err := checkURL(loc)

				mutex.Lock()
				if err != nil {
					writer.Write([]string{loc, status, fmt.Sprintf("%d", statusCode), err.Error()})
				} else {
					writer.Write([]string{loc, status, fmt.Sprintf("%d", statusCode), ""})
				}
				writer.Flush()
				mutex.Unlock()
			}(url.Loc)
		}

		wg.Wait()
		return
	}

	// If neither sitemap index nor single sitemap, record an error
	mutex.Lock()
	writer.Write([]string{sitemapURL, "ERROR", "0", "Invalid XML format: not a valid sitemap index or URL set"})
	writer.Flush()
	mutex.Unlock()
}

func validateSitemap(sitemapURL string, writer *csv.Writer, mutex *sync.Mutex) {
	// Fetch sitemap
	content, err := fetchURL(sitemapURL)
	if err != nil {
		mutex.Lock()
		writer.Write([]string{sitemapURL, "ERROR", "0", err.Error()})
		writer.Flush()
		mutex.Unlock()
		return
	}
	
	// Parse sitemap
	var urlset models.XMLURLSet
	if err := xml.Unmarshal(content, &urlset); err != nil {
		mutex.Lock()
		writer.Write([]string{sitemapURL, "ERROR", "0", "Invalid XML: " + err.Error()})
		writer.Flush()
		mutex.Unlock()
		return
	}
	
	// Record success for sitemap
	mutex.Lock()
	writer.Write([]string{sitemapURL, "OK", "200", ""})
	writer.Flush()
	mutex.Unlock()
	
	// Validate URLs with limited concurrency
	var wg sync.WaitGroup
	urlSemaphore := make(chan struct{}, 10) // Limit to 10 concurrent URL checks
	
	for _, url := range urlset.URLs {
		wg.Add(1)
		urlSemaphore <- struct{}{} // Acquire semaphore
		
		go func(loc string) {
			defer wg.Done()
			defer func() { <-urlSemaphore }() // Release semaphore
			
			status, statusCode, err := checkURL(loc)
			
			mutex.Lock()
			if err != nil {
				writer.Write([]string{loc, status, fmt.Sprintf("%d", statusCode), err.Error()})
			} else {
				writer.Write([]string{loc, status, fmt.Sprintf("%d", statusCode), ""})
			}
			writer.Flush()
			mutex.Unlock()
		}(url.Loc)
	}
	
	wg.Wait()
}

func checkURL(url string) (string, int, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	
	resp, err := client.Head(url)
	if err != nil {
		return "ERROR", 0, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return "OK", resp.StatusCode, nil
	}
	
	return "ERROR", resp.StatusCode, fmt.Errorf("status code %d", resp.StatusCode)
}

func fetchURL(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %d", resp.StatusCode)
	}
	
	return ioutil.ReadAll(resp.Body)
}
