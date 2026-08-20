package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// OpenProjectClient interacts with the OpenProject API v3.
type OpenProjectClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOpenProjectClient creates a new OpenProject API client.
func NewOpenProjectClient(baseURL, apiKey string) *OpenProjectClient {
	return &OpenProjectClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// activityRequest is the JSON body for creating an activity (comment) on a work package.
type activityRequest struct {
	Comment activityComment `json:"comment"`
}

// activityComment holds the comment content.
type activityComment struct {
	Raw    string `json:"raw"`
	Format string `json:"format"`
}

// PostComment posts a comment (activity note) on a work package.
// Uses: POST /api/v3/work_packages/{id}/activities
// Auth: Basic auth with "apikey" as username and the API key as password.
func (c *OpenProjectClient) PostComment(workPackageID int, comment string) error {
	url := fmt.Sprintf("%s/api/v3/work_packages/%d/activities", c.baseURL, workPackageID)

	payload := activityRequest{
		Comment: activityComment{
			Raw:    comment,
			Format: "markdown",
		},
	}

	// Try sending, retry once on failure.
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			log.Printf("[openproject] retrying comment post (attempt %d)...", attempt+1)
			time.Sleep(2 * time.Second)
		}

		lastErr = c.doPost(url, payload)
		if lastErr == nil {
			return nil
		}
		log.Printf("[openproject] post comment failed: %v", lastErr)
	}

	return fmt.Errorf("failed to post comment after 2 attempts: %w", lastErr)
}

// doPost performs the actual HTTP POST to OpenProject.
func (c *OpenProjectClient) doPost(url string, payload interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")
	req.Header.Set("X-Forwarded-Ssl", "on")
	req.SetBasicAuth("apikey", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("OpenProject API returned %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
