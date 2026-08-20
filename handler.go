package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// WebhookHandler handles incoming webhook requests from OpenProject
// and validates work package titles.
type WebhookHandler struct {
	config *Config
	client *OpenProjectClient
}

// NewWebhookHandler creates a new WebhookHandler.
func NewWebhookHandler(cfg *Config, client *OpenProjectClient) *WebhookHandler {
	return &WebhookHandler{
		config: cfg,
		client: client,
	}
}

// ServeHTTP implements http.Handler for the webhook endpoint.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only accept POST requests.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body.
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		log.Printf("[webhook] failed to read body: %v", err)
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Verify signature if a secret is configured.
	if h.config.WebhookSecret != "" {
		if !h.verifySignature(r, body) {
			log.Printf("[webhook] signature verification failed")
			http.Error(w, "Invalid signature", http.StatusUnauthorized)
			return
		}
	}

	// Parse the webhook payload.
	var payload WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("[webhook] failed to parse JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if payload.Action == "" {
		log.Printf("[webhook] missing action field")
		http.Error(w, "Missing action field", http.StatusBadRequest)
		return
	}

	log.Printf("[webhook] received event: %s", payload.Action)

	// Only process work_package:created and work_package:updated events.
	if payload.Action != "work_package:created" && payload.Action != "work_package:updated" {
		log.Printf("[webhook] ignoring event: %s", payload.Action)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ignored","reason":"event not relevant"}`)
		return
	}

	// Parse the work package data.
	if len(payload.WorkPackage) == 0 {
		log.Printf("[webhook] missing work_package data")
		http.Error(w, "Missing work_package data", http.StatusBadRequest)
		return
	}

	var wp WorkPackage
	if err := json.Unmarshal(payload.WorkPackage, &wp); err != nil {
		log.Printf("[webhook] failed to parse work_package: %v", err)
		http.Error(w, "Invalid work_package data", http.StatusBadRequest)
		return
	}

	log.Printf("[webhook] work package #%d: %q (author: %s)", wp.ID, wp.Subject, wp.Links.Author.Title)

	// Validate the title.
	valid, violations := ValidateTitle(wp.Subject)
	if valid {
		log.Printf("[webhook] title is valid for WP #%d", wp.ID)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","valid":true}`)
		return
	}

	log.Printf("[webhook] title is INVALID for WP #%d: %v", wp.ID, violations)

	// Build the comment message.
	authorName := wp.Links.Author.Title
	if authorName == "" {
		authorName = "Pengguna"
	}
	comment := BuildCommentMessage(authorName, violations)

	// Post the comment to OpenProject.
	if err := h.client.PostComment(wp.ID, comment); err != nil {
		log.Printf("[webhook] failed to post comment on WP #%d: %v", wp.ID, err)
		http.Error(w, "Failed to post comment", http.StatusInternalServerError)
		return
	}

	log.Printf("[webhook] comment posted on WP #%d", wp.ID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","valid":false,"comment_posted":true}`)
}

// verifySignature checks the X-OP-Signature header against the request body
// using HMAC-SHA1 with the configured webhook secret.
func (h *WebhookHandler) verifySignature(r *http.Request, body []byte) bool {
	signature := r.Header.Get("X-OP-Signature")
	if signature == "" {
		return false
	}

	// OpenProject sends the signature as "sha1=<hex>".
	parts := strings.SplitN(signature, "=", 2)
	if len(parts) != 2 || parts[0] != "sha1" {
		return false
	}

	expectedMAC, err := hex.DecodeString(parts[1])
	if err != nil {
		return false
	}

	mac := hmac.New(sha1.New, []byte(h.config.WebhookSecret))
	mac.Write(body)
	actualMAC := mac.Sum(nil)

	return hmac.Equal(actualMAC, expectedMAC)
}

// HealthHandler responds to health check requests.
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"healthy"}`)
}
