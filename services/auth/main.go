package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"

	"github.com/novabase/novabase/services/auth/config"
	"github.com/novabase/novabase/services/auth/db"
	"github.com/novabase/novabase/services/auth/handlers"
)

func main() {
	// 1. Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting NovaBase Auth Service...")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Connect to PostgreSQL
	dbConn, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Database connection failed", "error", err)
		os.Exit(1)
	}

	// 4. Migrate database user schema
	migrateSchema(dbConn)

	// 5. Connect to Redis Session cache
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		slog.Error("Failed to connect to Redis cache", "addr", cfg.RedisAddr, "error", err)
		os.Exit(1)
	}
	slog.Info("Successfully connected to Redis", "addr", cfg.RedisAddr)

	// 6. Setup Handlers
	authHandler := handlers.NewAuthHandler(dbConn, rdb, cfg)

	// 7. Configure router
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	// Gateway mounts routes matching /v1/auth path prefix
	r.Route("/v1/auth", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok", "service": "auth-service"}`))
		})

		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
		r.Post("/logout", authHandler.Logout)

		// Verification route displaying claims injected by the gateway
		r.Get("/me", authHandler.Me)

		// Multi-Factor Authentication setups
		r.Post("/mfa/setup", authHandler.MFASetup)
	})

	// 8. Start HTTP Server
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Auth Service listening on port", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down Auth Service gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful HTTP server shutdown failed", "error", err)
	}

	if err := rdb.Close(); err != nil {
		slog.Error("Failed to close Redis client", "error", err)
	}

	if err := dbConn.Close(); err != nil {
		slog.Error("Failed to close database connection pool", "error", err)
	}

	slog.Info("Auth Service stopped.")
}

func migrateSchema(dbConn *sql.DB) {
	slog.Info("Checking user schema migrations...")
	userTableSQL := `
		CREATE TABLE IF NOT EXISTS public.users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) DEFAULT 'member' NOT NULL,
			org_id VARCHAR(50) DEFAULT 'org_default' NOT NULL,
			mfa_enabled BOOLEAN DEFAULT FALSE,
			mfa_secret VARCHAR(100),
			created_at TIMESTAMP DEFAULT NOW()
		);
	`
	_, err := dbConn.Exec(userTableSQL)
	if err != nil {
		slog.Error("User migration SQL failed", "error", err)
		os.Exit(1)
	}
	slog.Info("Database migrations completed.")
}
