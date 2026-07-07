package router

import (
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/novabase/novabase/services/api-gateway/config"
	"github.com/novabase/novabase/services/api-gateway/middleware"
)

func New(cfg *config.Config, rdb *redis.Client) chi.Router {
	r := chi.NewRouter()

	// Standard chi middlewares
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)

	// Custom metrics & structured logging
	r.Use(middleware.Metrics)
	r.Use(middleware.Logger)

	// Rate limiting via Redis sliding window (100 requests per 1 minute window)
	r.Use(middleware.RateLimiter(rdb, cfg.RateLimitMax, 1*time.Minute))

	// Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Gateway status / healthcheck
	r.Get("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok", "service": "api-gateway", "timestamp": "` + time.Now().Format(time.RFC3339) + `"}`))
	})

	// Setup Reverse Proxies for versioned microservices
	authProxy := newProxy(cfg.AuthURL)
	restProxy := newProxy(cfg.RestURL)
	graphqlProxy := newProxy(cfg.GraphqlURL)
	realtimeProxy := newProxy(cfg.RealtimeURL)
	storageProxy := newProxy(cfg.StorageURL)
	functionsProxy := newProxy(cfg.FunctionsURL)
	aiProxy := newProxy(cfg.AiURL)

	// Optional JWT parser (adds claims if present, doesn't block public requests)
	jwtOptional := middleware.JWT(cfg.JWTSecret, false)
	// Strict JWT validator (blocks request if missing/invalid JWT)
	jwtRequired := middleware.JWT(cfg.JWTSecret, true)

	// Version 1 Routes
	r.Route("/v1", func(r chi.Router) {
		// Use optional JWT parsing across all routes to enrich downstream headers
		r.Use(jwtOptional)

		// Auth Service routes (public endpoints like login/register)
		r.Mount("/auth", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			authProxy.ServeHTTP(w, req)
		}))

		// REST API Generator routes (Schema queries, public and private REST operations)
		r.Mount("/rest", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			restProxy.ServeHTTP(w, req)
		}))

		// GraphQL Auto-generated endpoint
		r.Mount("/graphql", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			graphqlProxy.ServeHTTP(w, req)
		}))

		// Realtime WebSocket service
		r.Mount("/realtime", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			realtimeProxy.ServeHTTP(w, req)
		}))

		// Protected services (strict authentication checks at the gateway layer)
		r.Group(func(r chi.Router) {
			r.Use(jwtRequired)

			// Storage buckets management and file operations
			r.Mount("/storage", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				storageProxy.ServeHTTP(w, req)
			}))

			// Edge Functions runtime triggers & deployments
			r.Mount("/functions", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				functionsProxy.ServeHTTP(w, req)
			}))

			// AI Engines & vector similarity searches
			r.Mount("/ai", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				aiProxy.ServeHTTP(w, req)
			}))
		})
	})

	return r
}

func newProxy(targetURL string) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		slog.Error("Failed to parse downstream microservice URL", "url", targetURL, "error", err)
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify the Director to correctly map Host and headers for downstream microservices
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
		slog.Debug("Proxying request", "src_path", req.URL.Path, "dst_host", target.Host)
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, req *http.Request, err error) {
		slog.Error("Proxy forwarding error", "path", req.URL.Path, "error", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "Bad Gateway", "message": "Failed to connect to downstream microservice"}`))
	}

	return proxy
}
