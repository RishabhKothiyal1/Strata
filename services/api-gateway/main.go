package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/novabase/novabase/services/api-gateway/config"
	"github.com/novabase/novabase/services/api-gateway/router"
)

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "NovaBase API Gateway",
    "description": "Enterprise-ready Backend-as-a-Service Core Gateway API Specification.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "",
      "description": "Base Gateway URL"
    }
  ],
  "paths": {
    "/v1/health": {
      "get": {
        "summary": "Gateway health status",
        "description": "Checks the health and status of the central API Gateway",
        "responses": {
          "200": {
            "description": "Successful health check response",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "status": { "type": "string", "example": "ok" },
                    "service": { "type": "string", "example": "api-gateway" },
                    "timestamp": { "type": "string", "example": "2026-07-07T23:54:00Z" }
                  }
                }
              }
            }
          }
        }
      }
    },
    "/v1/auth/login": {
      "post": {
        "summary": "Authenticate user",
        "description": "Log in with email and password to receive a JWT access token.",
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "type": "object",
                "properties": {
                  "email": { "type": "string", "format": "email", "example": "user@example.com" },
                  "password": { "type": "string", "format": "password", "example": "secure_pass_123" }
                },
                "required": ["email", "password"]
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Successful login returning tokens",
            "content": {
              "application/json": {
                "schema": {
                  "type": "object",
                  "properties": {
                    "access_token": { "type": "string" },
                    "refresh_token": { "type": "string" },
                    "expires_in": { "type": "integer", "example": 3600 }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Unauthorized due to invalid credentials"
          }
        }
      }
    },
    "/v1/storage/buckets": {
      "get": {
        "summary": "List buckets",
        "description": "Retrieve all storage buckets for the organization. Requires a valid JWT.",
        "security": [
          {
            "BearerAuth": []
          }
        ],
        "responses": {
          "200": {
            "description": "List of buckets",
            "content": {
              "application/json": {
                "schema": {
                  "type": "array",
                  "items": {
                    "type": "object",
                    "properties": {
                      "id": { "type": "string" },
                      "name": { "type": "string" },
                      "public": { "type": "boolean" },
                      "created_at": { "type": "string" }
                    }
                  }
                }
              }
            }
          },
          "401": {
            "description": "Missing or invalid token"
          }
        }
      }
    }
  },
  "components": {
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    }
  }
}`

const swaggerUIHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>NovaBase API Gateway Docs</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui.css" />
    <style>
        html { box-sizing: border-box; overflow: -y-scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #121214; }
        .swagger-ui .topbar { display: none; }
        /* Premium custom styling for dark mode Swagger UI */
        .swagger-ui { filter: invert(88%) hue-rotate(180deg); }
        .swagger-ui .info { margin: 30px 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.11.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/swagger.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.api,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
            window.ui = ui;
        };
    </script>
</body>
</html>`

func main() {
	// 1. Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	slog.Info("Starting NovaBase API Gateway...")

	// 2. Load Configuration
	cfg := config.Load()

	// 3. Connect to Redis for Rate Limiting
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

	// 4. Initialize Router
	r := router.New(cfg, rdb)

	// Expose OpenAPI specs & Swagger UI documentation
	r.Get("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(openAPISpec))
	})
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(swaggerUIHTML))
	})

	// 5. Start Server with graceful shutdown
	server := &http.Server{
		Addr:         cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("API Gateway server listening on port", "port", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down API Gateway gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("Graceful shutdown failed", "error", err)
	}

	if err := rdb.Close(); err != nil {
		slog.Error("Failed to close Redis client", "error", err)
	}

	slog.Info("API Gateway server stopped.")
}
