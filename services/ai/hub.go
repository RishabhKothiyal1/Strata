package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
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
		`CREATE TABLE IF NOT EXISTS ai_settings (
			id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id      VARCHAR(255) NOT NULL,
			setting_key  VARCHAR(128) NOT NULL,
			setting_value TEXT NOT NULL DEFAULT '',
			created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			UNIQUE(user_id, setting_key)
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

func (s *server) handleAgentChat(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	agentID := chi.URLParam(r, "id")

	var agent HubAgent
	err := s.db.QueryRowContext(r.Context(), `
		SELECT id,user_id,name,system_prompt,model,temperature,memory,knowledge_base FROM ai_agents
		WHERE id=$1 AND user_id=$2 AND enabled=true
	`, agentID, userID).Scan(&agent.ID, &agent.UserID, &agent.Name, &agent.SystemPrompt, &agent.Model, &agent.Temperature, &agent.Memory, &agent.KnowledgeBase)
	if err != nil {
		writeError(w, http.StatusNotFound, "agent not found or disabled")
		return
	}

	var req struct {
		Message string `json:"message"`
		History []ChatMessage `json:"history,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	messages := []ChatMessage{}
	if agent.SystemPrompt != "" {
		messages = append(messages, ChatMessage{Role: "system", Content: agent.SystemPrompt})
	}
	if len(req.History) > 0 {
		messages = append(messages, req.History...)
	}
	messages = append(messages, ChatMessage{Role: "user", Content: req.Message})

	provider := ""
	model := agent.Model
	if model == "" {
		primary, err := s.loadPrimaryProvider(r.Context(), userID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "no provider configured")
			return
		}
		provider = primary
		modelInfo := getModelInfo(model)
		if modelInfo != nil {
			provider = modelInfo.Provider
		}
	}

	modelInfo := getModelInfo(model)
	if modelInfo != nil {
		provider = modelInfo.Provider
	}

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
	result, err := adapter.Chat(r.Context(), &ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: agent.Temperature,
	}, cfg.APIKey, cfg.BaseURL)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		s.db.ExecContext(r.Context(), `INSERT INTO ai_logs (user_id,provider,model,latency_ms,tokens,status,error) VALUES ($1,$2,$3,$4,0,'error',$5)`,
			userID, provider, model, latency, err.Error())
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	tokens := 0
	if result.Usage != nil {
		tokens = result.Usage.TotalTokens
	}
	s.db.ExecContext(r.Context(), `INSERT INTO ai_logs (user_id,provider,model,latency_ms,tokens,status) VALUES ($1,$2,$3,$4,$5,'success')`,
		userID, provider, model, latency, tokens)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"response":   result.Choices[0].Message.Content,
		"provider":   provider,
		"model":      result.Model,
		"latency_ms": latency,
		"usage":      result.Usage,
	})
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
		var wf HubWorkflow
		var nodesJSON, edgesJSON []byte
		if err := rows.Scan(&wf.ID, &wf.UserID, &wf.Name, &nodesJSON, &edgesJSON, &wf.Enabled, &wf.CreatedAt, &wf.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		json.Unmarshal(nodesJSON, &wf.Nodes)
		json.Unmarshal(edgesJSON, &wf.Edges)
		wfs = append(wfs, wf)
	}
	writeJSON(w, http.StatusOK, wfs)
}

func (s *server) handleCreateWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var wf HubWorkflow
	json.NewDecoder(r.Body).Decode(&wf)
	nodesJSON, _ := json.Marshal(wf.Nodes)
	edgesJSON, _ := json.Marshal(wf.Edges)
	err := s.db.QueryRowContext(r.Context(), `INSERT INTO ai_workflows (user_id,name,nodes,edges) VALUES ($1,$2,$3,$4) RETURNING id,created_at,updated_at`,
		userID, wf.Name, string(nodesJSON), string(edgesJSON)).Scan(&wf.ID, &wf.CreatedAt, &wf.UpdatedAt)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, wf)
}

