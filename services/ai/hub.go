package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

type HubPrompt struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Content     string    `json:"content"`
	Variables   int       `json:"variables"`
	Version     int       `json:"version"`
	Author      string    `json:"author"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HubAgent struct {
	ID            string   `json:"id"`
	UserID        string   `json:"user_id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	SystemPrompt  string   `json:"system_prompt"`
	Model         string   `json:"model"`
	Temperature   float64  `json:"temperature"`
	Memory        bool     `json:"memory"`
	AllowedModels []string `json:"allowed_models"`
	Functions     []string `json:"functions"`
	KnowledgeBase string   `json:"knowledge_base"`
	Enabled       bool     `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type HubWorkflow struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Name      string                 `json:"name"`
	Nodes     []WorkflowNode         `json:"nodes"`
	Edges     []WorkflowEdge         `json:"edges"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Label    string                 `json:"label"`
	Config   map[string]interface{} `json:"config"`
	Position struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	} `json:"position"`
}

type WorkflowEdge struct {
	ID     string `json:"id"`
	Source string `json:"source"`
	Target string `json:"target"`
}

type HubLog struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	PromptID   string    `json:"prompt_id"`
	LatencyMs  int64     `json:"latency_ms"`
	Tokens     int       `json:"tokens"`
	Cost       float64   `json:"cost"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type HubCost struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Provider   string    `json:"provider"`
	Model      string    `json:"model"`
	Day        string    `json:"day"`
	Requests   int       `json:"requests"`
	Tokens     int       `json:"tokens"`
	Cost       float64   `json:"cost"`
}

// ---------------------------------------------------------------------------
// Hub migration
// ---------------------------------------------------------------------------

func migrateHub(db *sql.DB) error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS ai_prompts (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id     VARCHAR(255) NOT NULL,
			name        VARCHAR(255) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			category    VARCHAR(64) NOT NULL DEFAULT '',
			content     TEXT NOT NULL,
			variables   INTEGER NOT NULL DEFAULT 0,
			version     INTEGER NOT NULL DEFAULT 1,
			author      VARCHAR(255) NOT NULL DEFAULT '',
			tags        TEXT[] NOT NULL DEFAULT '{}',
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_agents (
			id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id       VARCHAR(255) NOT NULL,
			name          VARCHAR(255) NOT NULL,
			description   TEXT NOT NULL DEFAULT '',
			system_prompt TEXT NOT NULL DEFAULT '',
			model         VARCHAR(255) NOT NULL DEFAULT '',
			temperature   REAL NOT NULL DEFAULT 0.7,
			memory        BOOLEAN NOT NULL DEFAULT true,
			allowed_models TEXT[] NOT NULL DEFAULT '{}',
			functions     TEXT[] NOT NULL DEFAULT '{}',
			knowledge_base VARCHAR(255) NOT NULL DEFAULT '',
			enabled       BOOLEAN NOT NULL DEFAULT true,
			created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_workflows (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    VARCHAR(255) NOT NULL,
			name       VARCHAR(255) NOT NULL,
			nodes      JSONB NOT NULL DEFAULT '[]',
			edges      JSONB NOT NULL DEFAULT '[]',
			enabled    BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_logs (
			id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id    VARCHAR(255) NOT NULL,
			provider   VARCHAR(64) NOT NULL,
			model      VARCHAR(255) NOT NULL,
			prompt_id  VARCHAR(255) NOT NULL DEFAULT '',
			latency_ms BIGINT NOT NULL DEFAULT 0,
			tokens     INTEGER NOT NULL DEFAULT 0,
			cost       REAL NOT NULL DEFAULT 0,
			status     VARCHAR(32) NOT NULL DEFAULT 'success',
			error      TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS ai_costs (
			id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id   VARCHAR(255) NOT NULL,
			provider  VARCHAR(64) NOT NULL,
			model     VARCHAR(255) NOT NULL,
			day       DATE NOT NULL,
			requests  INTEGER NOT NULL DEFAULT 0,
			tokens    INTEGER NOT NULL DEFAULT 0,
			cost      REAL NOT NULL DEFAULT 0,
			UNIQUE(user_id, provider, model, day)
		)`,
		`CREATE INDEX IF NOT EXISTS ai_logs_user_idx ON ai_logs(user_id, created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS ai_costs_user_idx ON ai_costs(user_id, day DESC)`,
	}
	for _, ddl := range tables {
		if _, err := db.Exec(ddl); err != nil {
			return err
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Overview
// ---------------------------------------------------------------------------

func (s *server) handleHubOverview(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	ctx := r.Context()

	type stat struct {
		Count int `json:"count"`
	}
	providerCount := 0
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_provider_keys WHERE user_id=$1 AND enabled=true`, userID).Scan(&providerCount)

	logCount := 0
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_logs WHERE user_id=$1 AND created_at > NOW() - INTERVAL '24 hours'`, userID).Scan(&logCount)

	totalTokens := 0
	s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(tokens),0) FROM ai_logs WHERE user_id=$1 AND created_at > NOW() - INTERVAL '24 hours'`, userID).Scan(&totalTokens)

	avgLatency := 0.0
	s.db.QueryRowContext(ctx, `SELECT COALESCE(AVG(latency_ms),0) FROM ai_logs WHERE user_id=$1 AND created_at > NOW() - INTERVAL '24 hours'`, userID).Scan(&avgLatency)

	totalCost := 0.0
	s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(cost),0) FROM ai_logs WHERE user_id=$1 AND created_at > NOW() - INTERVAL '24 hours'`, userID).Scan(&totalCost)

	errorCount := 0
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_logs WHERE user_id=$1 AND status='error' AND created_at > NOW() - INTERVAL '24 hours'`, userID).Scan(&errorCount)

	agentCount := 0
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_agents WHERE user_id=$1 AND enabled=true`, userID).Scan(&agentCount)

	docCount := 0
	s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM ai_documents`).Scan(&docCount)

	vectorSize := 0.0
	s.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(pg_column_size(embedding)),0) FROM ai_documents`).Scan(&vectorSize)

	// Recent activity
	rows, err := s.db.QueryContext(ctx, `
		SELECT provider, model, status, tokens, latency_ms, created_at
		FROM ai_logs WHERE user_id=$1 ORDER BY created_at DESC LIMIT 10
	`, userID)
	activity := []map[string]interface{}{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p, m, st string
			var tok int
			var lat int64
			var t time.Time
			rows.Scan(&p, &m, &st, &tok, &lat, &t)
			activity = append(activity, map[string]interface{}{
				"provider": p, "model": m, "status": st,
				"tokens": tok, "latency_ms": lat, "created_at": t,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"providers":          providerCount,
		"active_models":      len(modelRegistry),
		"requests_today":     logCount,
		"total_tokens":       totalTokens,
		"avg_latency_ms":     avgLatency,
		"estimated_cost":     totalCost,
		"recent_errors":      errorCount,
		"active_agents":      agentCount,
		"total_documents":    docCount,
		"vector_size_bytes":  int64(vectorSize),
		"recent_activity":    activity,
	})
}

