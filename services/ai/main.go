package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log/slog"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const embeddingDim = 512

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

type Collection struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DocCount    int       `json:"doc_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type Document struct {
	ID           string                 `json:"id"`
	CollectionID string                 `json:"collection_id"`
	Content      string                 `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

type SearchResult struct {
	Document
	Score float64 `json:"score"`
}

type createCollectionReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type addDocumentReq struct {
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}

type searchReq struct {
	Query string `json:"query"`
	TopK  int    `json:"top_k"`
}

// ---------------------------------------------------------------------------
// Provider types
// ---------------------------------------------------------------------------

type AIProviderKey struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id"`
	Provider      string    `json:"provider"`
	BaseURL       string    `json:"base_url,omitempty"`
	DefaultModel  string    `json:"default_model"`
	Enabled       bool      `json:"enabled"`
	IsPrimary     bool      `json:"is_primary"`
	FallbackOrder int       `json:"fallback_order,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type createProviderReq struct {
	Provider     string `json:"provider"`
	APIKey       string `json:"api_key"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model"`
}

type updateProviderReq struct {
	APIKey       string `json:"api_key,omitempty"`
	BaseURL      string `json:"base_url,omitempty"`
	DefaultModel string `json:"default_model,omitempty"`
	Enabled      *bool  `json:"enabled,omitempty"`
	IsPrimary    *bool  `json:"is_primary,omitempty"`
}

type chatReq struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
	Provider    string        `json:"provider,omitempty"`
}

type chatResp struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
	Provider string  `json:"provider,omitempty"`
	LatencyMs int64   `json:"latency_ms"`
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

type server struct {
	db *sql.DB
}

