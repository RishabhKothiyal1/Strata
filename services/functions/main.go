package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dop251/goja"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


// ---------------------------------------------------------------------------
// Domain types
// ---------------------------------------------------------------------------

type Function struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Code        string    `json:"code"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type deployRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Code        string `json:"code"`
}

type updateRequest struct {
	Description string `json:"description"`
	Code        string `json:"code"`
}

type invokeRequest struct {
	Body    json.RawMessage   `json:"body"`
	Headers map[string]string `json:"headers"`
}

type invokeResponse struct {
	StatusCode int         `json:"status_code"`
	Body       interface{} `json:"body"`
	Duration   string      `json:"duration_ms"`
}

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

type server struct {
	db *sql.DB
}

func main() {
	slog.Info("Starting Strata Functions Service...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://strata_admin:strata_secure_pass_123@strata-postgres:5432/strata?sslmode=disable"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = ":8085"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Failed to open database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	// Wait for PG to be ready
	for i := range 30 {
		if err := db.Ping(); err == nil {
			break
		}
		slog.Info("Waiting for PostgreSQL...", "attempt", i+1)
		time.Sleep(2 * time.Second)
	}
	if err := db.Ping(); err != nil {
		slog.Error("PostgreSQL unavailable after retries", "error", err)
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

	r.Get("/v1/functions/health", s.handleHealth)
	r.Get("/v1/functions", s.handleList)
	r.Post("/v1/functions", s.handleDeploy)
	r.Get("/v1/functions/{name}", s.handleGet)
	r.Put("/v1/functions/{name}", s.handleUpdate)
	r.Delete("/v1/functions/{name}", s.handleDelete)
	r.Post("/v1/functions/{name}/invoke", s.handleInvoke)

	srv := &http.Server{
		Addr:         port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("Functions service listening", "addr", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	slog.Info("Functions service shut down cleanly.")
}

// ---------------------------------------------------------------------------
// Migration
// ---------------------------------------------------------------------------

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS strata_functions (
			id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name        VARCHAR(255) UNIQUE NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			code        TEXT NOT NULL,
			created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "functions-go"})
}

func (s *server) handleList(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.QueryContext(r.Context(),
		`SELECT id, name, description, code, created_at, updated_at FROM strata_functions ORDER BY created_at DESC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}
	defer rows.Close()

	fns := []Function{}
	for rows.Next() {
		var f Function
		if err := rows.Scan(&f.ID, &f.Name, &f.Description, &f.Code, &f.CreatedAt, &f.UpdatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "scan failed: "+err.Error())
			return
		}
		fns = append(fns, f)
	}
	writeJSON(w, http.StatusOK, fns)
}