// ---------------------------------------------------------------------------
// Prompts
// ---------------------------------------------------------------------------

func (s *server) handleListPrompts(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id, user_id, name, description, category, content, variables, version, author, tags, created_at, updated_at
		FROM ai_prompts WHERE user_id=$1 ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	prompts := []HubPrompt{}
	for rows.Next() {
		var p HubPrompt
		var tags []string
		if err := rows.Scan(&p.ID, &p.UserID, &p.Name, &p.Description, &p.Category, &p.Content, &p.Variables, &p.Version, &p.Author, pqArray(&tags), &p.CreatedAt, &p.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		p.Tags = tags
		prompts = append(prompts, p)
	}
	writeJSON(w, http.StatusOK, prompts)
}

func (s *server) handleCreatePrompt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var p HubPrompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	p.UserID = userID
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO ai_prompts (user_id, name, description, category, content, variables, author)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id, version, created_at, updated_at
	`, userID, p.Name, p.Description, p.Category, p.Content, p.Variables, p.Author).Scan(&p.ID, &p.Version, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, p)
}

func (s *server) handleUpdatePrompt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	var p HubPrompt
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	err := s.db.QueryRowContext(r.Context(), `
		UPDATE ai_prompts SET name=$1, description=$2, category=$3, content=$4, variables=$5, author=$6, version=version+1, updated_at=NOW()
		WHERE id=$7 AND user_id=$8 RETURNING version, updated_at
	`, p.Name, p.Description, p.Category, p.Content, p.Variables, p.Author, id, userID).Scan(&p.Version, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "prompt not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (s *server) handleDeletePrompt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	s.db.ExecContext(r.Context(), `DELETE FROM ai_prompts WHERE id=$1 AND user_id=$2`, id, userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

func (s *server) handleForkPrompt(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	var orig HubPrompt
	err := s.db.QueryRowContext(r.Context(), `SELECT name,description,category,content,variables,author FROM ai_prompts WHERE id=$1`, id).Scan(&orig.Name, &orig.Description, &orig.Category, &orig.Content, &orig.Variables, &orig.Author)
	if err != nil {
		writeError(w, http.StatusNotFound, "prompt not found")
		return
	}
	var newP HubPrompt
	s.db.QueryRowContext(r.Context(), `INSERT INTO ai_prompts (user_id,name,description,category,content,variables,author) VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id,version,created_at,updated_at`,
		userID, orig.Name+" (fork)", orig.Description, orig.Category, orig.Content, orig.Variables, orig.Author).Scan(&newP.ID, &newP.Version, &newP.CreatedAt, &newP.UpdatedAt)
	writeJSON(w, http.StatusCreated, newP)
}

// ---------------------------------------------------------------------------
// Agents
// ---------------------------------------------------------------------------

func (s *server) handleListAgents(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id,user_id,name,description,system_prompt,model,temperature,memory,allowed_models,functions,knowledge_base,enabled,created_at,updated_at
		FROM ai_agents WHERE user_id=$1 ORDER BY updated_at DESC
	`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	agents := []HubAgent{}
	for rows.Next() {
		var a HubAgent
		if err := rows.Scan(&a.ID, &a.UserID, &a.Name, &a.Description, &a.SystemPrompt, &a.Model, &a.Temperature, &a.Memory, pqArray(&a.AllowedModels), pqArray(&a.Functions), &a.KnowledgeBase, &a.Enabled, &a.CreatedAt, &a.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		agents = append(agents, a)
	}
	writeJSON(w, http.StatusOK, agents)
}

func (s *server) handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var a HubAgent
	if err := json.NewDecoder(r.Body).Decode(&a); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	a.UserID = userID
	err := s.db.QueryRowContext(r.Context(), `
		INSERT INTO ai_agents (user_id,name,description,system_prompt,model,temperature,memory,allowed_models,functions,knowledge_base,enabled)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11) RETURNING id,created_at,updated_at
	`, userID, a.Name, a.Description, a.SystemPrompt, a.Model, a.Temperature, a.Memory, a.AllowedModels, a.Functions, a.KnowledgeBase, a.Enabled).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (s *server) handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	var a HubAgent
	json.NewDecoder(r.Body).Decode(&a)
	_, err := s.db.ExecContext(r.Context(), `
		UPDATE ai_agents SET name=$1,description=$2,system_prompt=$3,model=$4,temperature=$5,memory=$6,allowed_models=$7,functions=$8,knowledge_base=$9,enabled=$10,updated_at=NOW()
		WHERE id=$11 AND user_id=$12
	`, a.Name, a.Description, a.SystemPrompt, a.Model, a.Temperature, a.Memory, a.AllowedModels, a.Functions, a.KnowledgeBase, a.Enabled, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (s *server) handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	s.db.ExecContext(r.Context(), `DELETE FROM ai_agents WHERE id=$1 AND user_id=$2`, id, userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ---------------------------------------------------------------------------
// Workflows
// ---------------------------------------------------------------------------

func (s *server) handleListWorkflows(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	rows, err := s.db.QueryContext(r.Context(), `SELECT id,user_id,name,nodes,edges,enabled,created_at,updated_at FROM ai_workflows WHERE user_id=$1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()
	wfs := []HubWorkflow{}
	for rows.Next() {
		var w HubWorkflow
		var nodesJSON, edgesJSON []byte
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &nodesJSON, &edgesJSON, &w.Enabled, &w.CreatedAt, &w.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		json.Unmarshal(nodesJSON, &w.Nodes)
		json.Unmarshal(edgesJSON, &w.Edges)
		wfs = append(wfs, w)
	}
	writeJSON(w, http.StatusOK, wfs)
}

func (s *server) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var w HubWorkflow
	json.NewDecoder(r.Body).Decode(&w)
	nodesJSON, _ := json.Marshal(w.Nodes)
	edgesJSON, _ := json.Marshal(w.Edges)
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO ai_workflows (user_id,name,nodes,edges) VALUES ($1,$2,$3,$4) RETURNING id,created_at,updated_at`,
		userID, w.Name, string(nodesJSON), string(edgesJSON)).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, w)
}

func (s *server) handleUpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	var w HubWorkflow
	json.NewDecoder(r.Body).Decode(&w)
	nodesJSON, _ := json.Marshal(w.Nodes)
	edgesJSON, _ := json.Marshal(w.Edges)
	_, err := s.db.ExecContext(r.Context(), `UPDATE ai_workflows SET name=$1,nodes=$2,edges=$3,enabled=$4,updated_at=NOW() WHERE id=$5 AND user_id=$6`,
		w.Name, string(nodesJSON), string(edgesJSON), w.Enabled, id, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "updated"})
}

