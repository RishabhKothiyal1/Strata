package main

import (
	"context"
	"database/sql"
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

// embeddingDim is the fixed vector dimensionality used for all collections.
// Hash-based bag-of-words with bigram subword features — no external API needed.
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
// Server
// ---------------------------------------------------------------------------

type server struct {
	db *sql.DB
}

func main() {
	slog.Info("Starting NovaBase AI Service (pgvector + local embeddings)...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://novabase_admin:novabase_secure_pass_123@novabase-postgres:5432/novabase?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8086"
	}

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

	s := &server{db: db}

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Handle("/metrics", promhttp.Handler())

	r.Get("/v1/ai/health", s.handleHealth)
	r.Get("/v1/ai/collections", s.handleListCollections)
	r.Post("/v1/ai/collections", s.handleCreateCollection)
	r.Delete("/v1/ai/collections/{name}", s.handleDeleteCollection)
	r.Get("/v1/ai/collections/{name}/documents", s.handleListDocuments)
	r.Post("/v1/ai/collections/{name}/documents", s.handleAddDocument)
	r.Delete("/v1/ai/collections/{name}/documents/{id}", s.handleDeleteDocument)
	r.Post("/v1/ai/collections/{name}/search", s.handleSearch)

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
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
	// Enable pgvector extension (already installed in the postgres image)
	if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector;`); err != nil {
		return fmt.Errorf("vector extension: %w", err)
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
//
// Hash-based bag-of-words with unigram + bigram character features.
// 512-dimensional, L2-normalized. Fully deterministic, zero external deps.
// Provides meaningful cosine similarity for keyword-overlapping documents.
// ---------------------------------------------------------------------------

func embed(text string) []float32 {
	vec := make([]float32, embeddingDim)
	words := strings.Fields(strings.ToLower(text))

	for _, word := range words {
		word = strings.Trim(word, `.,!?;:"'()-[]{}`)
		if word == "" {
			continue
		}

		// Full word → primary feature
		h := fnv.New32a()
		h.Write([]byte(word))
		vec[h.Sum32()%embeddingDim] += 1.0

		// Character bigrams → subword morphology features
		for i := 0; i < len(word)-1; i++ {
			hb := fnv.New32a()
			hb.Write([]byte(word[i : i+2]))
			vec[(hb.Sum32()+uint32(i)*31)%embeddingDim] += 0.4
		}
	}

	// L2 normalization for cosine similarity
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

// vectorToSQL serializes a float32 slice to PostgreSQL vector literal format.
func vectorToSQL(v []float32) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = strconv.FormatFloat(float64(f), 'f', 8, 32)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":          "ok",
		"service":         "ai-go",
		"embedding_model": fmt.Sprintf("local-hash-bgram-%dd", embeddingDim),
	})
}

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
		writeError(w, http.StatusConflict, "create failed: "+err.Error())
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
	slog.Info("AI collection deleted", "name", name)
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

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, collection_id, content, metadata, created_at
		FROM ai_documents WHERE collection_id=$1
		ORDER BY created_at DESC
	`, collID)
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
	writeJSON(w, http.StatusOK, docs)
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

	slog.Info("AI document added", "collection", name, "doc_id", id)
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

	// pgvector <=> is cosine distance; score = 1 - distance
	rows, err := s.db.QueryContext(r.Context(), fmt.Sprintf(`
		SELECT id, collection_id, content, metadata, created_at,
		       1 - (embedding <=> '%s'::vector) AS score
		FROM ai_documents
		WHERE collection_id = $1
		ORDER BY score DESC
		LIMIT $2
	`, querySQL), collID, req.TopK)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed: "+err.Error())
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

	slog.Info("AI semantic search", "collection", name, "query", req.Query, "hits", len(results))
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
