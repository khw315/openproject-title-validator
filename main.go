package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("========================================")
	log.Println("  OpenProject Title Validator Webhook")
	log.Println("========================================")

	// Load configuration.
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("[main] configuration error: %v", err)
	}

	log.Printf("[main] port=%s", cfg.Port)
	log.Printf("[main] openproject_url=%s", cfg.OpenProjectURL)
	if cfg.WebhookSecret != "" {
		log.Printf("[main] webhook signature verification: enabled")
	} else {
		log.Printf("[main] webhook signature verification: disabled")
	}

	// Initialize OpenProject client.
	opClient := NewOpenProjectClient(cfg.OpenProjectURL, cfg.OpenProjectAPIKey)

	// Initialize webhook handler.
	webhookHandler := NewWebhookHandler(cfg, opClient)

	// Set up HTTP routes.
	mux := http.NewServeMux()
	mux.Handle("/webhook", webhookHandler)
	mux.HandleFunc("/health", HealthHandler)

	// Add a simple root handler.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprint(w, "OpenProject Title Validator Webhook\n\nEndpoints:\n  POST /webhook  - Receive OpenProject webhooks\n  GET  /health   - Health check\n")
	})

	// Create the HTTP server.
	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine.
	go func() {
		log.Printf("[main] server listening on :%s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[main] server error: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("[main] received signal %s, shutting down...", sig)

	// Give outstanding requests 10 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[main] server forced to shutdown: %v", err)
	}

	log.Println("[main] server stopped gracefully")
}

// loggingMiddleware logs each incoming HTTP request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[http] %s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[http] %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