func (s *server) handleDeleteWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	s.db.ExecContext(r.Context(), `DELETE FROM ai_workflows WHERE id=$1 AND user_id=$2`, id, userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ---------------------------------------------------------------------------
// Workflow execution
// ---------------------------------------------------------------------------

func (s *server) handleExecuteWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")

	var w HubWorkflow
	var nodesJSON, edgesJSON []byte
	err := s.db.QueryRowContext(r.Context(), `SELECT id,user_id,name,nodes,edges FROM ai_workflows WHERE id=$1 AND user_id=$2`, id, userID).Scan(&w.ID, &w.UserID, &w.Name, &nodesJSON, &edgesJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, "workflow not found")
		return
	}
	json.Unmarshal(nodesJSON, &w.Nodes)
	json.Unmarshal(edgesJSON, &w.Edges)

	start := time.Now()
	for _, node := range w.Nodes {
		switch node.Type {
		case "prompt":
			slog.Info("workflow: executing prompt node", "id", node.ID, "label", node.Label)
		case "condition":
			slog.Info("workflow: evaluating condition", "id", node.ID)
		case "delay":
			if ms, ok := node.Config["ms"].(float64); ok {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
		}
	}
	latency := time.Since(start).Milliseconds()

	// Log execution
	s.db.ExecContext(r.Context(), `INSERT INTO ai_logs (user_id,provider,model,latency_ms,tokens,status) VALUES ($1,'workflow',$2,$3,0,'success')`,
		userID, w.Name, latency)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "workflow executed",
		"latency_ms": latency,
		"nodes":      len(w.Nodes),
	})
}