func (s *server) handleDeploy(w http.ResponseWriter, r *http.Request) {
	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Validate the JS compiles before storing
	if _, err := compileJS(req.Code); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "JavaScript compile error: "+err.Error())
		return
	}

	id := uuid.New().String()
	now := time.Now().UTC()
	_, err := s.db.ExecContext(r.Context(),
		`INSERT INTO strata_functions (id, name, description, code, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		id, req.Name, req.Description, req.Code, now, now)
	if err != nil {
		writeError(w, http.StatusConflict, "deploy failed: "+err.Error())
		return
	}

	slog.Info("Function deployed", "name", req.Name)
	writeJSON(w, http.StatusCreated, Function{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Code:        req.Code,
		CreatedAt:   now,
		UpdatedAt:   now,
	})
}

func (s *server) handleGet(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	f, err := s.getByName(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	if _, err := compileJS(req.Code); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "JavaScript compile error: "+err.Error())
		return
	}

	res, err := s.db.ExecContext(r.Context(),
		`UPDATE strata_functions SET code=$1, description=$2, updated_at=NOW() WHERE name=$3`,
		req.Code, req.Description, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update failed: "+err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}

	slog.Info("Function updated", "name", name)
	f, _ := s.getByName(r.Context(), name)
	writeJSON(w, http.StatusOK, f)
}

func (s *server) handleDelete(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	res, err := s.db.ExecContext(r.Context(),
		`DELETE FROM strata_functions WHERE name=$1`, name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete failed: "+err.Error())
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}
	slog.Info("Function deleted", "name", name)
	writeJSON(w, http.StatusOK, map[string]string{"message": "function deleted"})
}

func (s *server) handleInvoke(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")

	fn, err := s.getByName(r.Context(), name)
	if err == sql.ErrNoRows {
		writeError(w, http.StatusNotFound, "function not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Parse optional request body
	var req invokeRequest
	if r.ContentLength != 0 {
		json.NewDecoder(r.Body).Decode(&req)
	}

	// Build request object for JS
	var bodyObj interface{}
	if len(req.Body) > 0 {
		json.Unmarshal(req.Body, &bodyObj)
	}
	headers := map[string]string{}
	for k, vals := range r.Header {
		if len(vals) > 0 {
			headers[k] = vals[0]
		}
	}

	start := time.Now()
	result, statusCode, invokeErr := invokeFunction(fn.Code, bodyObj, headers, r.Method)
	elapsed := time.Since(start)

	slog.Info("Function invoked", "name", name, "status", statusCode, "duration_ms", elapsed.Milliseconds())

	if invokeErr != nil {
		writeError(w, http.StatusInternalServerError, "execution error: "+invokeErr.Error())
		return
	}

	writeJSON(w, statusCode, invokeResponse{
		StatusCode: statusCode,
		Body:       result,
		Duration:   elapsed.String(),
	})
}

// ---------------------------------------------------------------------------
// JavaScript execution engine
// ---------------------------------------------------------------------------

func compileJS(code string) (*goja.Program, error) {
	return goja.Compile("function.js", code, false)
}

func invokeFunction(code string, body interface{}, headers map[string]string, method string) (interface{}, int, error) {
	vm := goja.New()

	// Wire console.log to slog
	console := vm.NewObject()
	console.Set("log", func(call goja.FunctionCall) goja.Value {
		args := make([]interface{}, len(call.Arguments))
		for i, a := range call.Arguments {
			args[i] = a.Export()
		}
		slog.Info("function console.log", "args", args)
		return goja.Undefined()
	})
	vm.Set("console", console)

	// 10-second hard timeout
	timer := time.AfterFunc(10*time.Second, func() {
		vm.Interrupt("execution timeout exceeded (10s)")
	})
	defer timer.Stop()

	// Run the function code (defines handler in the VM)
	prog, err := compileJS(code)
	if err != nil {
		return nil, http.StatusUnprocessableEntity, err
	}
	if _, err := vm.RunProgram(prog); err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Call handler(request)
	handlerFn, ok := goja.AssertFunction(vm.Get("handler"))
	if !ok {
		return nil, http.StatusUnprocessableEntity, fmt.Errorf("function must export a `handler` function")
	}

	reqObj := vm.NewObject()
	reqObj.Set("body", vm.ToValue(body))
	reqObj.Set("headers", vm.ToValue(headers))
	reqObj.Set("method", vm.ToValue(method))

	val, err := handlerFn(goja.Undefined(), vm.ToValue(reqObj))
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Expect { statusCode: int, body: any }
	result := val.Export()
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return result, http.StatusOK, nil
	}

	statusCode := http.StatusOK
	if sc, ok := resultMap["statusCode"]; ok {
		switch v := sc.(type) {
		case int64:
			statusCode = int(v)
		case float64:
			statusCode = int(v)
		case int:
			statusCode = v
		}
	}

	responseBody := resultMap["body"]
	if responseBody == nil {
		responseBody = resultMap
	}
	return responseBody, statusCode, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func (s *server) getByName(ctx context.Context, name string) (*Function, error) {
	var f Function
	err := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, code, created_at, updated_at FROM strata_functions WHERE name=$1`, name).
		Scan(&f.ID, &f.Name, &f.Description, &f.Code, &f.CreatedAt, &f.UpdatedAt)
	return &f, err
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