func main() {
	slog.Info("Starting Strata AI Service...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://strata_admin:strata_secure_pass_123@strata-postgres:5432/strata?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8086"
	}

	initEncryption()

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	for i := range 30 {
		if err := db.Ping(); err == nil {
			break
		}
		slog.Info("Waiting for PostgreSQL...", "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		slog.Error("PostgreSQL still unavailable after retries", "error", err)
		os.Exit(1)
	}

	if err := migrate(db); err != nil {
		slog.Error("Database migration failed", "error", err)
		os.Exit(1)
	}
	if err := migrateHub(db); err != nil {
		slog.Error("Hub migration failed", "error", err)
		os.Exit(1)
	}

	s := &server{db: db}
	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/v1/ai/health", s.handleHealth)
	r.Get("/v1/ai/models", s.handleListModels)

	r.Get("/v1/ai/providers", s.handleListProviders)
	r.Post("/v1/ai/providers", s.handleCreateProvider)
	r.Put("/v1/ai/providers/{provider}", s.handleUpdateProvider)
	r.Delete("/v1/ai/providers/{provider}", s.handleDeleteProvider)
	r.Post("/v1/ai/providers/{provider}/test", s.handleTestProvider)
	r.Get("/v1/ai/providers/{provider}/models", s.handleProviderModels)

	r.Post("/v1/ai/chat", s.handleChat)
	r.Post("/v1/ai/chat/stream", s.handleChatStream)

	r.Get("/v1/ai/usage", s.handleGetUsage)

	r.Get("/v1/ai/collections", s.handleListCollections)
	r.Post("/v1/ai/collections", s.handleCreateCollection)
	r.Delete("/v1/ai/collections/{name}", s.handleDeleteCollection)
	r.Get("/v1/ai/collections/{name}/documents", s.handleListDocuments)
	r.Post("/v1/ai/collections/{name}/documents", s.handleAddDocument)
	r.Delete("/v1/ai/collections/{name}/documents/{id}", s.handleDeleteDocument)
	r.Post("/v1/ai/collections/{name}/search", s.handleSearch)

	// AI Hub routes
	r.Route("/v1/ai/hub", func(r chi.Router) {
		r.Get("/overview", s.handleHubOverview)

		r.Get("/prompts", s.handleListPrompts)
		r.Post("/prompts", s.handleCreatePrompt)
		r.Put("/prompts/{id}", s.handleUpdatePrompt)
		r.Delete("/prompts/{id}", s.handleDeletePrompt)
		r.Post("/prompts/{id}/fork", s.handleForkPrompt)

		r.Get("/agents", s.handleListAgents)
		r.Post("/agents", s.handleCreateAgent)
		r.Put("/agents/{id}", s.handleUpdateAgent)
		r.Delete("/agents/{id}", s.handleDeleteAgent)

		r.Get("/workflows", s.handleListWorkflows)
		r.Post("/workflows", s.handleCreateWorkflow)
		r.Put("/workflows/{id}", s.handleUpdateWorkflow)
		r.Delete("/workflows/{id}", s.handleDeleteWorkflow)
		r.Post("/workflows/{id}/execute", s.handleExecuteWorkflow)

		r.Get("/costs", s.handleListCosts)
		r.Get("/logs", s.handleListLogs)
		r.Get("/settings", s.handleHubSettings)
		r.Put("/settings", s.handleHubSettings)
	})

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("AI service listening", "addr", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("AI service shut down cleanly.")
}

// ---------------------------------------------------------------------------
// Migration
// ---------------------------------------------------------------------------

func migrate(db *sql.DB) error {
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector;`); err != nil {
		return fmt.Errorf("vector extension: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ai_provider_keys (
			id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id         VARCHAR(255) NOT NULL,
			provider        VARCHAR(64)  NOT NULL,
			encrypted_key   TEXT         NOT NULL,
			base_url        TEXT         NOT NULL DEFAULT '',
			default_model   VARCHAR(255) NOT NULL DEFAULT '',
			enabled         BOOLEAN      NOT NULL DEFAULT true,
			is_primary      BOOLEAN      NOT NULL DEFAULT false,
			fallback_order  INTEGER      NOT NULL DEFAULT 0,
			created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, provider)
		);
	`); err != nil {
		return fmt.Errorf("ai_provider_keys: %w", err)
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS ai_usage_logs (
			id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id           VARCHAR(255)  NOT NULL,
			provider          VARCHAR(64)   NOT NULL,
			model             VARCHAR(255)  NOT NULL,
			latency_ms        BIGINT        NOT NULL DEFAULT 0,
			prompt_tokens     INTEGER       NOT NULL DEFAULT 0,
			completion_tokens INTEGER       NOT NULL DEFAULT 0,
			total_tokens      INTEGER       NOT NULL DEFAULT 0,
			estimated_cost    REAL          NOT NULL DEFAULT 0,
			success           BOOLEAN       NOT NULL DEFAULT true,
			error             TEXT          NOT NULL DEFAULT '',
			created_at        TIMESTAMPTZ   NOT NULL DEFAULT NOW()
		);
	`); err != nil {
		return fmt.Errorf("ai_usage_logs: %w", err)
	}

	if _, err := db.Exec(`
		CREATE INDEX IF NOT EXISTS ai_usage_user_idx ON ai_usage_logs(user_id, created_at DESC);
	`); err != nil {
		return fmt.Errorf("ai_usage_logs index: %w", err)
	}

	_, err := db.Exec(fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS ai_collections (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        VARCHAR(255) UNIQUE NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS ai_documents (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			collection_id UUID NOT NULL REFERENCES ai_collections(id) ON DELETE CASCADE,
			content       TEXT NOT NULL,
			metadata      JSONB NOT NULL DEFAULT '{}',
			embedding     vector(%d),
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS ai_docs_collection_idx
			ON ai_documents(collection_id);

		CREATE INDEX IF NOT EXISTS ai_docs_embedding_ivfflat_idx
			ON ai_documents USING ivfflat (embedding vector_cosine_ops)
			WITH (lists = 10);
	`, embeddingDim))
	return err
}

// ---------------------------------------------------------------------------
// Embedding engine
// ---------------------------------------------------------------------------

func embed(text string) []float32 {
	vec := make([]float32, embeddingDim)
	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		word = strings.Trim(word, `.,!?;:"'()-[]{}`)
		if word == "" {
			continue
		}
		h := fnv.New32a()
		h.Write([]byte(word))
		vec[h.Sum32()%embeddingDim] += 1.0
		for i := 0; i < len(word)-1; i++ {
			hb := fnv.New32a()
			hb.Write([]byte(word[i : i+2]))
			vec[(hb.Sum32()+uint32(i)*31)%embeddingDim] += 0.4
		}
	}
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = float32(math.Sqrt(float64(norm)))
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}