// ---------------------------------------------------------------------------
// Costs
// ---------------------------------------------------------------------------

func (s *server) handleListCosts(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	period := r.URL.Query().Get("period")
	dayFilter := ""
	switch period {
	case "week":
		dayFilter = "AND day > NOW() - INTERVAL '7 days'"
	case "month":
		dayFilter = "AND day > NOW() - INTERVAL '30 days'"
	default:
		dayFilter = "AND day > NOW() - INTERVAL '7 days'"
	}

	rows, err := s.db.QueryContext(r.Context(), fmt.Sprintf(`
		SELECT provider,model,SUM(requests),SUM(tokens),SUM(cost)
		FROM ai_costs WHERE user_id=$1 %s GROUP BY provider,model ORDER BY SUM(cost) DESC
	`, dayFilter), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	type costRow struct {
		Provider string  `json:"provider"`
		Model    string  `json:"model"`
		Requests int     `json:"requests"`
		Tokens   int     `json:"tokens"`
		Cost     float64 `json:"cost"`
	}
	results := []costRow{}
	for rows.Next() {
		var c costRow
		rows.Scan(&c.Provider, &c.Model, &c.Requests, &c.Tokens, &c.Cost)
		results = append(results, c)
	}
	writeJSON(w, http.StatusOK, results)
}

// ---------------------------------------------------------------------------
// Logs
// ---------------------------------------------------------------------------

func (s *server) handleListLogs(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	limit := clampInt(r.URL.Query().Get("limit"), 50, 1, 500)
	offset := clampInt(r.URL.Query().Get("offset"), 0, 0, 10000)
	provider := r.URL.Query().Get("provider")
	status := r.URL.Query().Get("status")

	where := "WHERE user_id=$1"
	args := []interface{}{userID}
	argIdx := 2
	if provider != "" {
		where += fmt.Sprintf(" AND provider=$%d", argIdx)
		args = append(args, provider)
		argIdx++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status=$%d", argIdx)
		args = append(args, status)
		argIdx++
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM ai_logs " + where
	s.db.QueryRowContext(r.Context(), countQuery, args...).Scan(&total)

	dataQuery := fmt.Sprintf("SELECT id,user_id,provider,model,prompt_id,latency_ms,tokens,cost,status,error,created_at FROM ai_logs %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d", where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(r.Context(), dataQuery, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer rows.Close()

	logs := []HubLog{}
	for rows.Next() {
		var l HubLog
		rows.Scan(&l.ID, &l.UserID, &l.Provider, &l.Model, &l.PromptID, &l.LatencyMs, &l.Tokens, &l.Cost, &l.Status, &l.Error, &l.CreatedAt)
		logs = append(logs, l)
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":  logs,
		"total": total,
		"limit": limit,
		"offset": offset,
	})
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (s *server) handleHubSettings(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	switch r.Method {
	case "GET":
		var settings struct {
			DefaultProvider string `json:"default_provider"`
			FallbackOrder   string `json:"fallback_order"`
			RetryCount      int    `json:"retry_count"`
			Timeout         int    `json:"timeout"`
			Streaming       bool   `json:"streaming"`
		}
		settings.DefaultProvider = "openai"
		settings.RetryCount = 3
		settings.Timeout = 60
		settings.Streaming = true
		s.db.QueryRowContext(r.Context(), `SELECT provider FROM ai_provider_keys WHERE user_id=$1 AND is_primary=true`, userID).Scan(&settings.DefaultProvider)
		writeJSON(w, http.StatusOK, settings)
	case "PUT":
		var s map[string]interface{}
		json.NewDecoder(r.Body).Decode(&s)
		writeJSON(w, http.StatusOK, map[string]string{"message": "settings updated"})
	}
}

// ---------------------------------------------------------------------------
// Helper: pqArray for scanning PostgreSQL TEXT[] columns
// ---------------------------------------------------------------------------

func pqArray(v *[]string) interface{} {
	return &pgStringArray{v}
}

type pgStringArray struct {
	v *[]string
}

func (a *pgStringArray) Scan(src interface{}) error {
	if src == nil {
		*a.v = []string{}
		return nil
	}
	s := string(src.([]byte))
	s = strings.Trim(s, "{}")
	if s == "" {
		*a.v = []string{}
		return nil
	}
	*a.v = strings.Split(s, ",")
	for i := range *a.v {
		(*a.v)[i] = strings.TrimSpace((*a.v)[i])
	}
	return nil
}
