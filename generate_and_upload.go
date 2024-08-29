package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// PageConfig represents the configuration for a single page
type PageConfig struct {
	PageID   string `yaml:"page_id"`
	PageTitle string `yaml:"page_title"`
	Version   int    `yaml:"version"`
	Fields    map[string]interface{} // Dynamic fields for templating
}

// YAMLFileData represents the structure of the entire YAML file
type YAMLFileData struct {
	CommonData map[string]interface{} `yaml:"common_data"`
	IDCardTemplate  PageConfig             `yaml:"id_card_template"`
	DetailedTemplate  PageConfig             `yaml:"detailed_template"`
}

// Payload structure for Confluence API
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
	// Read environment variables for Confluence
	pat := os.Getenv("CONFLUENCE_PAT")
	confluenceURL := os.Getenv("CONFLUENCE_URL")

	if pat == "" || confluenceURL == "" {
		log.Fatal("Environment variables CONFLUENCE_PAT and CONFLUENCE_URL must be set")
	}

	// Define paths
	yamlFolder := "./yaml_files"
	templateFolder := "./templates"

	// Load templates
	tmpl1, err := template.ParseFiles(filepath.Join(templateFolder, "id_card_template.html"))
	if err != nil {
		log.Fatalf("Failed to load id_card_template.html: %v", err)
	}
	tmpl2, err := template.ParseFiles(filepath.Join(templateFolder, "detailed_template.html"))
	if err != nil {
		log.Fatalf("Failed to load detailed_template.html: %v", err)
	}

	// Iterate through YAML files
	err = filepath.Walk(yamlFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil // Skip directories and errors
		}

		// Read YAML file
		yamlData, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read YAML file %s: %v", path, err)
			return nil // Skip the current file on error
		}

		// Parse YAML file
		var yamlContent YAMLFileData
		err = yaml.Unmarshal(yamlData, &yamlContent)
		if err != nil {
			log.Printf("Failed to parse YAML file %s: %v", path, err)
			return nil // Skip the current file on error
		}

		// Generate HTML from IDcard
		var tmpl1Buffer bytes.Buffer
		err = tmpl1.Execute(&tmpl1Buffer, struct {
			CommonData map[string]interface{}
			PageConfig PageConfig
		}{
			CommonData: yamlContent.CommonData,
			PageConfig: yamlContent.IDCardTemplate,
		})
		if err != nil {
			log.Printf("Failed to generate HTML from id_card_template for %s: %v", path, err)
			return nil // Skip the current file on error
		}
		uploadToConfluence(confluenceURL, pat, tmpl1Buffer.String(), yamlContent.IDCardTemplate)

		// Generate HTML from detailed_template
		var tmpl2Buffer bytes.Buffer
		err = tmpl2.Execute(&tmpl2Buffer, struct {
			CommonData map[string]interface{}
			PageConfig PageConfig
		}{
			CommonData: yamlContent.CommonData,
			PageConfig: yamlContent.Template2,
		})
		if err != nil {
			log.Printf("Failed to generate HTML from detailed_template for %s: %v", path, err)
			return nil // Skip the current file on error
		}
		uploadToConfluence(confluenceURL, pat, tmpl2Buffer.String(), yamlContent.Template2)

		return nil
	})

	if err != nil {
		log.Fatalf("Error walking through YAML folder: %v", err)
	}
}

func uploadToConfluence(confluenceURL, pat, htmlContent string, pageConfig PageConfig) {
	// Create the payload
	payload := Payload{
		Title: pageConfig.PageTitle,
		Type:  "page",
	}
	payload.Version.Number = pageConfig.Version
	payload.Body.Storage.Value = htmlContent
	payload.Body.Storage.Representation = "storage"

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Failed to marshal JSON for page %s: %v", pageConfig.PageID, err)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("PUT", confluenceURL+pageConfig.PageID, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Printf("Failed to create HTTP request for page %s: %v", pageConfig.PageID, err)
		return
	}

	// Set headers and authentication
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pat)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to send HTTP request for page %s: %v", pageConfig.PageID, err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body for page %s: %v", pageConfig.PageID, err)
		return
	}

	// Check the response status
	if resp.StatusCode == http.StatusOK {
		fmt.Printf("Page %s updated successfully!\n", pageConfig.PageID)
	} else {
		fmt.Printf("Failed to update page %s. Status: %s, Response: %s\n", pageConfig.PageID, resp.Status, string(body))
	}
}

