package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/novabase/novabase/services/rest/config"
	"github.com/novabase/novabase/services/rest/db"
	"github.com/novabase/novabase/services/rest/handlers"
)

func main() {
	// 1. Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting NovaBase REST Generator Service...")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Connect to Database
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	// 4. Introspect schema to build Registry
	registry := db.NewRegistry()
	if err := registry.Introspect(dbConn); err != nil {
		slog.Error("Schema introspection failed", "error", err)
		os.Exit(1)
	}

	// 5. Initialize Handlers
	restHandler := handlers.NewRESTHandler(dbConn, registry)
	rpcHandler := handlers.NewRPCHandler(dbConn)

	// 6. Setup Router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger) // Chi standard logger
	r.Use(chimiddleware.Recoverer)

	// Health check endpoint
	r.Get("/v1/rest/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "rest-generator"}`))
	})

	// Dynamic CRUD routing mapping exactly to paths forwarded by the gateway
	r.Route("/v1/rest", func(r chi.Router) {
		// RPC Endpoint for database function execution
		r.Post("/rpc/{function}", rpcHandler.Execute)

		// Table CRUD operations
		r.Get("/{table}", restHandler.Get)
		r.Post("/{table}", restHandler.Post)
		r.Patch("/{table}", restHandler.Patch)
		r.Delete("/{table}", restHandler.Delete)
	})

	// 7. Start HTTP server
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("REST Generator service listening on port", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Periodically re-introspect database schema (e.g. every 5 minutes)
	// to capture schema updates without requiring restarts
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			if err := registry.Introspect(dbConn); err != nil {
				slog.Error("Scheduled schema re-introspection failed", "error", err)
			}
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down REST Generator gracefully...")
	ticker.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP server shutdown failed", "error", err)
	}

	if err := dbConn.Close(); err != nil {
		slog.Error("Failed to close database connection pool", "error", err)
	}

	slog.Info("REST Generator service stopped.")
}
