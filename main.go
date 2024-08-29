package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Config structure for the configuration file containing multiple pages
type Config struct {
	Pages []PageConfig `json:"pages"`
}

// PageConfig structure for individual page configuration
type PageConfig struct {
	PageID    string `json:"page_id"`
	PageTitle string `json:"page_title"`
	Version   int    `json:"version"`
	HTMLFile  string `json:"html_file"`
}

// Payload structure for the request body
type Payload struct {
	Version struct {
		Number int `json:"number"`
	} `json:"version"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Body  struct {
		Storage struct {
			Value         string `json:"value"`
			Representation string `json:"representation"`
		} `json:"storage"`
	} `json:"body"`
}

func main() {
	// Read environment variables
	pat := os.Getenv("CONFLUENCE_PAT")
	confluenceURL := os.Getenv("CONFLUENCE_URL")

	if pat == "" || confluenceURL == "" {
		log.Fatal("Environment variables CONFLUENCE_PAT and CONFLUENCE_URL must be set")
	}

	// Read configuration file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("Failed to open config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	if err := json.NewDecoder(configFile).Decode(&config); err != nil {
		log.Fatalf("Failed to decode config file: %v", err)
	}

	// Iterate through each page configuration and update the page
	for _, page := range config.Pages {
		// Read HTML content from file
		htmlContent, err := ioutil.ReadFile(page.HTMLFile)
		if err != nil {
			log.Printf("Failed to read HTML file for page %s: %v", page.PageID, err)
			continue // Skip to the next page if the file can't be read
		}

		// Create the payload
		payload := Payload{
			Title: page.PageTitle,
			Type:  "page",
		}
		payload.Version.Number = page.Version
		payload.Body.Storage.Value = string(htmlContent)
		payload.Body.Storage.Representation = "storage"

		// Convert payload to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Failed to marshal JSON for page %s: %v", page.PageID, err)
			continue // Skip to the next page if JSON marshaling fails
		}

		// Create HTTP request
		req, err := http.NewRequest("PUT", confluenceURL+page.PageID, bytes.NewBuffer(jsonPayload))
		if err != nil {
			log.Printf("Failed to create HTTP request for page %s: %v", page.PageID, err)
			continue // Skip to the next page if the request fails
		}

		// Set headers and authentication
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pat)

		// Send the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Failed to send HTTP request for page %s: %v", page.PageID, err)
			continue // Skip to the next page if the request fails
		}
		defer resp.Body.Close()

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read response body for page %s: %v", page.PageID, err)
			continue // Skip to the next page if reading the response fails
		}

		// Check the response status
		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Page %s updated successfully!\n", page.PageID)
		} else {
			fmt.Printf("Failed to update page %s. Status: %s, Response: %s\n", page.PageID, resp.Status, string(body))
		}
	}
}

