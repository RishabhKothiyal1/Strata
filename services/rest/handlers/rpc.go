package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type RPCHandler struct {
	DB *sql.DB
}

func NewRPCHandler(database *sql.DB) *RPCHandler {
	return &RPCHandler{
		DB: database,
	}
}

// Execute runs a PostgreSQL stored function/procedure (RPC) dynamically.
func (h *RPCHandler) Execute(w http.ResponseWriter, r *http.Request) {
	functionName := chi.URLParam(r, "function")
	if functionName == "" {
		h.respondError(w, http.StatusBadRequest, "Invalid Request", "Function name cannot be empty")
		return
	}

	// Read JSON body argument keys and values
	var argsMap map[string]interface{}
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&argsMap); err != nil {
			h.respondError(w, http.StatusBadRequest, "Malformed JSON Body", err.Error())
			return
		}
	}

	var placeholders []string
	var args []interface{}
	argCounter := 1

	// Construct named arguments query e.g. SELECT * FROM func_name(arg1 := $1, arg2 := $2)
	for k, v := range argsMap {
		// Secure key parameter name to prevent SQL injection
		safeKey := h.sanitizeIdentifier(k)
		placeholders = append(placeholders, fmt.Sprintf(`"%s" := $%d`, safeKey, argCounter))
		args = append(args, v)
		argCounter++
	}

	rpcSQL := fmt.Sprintf(`SELECT * FROM "%s"(%s)`, h.sanitizeIdentifier(functionName), strings.Join(placeholders, ", "))

	slog.Info("Executing RPC stored function", "sql", rpcSQL, "args", args)

	rows, err := h.DB.Query(rpcSQL, args...)
	if err != nil {
		slog.Error("RPC Execution failed", "sql", rpcSQL, "error", err)
		h.respondError(w, http.StatusInternalServerError, "RPC Execution Error", err.Error())
		return
	}
	defer rows.Close()

	data, err := h.scanRows(rows)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Serialization Error", err.Error())
		return
	}

	h.respondJSON(w, http.StatusOK, data)
}

// sanitizeIdentifier secures function and argument names from basic SQL injection
func (h *RPCHandler) sanitizeIdentifier(s string) string {
	s = strings.ReplaceAll(s, `"`, "")
	s = strings.ReplaceAll(s, `;`, "")
	s = strings.ReplaceAll(s, `'`, "")
	return s
}

// scanRows converts sql.Rows dynamically to slice of maps
func (h *RPCHandler) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range cols {
			val := columns[i]
			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

func (h *RPCHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RPCHandler) respondError(w http.ResponseWriter, status int, errType, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errType,
		"message": msg,
	})
}
