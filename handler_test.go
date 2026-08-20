package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createTestHandler(cfg *Config) (*WebhookHandler, error) {
	client := NewOpenProjectClient(cfg.OpenProjectURL, cfg.OpenProjectAPIKey)
	validator, err := NewValidator(cfg.TitlePattern, cfg.TitleCriteriaDesc, cfg.CommentTemplate)
	if err != nil {
		return nil, err
	}
	return NewWebhookHandler(cfg, client, validator), nil
}

func TestWebhookHandler_MethodNotAllowed(t *testing.T) {
	cfg := &Config{
		OpenProjectURL:    "http://localhost",
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestWebhookHandler_InvalidJSON(t *testing.T) {
	cfg := &Config{
		OpenProjectURL:    "http://localhost",
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestWebhookHandler_MissingAction(t *testing.T) {
	cfg := &Config{
		OpenProjectURL:    "http://localhost",
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	payload := `{}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestWebhookHandler_IgnoresIrrelevantEvent(t *testing.T) {
	cfg := &Config{
		OpenProjectURL:    "http://localhost",
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	payload := `{"action":"project:created"}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(payload))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "ignored") {
		t.Errorf("expected 'ignored' in response body, got: %s", rr.Body.String())
	}
}

func TestWebhookHandler_ValidTitle(t *testing.T) {
	cfg := &Config{
		OpenProjectURL:    "http://localhost",
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	wp := WorkPackage{
		ID:      1,
		Subject: "[TC-001][Login][Form Validation][Ahmad]",
		Links: WorkPackageLinks{
			Author: HALLink{Title: "Ahmad"},
		},
	}
	wpJSON, _ := json.Marshal(wp)

	payload := map[string]interface{}{
		"action":       "work_package:created",
		"work_package": json.RawMessage(wpJSON),
	}
	payloadJSON, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payloadJSON)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"valid":true`) {
		t.Errorf("expected valid:true in response, got: %s", rr.Body.String())
	}
}

func TestWebhookHandler_InvalidTitle_PostsComment(t *testing.T) {
	// Set up a mock OpenProject API server.
	var receivedComment string
	mockOP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/activities") {
			var reqBody struct {
				Comment struct {
					Raw string `json:"raw"`
				} `json:"comment"`
			}
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
				t.Errorf("failed to decode request body: %v", err)
			}
			receivedComment = reqBody.Comment.Raw

			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"_type":"Activity"}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockOP.Close()

	cfg := &Config{
		OpenProjectURL:    mockOP.URL,
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	wp := WorkPackage{
		ID:      42,
		Subject: "Fix login bug",
		Links: WorkPackageLinks{
			Author: HALLink{Title: "Budi"},
		},
	}
	wpJSON, _ := json.Marshal(wp)

	payload := map[string]interface{}{
		"action":       "work_package:created",
		"work_package": json.RawMessage(wpJSON),
	}
	payloadJSON, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payloadJSON)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"comment_posted":true`) {
		t.Errorf("expected comment_posted:true in response, got: %s", rr.Body.String())
	}
	if receivedComment == "" {
		t.Error("expected comment to be posted to OpenProject, but none received")
	}
	if !strings.Contains(receivedComment, "@Budi") {
		t.Errorf("expected comment to mention @Budi, got: %s", receivedComment)
	}
}

func TestWebhookHandler_UpdatedEvent(t *testing.T) {
	mockOP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"_type":"Activity"}`))
	}))
	defer mockOP.Close()

	cfg := &Config{
		OpenProjectURL:    mockOP.URL,
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	wp := WorkPackage{
		ID:      10,
		Subject: "Bad title",
		Links: WorkPackageLinks{
			Author: HALLink{Title: "Siti"},
		},
	}
	wpJSON, _ := json.Marshal(wp)

	payload := map[string]interface{}{
		"action":       "work_package:updated",
		"work_package": json.RawMessage(wpJSON),
	}
	payloadJSON, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payloadJSON)))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"comment_posted":true`) {
		t.Errorf("expected comment_posted:true for updated event, got: %s", rr.Body.String())
	}
}

func TestWebhookHandler_DuplicateUnchangedTitle_SkipsComment(t *testing.T) {
	postCount := 0
	mockOP := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		postCount++
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"_type":"Activity"}`))
	}))
	defer mockOP.Close()

	cfg := &Config{
		OpenProjectURL:    mockOP.URL,
		OpenProjectAPIKey: "test-key",
	}
	handler, err := createTestHandler(cfg)
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	wp := WorkPackage{
		ID:      150,
		Subject: "Invalid Title Here",
		Links: WorkPackageLinks{
			Author: HALLink{Title: "Faisal"},
		},
	}
	wpJSON, _ := json.Marshal(wp)

	payload := map[string]interface{}{
		"action":       "work_package:created",
		"work_package": json.RawMessage(wpJSON),
	}
	payloadJSON, _ := json.Marshal(payload)

	// First request: should post comment
	req1 := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payloadJSON)))
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req1)

	if !strings.Contains(rr1.Body.String(), `"comment_posted":true`) {
		t.Fatalf("first request should post comment, got: %s", rr1.Body.String())
	}
	if postCount != 1 {
		t.Fatalf("expected 1 post, got %d", postCount)
	}

	// Second request (e.g. triggered by OpenProject after comment added): same subject -> should skip!
	payloadUpdate := map[string]interface{}{
		"action":       "work_package:updated",
		"work_package": json.RawMessage(wpJSON),
	}
	payloadUpdateJSON, _ := json.Marshal(payloadUpdate)

	req2 := httptest.NewRequest(http.MethodPost, "/webhook", strings.NewReader(string(payloadUpdateJSON)))
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)

	if !strings.Contains(rr2.Body.String(), `"skipped":true`) {
		t.Errorf("second request should skip, got: %s", rr2.Body.String())
	}
	if postCount != 1 {
		t.Errorf("postCount should still be 1 (no duplicate post), got %d", postCount)
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	HealthHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "healthy") {
		t.Errorf("expected 'healthy' in response, got: %s", rr.Body.String())
	}
}
