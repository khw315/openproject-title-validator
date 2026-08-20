package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPostComment_Success(t *testing.T) {
	var receivedPath string
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path

		// Verify Basic auth.
		user, pass, ok := r.BasicAuth()
		if !ok || user != "apikey" || pass != "test-key" {
			t.Errorf("expected Basic auth with apikey:test-key, got %s:%s (ok=%v)", user, pass, ok)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Verify Content-Type.
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"_type":"Activity","id":1}`))
	}))
	defer server.Close()

	client := NewOpenProjectClient(server.URL, "test-key")
	err := client.PostComment(42, "Test comment")

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if receivedPath != "/api/v3/work_packages/42/activities" {
		t.Errorf("expected path /api/v3/work_packages/42/activities, got %s", receivedPath)
	}

	if receivedBody == nil {
		t.Fatal("expected request body, got nil")
	}

	comment, ok := receivedBody["comment"].(map[string]interface{})
	if !ok {
		t.Fatal("expected 'comment' field in request body")
	}

	if raw, ok := comment["raw"].(string); !ok || raw != "Test comment" {
		t.Errorf("expected comment.raw = 'Test comment', got %v", comment["raw"])
	}
	if format, ok := comment["format"].(string); !ok || format != "markdown" {
		t.Errorf("expected comment.format = 'markdown', got %v", comment["format"])
	}
}

func TestPostComment_ServerError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"_type":"Error","message":"Internal Server Error"}`))
	}))
	defer server.Close()

	client := NewOpenProjectClient(server.URL, "test-key")
	err := client.PostComment(1, "Test comment")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to post comment after 2 attempts") {
		t.Errorf("expected retry failure message, got: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestPostComment_RetrySuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"_type":"Error","message":"Service Unavailable"}`))
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"_type":"Activity","id":1}`))
	}))
	defer server.Close()

	client := NewOpenProjectClient(server.URL, "test-key")
	err := client.PostComment(1, "Test comment")

	if err != nil {
		t.Fatalf("expected success on retry, got: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}