func vectorToSQL(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = strconv.FormatFloat(float64(f), 'f', 8, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// ---------------------------------------------------------------------------
// User ID extraction from headers
// ---------------------------------------------------------------------------

func userIDFromRequest(r *http.Request) string {
	uid := r.Header.Get("X-User-Id")
	if uid != "" {
		return uid
	}
	return "anonymous"
}

// ---------------------------------------------------------------------------
// Handlers - Health & Models
// ---------------------------------------------------------------------------

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	providerCount := len(providerRegistry)
	providers := listProviders()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "ok",
		"service":         "ai-go",
		"embedding_model": fmt.Sprintf("local-hash-bgram-%dd", embeddingDim),
		"providers":       providerCount,
		"available_providers": providers,
	})
}

func (s *server) handleListModels(w http.ResponseWriter, r *http.Request) {
	provider := r.URL.Query().Get("provider")
	var models []ModelInfo
	if provider != "" {
		models = getModelsForProvider(provider)
	} else {
		models = listAllModels()
	}
	writeJSON(w, http.StatusOK, models)
}

// ---------------------------------------------------------------------------
// Handlers - Provider Management
// ---------------------------------------------------------------------------

func (s *server) handleListProviders(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, user_id, provider, encrypted_key, base_url, default_model, enabled, is_primary, fallback_order, created_at, updated_at
		FROM ai_provider_keys
		WHERE user_id = $1
		ORDER BY is_primary DESC, fallback_order ASC, provider ASC
	`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	providers := []AIProviderKey{}
	for rows.Next() {
		var p AIProviderKey
		var encKey string
		if err := rows.Scan(&p.ID, &p.UserID, &p.Provider, &encKey, &p.BaseURL, &p.DefaultModel, &p.Enabled, &p.IsPrimary, &p.FallbackOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.Provider = strings.ToUpper(p.Provider[:1]) + p.Provider[1:]
		providers = append(providers, p)
	}
	writeJSON(w, http.StatusOK, providers)
}

func (s *server) handleCreateProvider(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var req createProviderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.Provider = strings.ToLower(req.Provider)

	if _, err := getProvider(req.Provider); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.APIKey == "" {
		writeError(w, http.StatusBadRequest, "api_key is required")
		return
	}
	if req.DefaultModel == "" {
		models := getModelsForProvider(req.Provider)
		if len(models) > 0 {
			req.DefaultModel = models[0].ID
		}
	}

	encrypted, err := encrypt(req.APIKey)
	if err != nil {
		slog.Error("encryption failed", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to secure API key")
		return
	}

	baseURL := req.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURLs[req.Provider]
	}

	var p AIProviderKey
	err = s.db.QueryRowContext(r.Context(), `
		INSERT INTO ai_provider_keys (user_id, provider, encrypted_key, base_url, default_model)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, provider, base_url, default_model, enabled, is_primary, fallback_order, created_at, updated_at
	`, userID, req.Provider, encrypted, baseURL, req.DefaultModel).Scan(
		&p.ID, &p.UserID, &p.Provider, &p.BaseURL, &p.DefaultModel, &p.Enabled, &p.IsPrimary, &p.FallbackOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeError(w, http.StatusConflict, "provider already configured - use PUT to update")
			return
		}
		writeError(w, http.StatusInternalServerError, "create failed: "+err.Error())
		return
	}

	slog.Info("AI provider configured", "user", userID, "provider", req.Provider)
	writeJSON(w, http.StatusCreated, p)
}

func (s *server) handleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	provider := strings.ToLower(chi.URLParam(r, "provider"))

	var req updateProviderReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.APIKey != "" {
		encrypted, err := encrypt(req.APIKey)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to secure API key")
			return
		}
		updates = append(updates, fmt.Sprintf("encrypted_key = $%d", argIdx))
		args = append(args, encrypted)
		argIdx++
	}
	if req.BaseURL != "" {
		updates = append(updates, fmt.Sprintf("base_url = $%d", argIdx))
		args = append(args, req.BaseURL)
		argIdx++
	}
	if req.DefaultModel != "" {
		updates = append(updates, fmt.Sprintf("default_model = $%d", argIdx))
		args = append(args, req.DefaultModel)
		argIdx++
	}
	if req.Enabled != nil {
		updates = append(updates, fmt.Sprintf("enabled = $%d", argIdx))
		args = append(args, *req.Enabled)
		argIdx++
	}
	if req.IsPrimary != nil {
		if *req.IsPrimary {
			s.db.ExecContext(r.Context(),
				`UPDATE ai_provider_keys SET is_primary = false WHERE user_id = $1`, userID)
		}
		updates = append(updates, fmt.Sprintf("is_primary = $%d", argIdx))
		args = append(args, *req.IsPrimary)
		argIdx++
	}

	if len(updates) == 0 {
		writeError(w, http.StatusBadRequest, "no fields to update")
		return
	}

	updates = append(updates, "updated_at = NOW()")
	query := fmt.Sprintf(`UPDATE ai_provider_keys SET %s WHERE user_id = $%d AND provider = $%d`,
		strings.Join(updates, ", "), argIdx, argIdx+1)
	args = append(args, userID, provider)

	res, err := s.db.ExecContext(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	slog.Info("AI provider updated", "user", userID, "provider", provider)
	writeJSON(w, http.StatusOK, map[string]string{"message": "provider updated"})
}

func (s *server) handleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	provider := strings.ToLower(chi.URLParam(r, "provider"))

	res, err := s.db.ExecContext(r.Context(),
		`DELETE FROM ai_provider_keys WHERE user_id = $1 AND provider = $2`, userID, provider)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "provider not found")
		return
	}

	slog.Info("AI provider deleted", "user", userID, "provider", provider)
	writeJSON(w, http.StatusOK, map[string]string{"message": "provider deleted"})
}

func (s *server) handleTestProvider(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	provider := strings.ToLower(chi.URLParam(r, "provider"))

	cfg, err := s.loadProviderConfig(r.Context(), userID, provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	adapter, err := getProvider(provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	start := time.Now()
	err = adapter.ValidateKey(r.Context(), cfg.APIKey, cfg.BaseURL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"status":   "error",
			"message":  err.Error(),
			"latency_ms": latency,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"message":    "connection successful",
		"latency_ms": latency,
	})
}

func (s *server) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	provider := strings.ToLower(chi.URLParam(r, "provider"))

	cfg, err := s.loadProviderConfig(r.Context(), userID, provider)
	if err != nil {
		staticModels := getModelsForProvider(provider)
		if len(staticModels) > 0 {
			writeJSON(w, http.StatusOK, staticModels)
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	adapter, err := getProvider(provider)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	models, err := adapter.Models(r.Context(), cfg.APIKey, cfg.BaseURL)
	if err != nil {
		staticModels := getModelsForProvider(provider)
		if len(staticModels) > 0 {
			writeJSON(w, http.StatusOK, staticModels)
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch models: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, models)
}

// ---------------------------------------------------------------------------
// Handlers - Chat
// ---------------------------------------------------------------------------

func (s *server) handleChat(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "messages are required")
		return
	}

	start := time.Now()
	result, providerName, err := s.executeChat(r.Context(), userID, &req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		logUsage(r.Context(), s.db, userID, providerName, req.Model, latency, 0, 0, false, err.Error())
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	tokens := 0
	promptTokens := 0
	completionTokens := 0
	if result.Usage != nil {
		tokens = result.Usage.TotalTokens
		promptTokens = result.Usage.PromptTokens
		completionTokens = result.Usage.CompletionTokens
	}

	logUsage(r.Context(), s.db, userID, providerName, result.Model, latency, promptTokens, completionTokens, true, "")

	resp := chatResp{
		ID:        result.ID,
		Model:     result.Model,
		Choices:   result.Choices,
		Usage:     result.Usage,
		Provider:  providerName,
		LatencyMs: latency,
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *server) handleChatStream(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	req.Stream = true

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming not supported")
		return
	}

	start := time.Now()

	providerName := req.Provider
	if providerName == "" {
		modelInfo := getModelInfo(req.Model)
		if modelInfo != nil {
			providerName = modelInfo.Provider
		}
	}
	if providerName == "" {
		primary, err := s.loadPrimaryProvider(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "no provider configured: "+err.Error())
			return
		}
		providerName = primary
	}

	cfg, err := s.loadProviderConfig(r.Context(), userID, providerName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	adapter, err := getProvider(providerName)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	chatReq := &ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		Stream:      true,
	}

	chunkCount := 0
	err = adapter.ChatStream(r.Context(), chatReq, cfg.APIKey, cfg.BaseURL, func(chunk StreamChunk) {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
		chunkCount++
	})

	latency := time.Since(start).Milliseconds()

	if err != nil {
		logUsage(r.Context(), s.db, userID, providerName, req.Model, latency, 0, 0, false, err.Error())
		return
	}

	logUsage(r.Context(), s.db, userID, providerName, req.Model, latency, chunkCount, chunkCount, true, "")
}

func (s *server) executeChat(ctx context.Context, userID string, req *chatReq) (*ChatResponse, string, error) {
	if req.Provider != "" {
		cfg, err := s.loadProviderConfig(ctx, userID, req.Provider)
		if err != nil {
			return nil, req.Provider, err
		}
		adapter, err := getProvider(req.Provider)
		if err != nil {
			return nil, req.Provider, err
		}
		chatReq := &ChatRequest{
			Model:       req.Model,
			Messages:    req.Messages,
			Temperature: req.Temperature,
			MaxTokens:   req.MaxTokens,
		}
		result, err := adapter.Chat(ctx, chatReq, cfg.APIKey, cfg.BaseURL)
		return result, req.Provider, err
	}

	modelInfo := getModelInfo(req.Model)
	if modelInfo != nil {
		cfg, err := s.loadProviderConfig(ctx, userID, modelInfo.Provider)
		if err == nil {
			adapter, err := getProvider(modelInfo.Provider)
			if err == nil {
				chatReq := &ChatRequest{
					Model:       req.Model,
					Messages:    req.Messages,
					Temperature: req.Temperature,
					MaxTokens:   req.MaxTokens,
				}
				result, err := adapter.Chat(ctx, chatReq, cfg.APIKey, cfg.BaseURL)
				if err == nil {
					return result, modelInfo.Provider, nil
				}
			}
		}
	}

	fallbacks, err := s.loadFallbackProviders(ctx, userID)
	if err != nil || len(fallbacks) == 0 {
		return nil, "", fmt.Errorf("no configured AI providers - add one at /v1/ai/providers")
	}

	return chatWithFallback(ctx, &ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
	}, fallbacks, func(provider string) (*ProviderConfig, error) {
		return s.loadProviderConfig(ctx, userID, provider)
	})
}

// ---------------------------------------------------------------------------
// Handlers - Usage
// ---------------------------------------------------------------------------

func (s *server) handleGetUsage(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 500 {
		limit = l
	}

	stats, err := getUsageStats(r.Context(), s.db, userID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if stats == nil {
		stats = []UsageStats{}
	}
	writeJSON(w, http.StatusOK, stats)
}

// ---------------------------------------------------------------------------
// Provider config loader
// ---------------------------------------------------------------------------

func (s *server) loadProviderConfig(ctx context.Context, userID, provider string) (*ProviderConfig, error) {
	var encryptedKey, baseURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT encrypted_key, base_url FROM ai_provider_keys
		WHERE user_id = $1 AND provider = $2 AND enabled = true
	`, userID, provider).Scan(&encryptedKey, &baseURL)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("provider '%s' not configured or disabled", provider)
	}
	if err != nil {
		return nil, err
	}

	apiKey, err := decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}

	if baseURL == "" {
		baseURL = defaultBaseURLs[provider]
	}

	return &ProviderConfig{APIKey: apiKey, BaseURL: baseURL}, nil
}

