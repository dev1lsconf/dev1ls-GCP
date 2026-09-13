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

	"devops-api/internal/config"
	"devops-api/internal/handlers"
)

func main() {
	cfg := config.Load()

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /", handlers.RootHandler)
	mux.HandleFunc("GET /healthz", handlers.HealthzHandler)
	mux.HandleFunc("GET /readyz", handlers.ReadyzHandler)
	mux.HandleFunc("GET /api/v1/info", handlers.InfoHandler(cfg))
	mux.HandleFunc("GET /api/v1/network", handlers.NetworkHandler)
	mux.HandleFunc("GET /metrics", handlers.MetricsHandler)
	mux.HandleFunc("GET /metrics/prometheus", handlers.PrometheusMetricsHandler)
	mux.HandleFunc("POST /api/v1/chaos/toggle-ready", handlers.ChaosToggleReadyHandler)
	mux.HandleFunc("GET /api/v1/chaos/delay", handlers.ChaosDelayHandler)

	// Wrap entire mux with structured JSON logging middleware
	loggedHandler := handlers.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           loggedHandler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Channel to listen for OS signals (SIGTERM from Cloud Run / Kubernetes / Docker)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on port %s (env=%s, version=%s)...", cfg.Port, cfg.Environment, cfg.Version)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()

	// Block until a signal is received
	sig := <-stop
	log.Printf("Received signal %v. Initiating graceful shutdown...", sig)

	// Cloud Run gives instances up to 10 seconds to finish in-flight requests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server stopped cleanly. Goodbye!")
}