func (s *server) handleUpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	id := chi.URLParam(r, "id")
	var wf HubWorkflow
	json.NewDecoder(r.Body).Decode(&wf)
	nodesJSON, _ := json.Marshal(wf.Nodes)
	edgesJSON, _ := json.Marshal(wf.Edges)
	_, err := s.db.ExecContext(r.Context(), `UPDATE ai_workflows SET name=$1,nodes=$2,edges=$3,enabled=$4,updated_at=NOW() WHERE id=$5 AND user_id=$6`,
		wf.Name, string(nodesJSON), string(edgesJSON), wf.Enabled, id, userID)
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

	var wf HubWorkflow
	var nodesJSON, edgesJSON []byte
	err := s.db.QueryRowContext(r.Context(), `SELECT id,user_id,name,nodes,edges FROM ai_workflows WHERE id=$1 AND user_id=$2`, id, userID).Scan(&wf.ID, &wf.UserID, &wf.Name, &nodesJSON, &edgesJSON)
	if err != nil {
		writeError(w, http.StatusNotFound, "workflow not found")
		return
	}
	json.Unmarshal(nodesJSON, &wf.Nodes)
	json.Unmarshal(edgesJSON, &wf.Edges)

	adj := map[string][]string{}
	for _, e := range wf.Edges {
		adj[e.Source] = append(adj[e.Source], e.Target)
	}

	inDegree := map[string]int{}
	for _, e := range wf.Edges {
		inDegree[e.Target]++
	}
	roots := []string{}
	for _, n := range wf.Nodes {
		if inDegree[n.ID] == 0 {
			roots = append(roots, n.ID)
		}
	}

	// Node outputs: map of nodeID -> output text
	outputs := map[string]string{}

	start := time.Now()
	nodeMap := map[string]WorkflowNode{}
	for _, n := range wf.Nodes {
		nodeMap[n.ID] = n
	}

	// BFS execution
	queue := append([]string{}, roots...)
	visited := map[string]bool{}
	totalTokens := 0

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true

		node, ok := nodeMap[current]
		if !ok {
			continue
		}

		switch node.Type {
		case "prompt":
			promptText := ""
			if p, ok := node.Config["prompt"].(string); ok {
				promptText = p
			}
			// Substitute variables from upstream outputs
			for k, v := range outputs {
				promptText = strings.ReplaceAll(promptText, "{{"+k+"}}", v)
			}

			provider := ""
			if p, ok := node.Config["provider"].(string); ok {
				provider = p
			}
			model := ""
			if m, ok := node.Config["model"].(string); ok {
				model = m
			}

			if provider == "" {
				primary, err := s.loadPrimaryProvider(r.Context(), userID)
				if err != nil {
					outputs[current] = fmt.Sprintf("error: %v", err)
					continue
				}
				provider = primary
			}

			cfg, err := s.loadProviderConfig(r.Context(), userID, provider)
			if err != nil {
				outputs[current] = fmt.Sprintf("error: %v", err)
				continue
			}

			adapter, err := getProvider(provider)
			if err != nil {
				outputs[current] = fmt.Sprintf("error: %v", err)
				continue
			}

			result, err := adapter.Chat(r.Context(), &ChatRequest{
				Model:    model,
				Messages: []ChatMessage{{Role: "user", Content: promptText}},
			}, cfg.APIKey, cfg.BaseURL)
			if err != nil {
				outputs[current] = fmt.Sprintf("error: %v", err)
				continue
			}

			output := ""
			if len(result.Choices) > 0 {
				output = result.Choices[0].Message.Content
			}
			outputs[current] = output
			if result.Usage != nil {
				totalTokens += result.Usage.TotalTokens
			}

		case "condition":
			conditionExpr := ""
			if c, ok := node.Config["condition"].(string); ok {
				conditionExpr = c
			}
			for k, v := range outputs {
				conditionExpr = strings.ReplaceAll(conditionExpr, "{{"+k+"}}", v)
			}
			// Simple evaluation: check if condition contains "true" or evaluates
			result := "false"
			if strings.Contains(strings.ToLower(conditionExpr), "true") ||
				strings.EqualFold(conditionExpr, "true") {
				result = "true"
			}
			outputs[current] = result

		case "delay":
			if ms, ok := node.Config["ms"].(float64); ok {
				time.Sleep(time.Duration(ms) * time.Millisecond)
			}
			outputs[current] = "delayed"

		case "function":
			if fn, ok := node.Config["endpoint"].(string); ok && fn != "" {
				httpReq, _ := http.NewRequestWithContext(r.Context(), "POST", fn, nil)
				resp, err := http.DefaultClient.Do(httpReq)
				if err == nil {
					resp.Body.Close()
					outputs[current] = fmt.Sprintf("called %s: HTTP %d", fn, resp.StatusCode)
				} else {
					outputs[current] = fmt.Sprintf("error: %v", err)
				}
			}
		}

		// Enqueue children
		for _, child := range adj[current] {
			queue = append(queue, child)
		}
	}

	latency := time.Since(start).Milliseconds()
	s.db.ExecContext(r.Context(), `INSERT INTO ai_logs (user_id,provider,model,latency_ms,tokens,status) VALUES ($1,'workflow',$2,$3,$4,'success')`,
		userID, wf.Name, latency, totalTokens)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "workflow executed",
		"latency_ms":   latency,
		"nodes":        len(wf.Nodes),
		"total_tokens": totalTokens,
		"outputs":      outputs,
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

	// Grouped by provider+model
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

	// Daily time-series for charts
	seriesRows, err := s.db.QueryContext(r.Context(), fmt.Sprintf(`
		SELECT day::TEXT, SUM(requests), SUM(tokens), SUM(cost)
		FROM ai_costs WHERE user_id=$1 %s GROUP BY day ORDER BY day ASC
	`, dayFilter), userID)
	series := []map[string]interface{}{}
	if err == nil {
		defer seriesRows.Close()
		for seriesRows.Next() {
			var d string
			var reqs, tok int
			var cst float64
			seriesRows.Scan(&d, &reqs, &tok, &cst)
			series = append(series, map[string]interface{}{
				"day": d, "requests": reqs, "tokens": tok, "cost": cst,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"breakdown": results,
		"series":    series,
	})
}

func (s *server) handleSetBudget(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	var req struct {
		MonthlyLimit float64 `json:"monthly_limit"`
		AlertAt      float64 `json:"alert_at"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	_, err := s.db.ExecContext(r.Context(), `
		INSERT INTO ai_settings (user_id, setting_key, setting_value)
		VALUES ($1, 'budget_monthly', $2::TEXT)
		ON CONFLICT (user_id, setting_key) DO UPDATE SET setting_value = $2::TEXT, updated_at = NOW()
	`, userID, fmt.Sprintf("%f", req.MonthlyLimit))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_, err = s.db.ExecContext(r.Context(), `
		INSERT INTO ai_settings (user_id, setting_key, setting_value)
		VALUES ($1, 'budget_alert_at', $2::TEXT)
		ON CONFLICT (user_id, setting_key) DO UPDATE SET setting_value = $2::TEXT, updated_at = NOW()
	`, userID, fmt.Sprintf("%f", req.AlertAt))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "budget set"})
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
		"data":   logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (s *server) handleExportLogs(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	rows, err := s.db.QueryContext(r.Context(), `
		SELECT id,user_id,provider,model,prompt_id,latency_ms,tokens,cost,status,error,created_at
		FROM ai_logs WHERE user_id=$1 ORDER BY created_at DESC LIMIT 10000
	`, userID)
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

	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=ai-logs.csv")
		w.Write([]byte("id,user_id,provider,model,latency_ms,tokens,cost,status,error,created_at\n"))
		for _, l := range logs {
			line := fmt.Sprintf("%s,%s,%s,%s,%d,%d,%.4f,%s,%s,%s\n", l.ID, l.UserID, l.Provider, l.Model, l.LatencyMs, l.Tokens, l.Cost, l.Status, l.Error, l.CreatedAt.Format(time.RFC3339))
			w.Write([]byte(line))
		}
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func (s *server) handleHubSettings(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	switch r.Method {
	case "GET":
		type Setting struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		rows, err := s.db.QueryContext(r.Context(), `SELECT setting_key, setting_value FROM ai_settings WHERE user_id=$1`, userID)
		settings := map[string]string{
			"default_provider": "openai",
			"retry_count":      "3",
			"timeout":          "60",
			"streaming":        "true",
			"budget_monthly":   "0",
			"budget_alert_at":  "0",
		}
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var k, v string
				rows.Scan(&k, &v)
				settings[k] = v
			}
		}
		var defaultProvider string
		s.db.QueryRowContext(r.Context(), `SELECT provider FROM ai_provider_keys WHERE user_id=$1 AND is_primary=true`, userID).Scan(&defaultProvider)
		if defaultProvider != "" {
			settings["default_provider"] = defaultProvider
		}
		writeJSON(w, http.StatusOK, settings)
	case "PUT":
		var incoming map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&incoming); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
		for k, v := range incoming {
			val := fmt.Sprintf("%v", v)
			s.db.ExecContext(r.Context(), `
				INSERT INTO ai_settings (user_id, setting_key, setting_value)
				VALUES ($1, $2, $3)
				ON CONFLICT (user_id, setting_key) DO UPDATE SET setting_value = $3, updated_at = NOW()
			`, userID, k, val)
		}
		writeJSON(w, http.StatusOK, map[string]string{"message": "settings saved"})
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