func (s *server) loadPrimaryProvider(ctx context.Context, userID string) (string, error) {
	var provider string
	err := s.db.QueryRowContext(ctx, `
		SELECT provider FROM ai_provider_keys
		WHERE user_id = $1 AND enabled = true
		ORDER BY is_primary DESC, fallback_order ASC
		LIMIT 1
	`, userID).Scan(&provider)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("no AI providers configured")
	}
	return provider, err
}

func (s *server) loadFallbackProviders(ctx context.Context, userID string) ([]FallbackProvider, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT provider, default_model FROM ai_provider_keys
		WHERE user_id = $1 AND enabled = true
		ORDER BY is_primary DESC, fallback_order ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fallbacks []FallbackProvider
	for rows.Next() {
		var fb FallbackProvider
		if err := rows.Scan(&fb.Provider, &fb.Model); err != nil {
			return nil, err
		}
		fallbacks = append(fallbacks, fb)
	}
	return fallbacks, nil
}

// ---------------------------------------------------------------------------
// Handlers - Collections & Documents (unchanged)
// ---------------------------------------------------------------------------

func (s *server) handleListCollections(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT c.id, c.name, c.description, c.created_at,
		       COUNT(d.id) AS doc_count
		FROM ai_collections c
		LEFT JOIN ai_documents d ON d.collection_id = c.id
		GROUP BY c.id, c.name, c.description, c.created_at
		ORDER BY c.created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	cols := []Collection{}
	for rows.Next() {
		var c Collection
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt, &c.DocCount); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		cols = append(cols, c)
	}
	writeJSON(w, http.StatusOK, cols)
}

