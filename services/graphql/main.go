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

	"github.com/novabase/novabase/services/graphql/config"
	"github.com/novabase/novabase/services/graphql/db"
	"github.com/novabase/novabase/services/graphql/handlers"
	"github.com/novabase/novabase/services/graphql/schema"
)

func main() {
	// 1. Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting NovaBase GraphQL Engine...")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Connect to Database
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	// 4. Connect to NATS JetStream PubSub
	pubsub, err := schema.NewPubSub(cfg.NATSURL)
	if err != nil {
		slog.Error("NATS connection failed", "url", cfg.NATSURL, "error", err)
		os.Exit(1)
	}

	// 5. Introspect Schema
	registry := db.NewRegistry()
	if err := registry.Introspect(dbConn); err != nil {
		slog.Error("Introspection routine failed", "error", err)
		os.Exit(1)
	}

	// 6. Build Executable GraphQL Schema
	gSchema, err := schema.GenerateSchema(dbConn, registry, pubsub)
	if err != nil {
		slog.Error("GraphQL Schema compilation failed", "error", err)
		os.Exit(1)
	}

	// 7. Setup Handler and Router
	graphqlHandler := handlers.NewGraphQLHandler(gSchema)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Health check endpoint
	r.Get("/v1/graphql/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "graphql-engine"}`))
	})

	// Match path mapped and proxied from API Gateway
	r.Handle("/v1/graphql", graphqlHandler)

	// 8. Start HTTP Server
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("GraphQL engine service listening on port", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("GraphQL HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down GraphQL Engine gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP server shutdown failed", "error", err)
	}

	pubsub.Close()

	if err := dbConn.Close(); err != nil {
		slog.Error("Failed to close database connection pool", "error", err)
	}

	slog.Info("GraphQL Engine service stopped.")
}
