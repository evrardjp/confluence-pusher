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

// ConfluencePage structure for the configuration file
type ConfluencePage struct {
	PageID    string `json:"page_id"`
	PageTitle string `json:"page_title"`
	Version   int    `json:"version"`
	HTMLContentFile  string `json:"html_content_file"`
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
	apiToken := os.Getenv("CONFLUENCE_API_TOKEN")
	//email := os.Getenv("CONFLUENCE_EMAIL")
	confluenceURL := os.Getenv("CONFLUENCE_URL")

	//if apiToken == "" || email == "" || confluenceURL == "" {
	if apiToken == "" || confluenceURL == "" {
		log.Fatal("Environment variables CONFLUENCE_API_TOKEN, CONFLUENCE_EMAIL, and CONFLUENCE_URL must be set")
	}

	// Read configuration file
	pageMetaFile, err := os.Open("page.json")
	if err != nil {
		log.Fatalf("Failed to open pagemetafile: %v", err)
	}
	defer pageMetaFile.Close()

	var pageMeta ConfluencePage
	if err := json.NewDecoder(pageMetaFile).Decode(&config); err != nil {
		log.Fatalf("Failed to decode Page metadata file: %v", err)
	}

	// Read HTML content from file
	htmlContent, err := ioutil.ReadFile(config.HTMLContentFile)
	if err != nil {
		log.Fatalf("Failed to read HTML file: %v", err)
	}

	// Create the payload
	payload := Payload{
		Title: config.PageTitle,
		Type:  "page",
	}
	payload.Version.Number = config.Version
	payload.Body.Storage.Value = string(htmlContent)
	payload.Body.Storage.Representation = "storage"

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Fatalf("Failed to marshal JSON: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("PUT", confluenceURL+config.PageID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatalf("Failed to create HTTP request: %v", err)
	}

	// Set headers and authentication
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + apiToken)
	//req.SetBasicAuth(email, apiToken)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Check the response status
	if resp.StatusCode == http.StatusOK {
		fmt.Println("Page updated successfully!")
	} else {
		fmt.Printf("Failed to update page. Status: %s, Response: %s\n", resp.Status, string(body))
	}
}