func (s *server) handleCreateCollection(w http.ResponseWriter, r *http.Request) {
	var req createCollectionReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	var c Collection
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO ai_collections (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at
	`, req.Name, req.Description).Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeError(w, http.StatusConflict, "collection already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("AI collection created", "name", req.Name)
	writeJSON(w, http.StatusCreated, c)
}

func (s *server) handleDeleteCollection(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	res, err := s.db.ExecContext(r.Context(),
		`DELETE FROM ai_collections WHERE name=$1`, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "collection deleted"})
}

func (s *server) collectionID(ctx context.Context, name string) (string, error) {
	var id string
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM ai_collections WHERE name=$1`, name).Scan(&id)
	return id, err
}

func (s *server) handleListDocuments(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	collID, err := s.collectionID(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	limit := clampInt(r.URL.Query().Get("limit"), 50, 1, 500)
	offset := clampInt(r.URL.Query().Get("offset"), 0, 0, 10000)

	var total int
	s.db.QueryRowContext(r.Context(), `SELECT COUNT(*) FROM ai_documents WHERE collection_id=$1`, collID).Scan(&total)

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, collection_id, content, metadata, created_at
		FROM ai_documents WHERE collection_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, collID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	docs := []Document{}
	for rows.Next() {
		var d Document
		var metaRaw []byte
		if err := rows.Scan(&d.ID, &d.CollectionID, &d.Content, &metaRaw, &d.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		json.Unmarshal(metaRaw, &d.Metadata)
		docs = append(docs, d)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":   docs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *server) handleAddDocument(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req addDocumentReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "content is required")
		return
	}
	if req.Metadata == nil {
		req.Metadata = map[string]interface{}{}
	}
	collID, err := s.collectionID(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	vec := embed(req.Content)
	vecSQL := vectorToSQL(vec)
	metaJSON, _ := json.Marshal(req.Metadata)
	id := uuid.New().String()
	now := time.Now().UTC()
	_, err = s.db.ExecContext(r.Context(), fmt.Sprintf(`
		INSERT INTO ai_documents (id, collection_id, content, metadata, embedding, created_at)
		VALUES ($1, $2, $3, $4, '%s'::vector, $5)
	`, vecSQL), id, collID, req.Content, string(metaJSON), now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "insert failed: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, Document{
		ID: id, CollectionID: collID, Content: req.Content,
		Metadata: req.Metadata, CreatedAt: now,
	})
}

func (s *server) handleDeleteDocument(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	docID := chi.URLParam(r, "id")
	collID, err := s.collectionID(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	res, err := s.db.ExecContext(r.Context(),
		`DELETE FROM ai_documents WHERE id=$1 AND collection_id=$2`, docID, collID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "document not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "document deleted"})
}

func (s *server) handleSearch(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req searchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Query == "" {
		writeError(w, http.StatusBadRequest, "query is required")
		return
	}
	if req.TopK <= 0 || req.TopK > 100 {
		req.TopK = 5
	}
	collID, err := s.collectionID(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "collection not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	queryVec := embed(req.Query)
	querySQL := vectorToSQL(queryVec)
	rows, err := s.db.QueryContext(r.Context(), fmt.Sprintf(`
		SELECT id, collection_id, content, metadata, created_at,
		       1 - (embedding <=> '%s'::vector) AS score
		FROM ai_documents
		WHERE collection_id = $1
		ORDER BY score DESC
		LIMIT $2
	`, querySQL), collID, req.TopK)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	results := []SearchResult{}
	for rows.Next() {
		var sr SearchResult
		var metaRaw []byte
		if err := rows.Scan(&sr.ID, &sr.CollectionID, &sr.Content, &metaRaw, &sr.CreatedAt, &sr.Score); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		json.Unmarshal(metaRaw, &sr.Metadata)
		results = append(results, sr)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   req.Query,
		"top_k":   req.TopK,
		"results": results,
	})
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func clampInt(s string, defaultVal, minVal, maxVal int) int {
	val, err := strconv.Atoi(s)
	if err != nil || val < minVal {
		return defaultVal
	}
	if val > maxVal {
		return maxVal
	}
	return val
}

// GenerateEncryptionKey generates a 32-byte hex key for ENCRYPTION_KEY env var
func GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}
