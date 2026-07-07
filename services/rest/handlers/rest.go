package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/novabase/novabase/services/rest/db"
	"github.com/novabase/novabase/services/rest/query"
)

type RESTHandler struct {
	DB       *sql.DB
	Registry *db.Registry
}

func NewRESTHandler(database *sql.DB, reg *db.Registry) *RESTHandler {
	return &RESTHandler{
		DB:       database,
		Registry: reg,
	}
}

// Get handles SELECT operations with dynamic filtering/sorting/pagination.
func (h *RESTHandler) Get(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "table")
	if !h.Registry.IsValidTable(tableName) {
		h.respondError(w, http.StatusNotFound, "Table not found", fmt.Sprintf("Table '%s' does not exist in schema", tableName))
		return
	}

	tableSchema, _ := h.Registry.GetTableSchema(tableName)
	qResult, err := query.BuildSelect(tableName, tableSchema, r.URL.Query())
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid Query Parameters", err.Error())
		return
	}

	rows, err := h.DB.Query(qResult.SQL, qResult.Args...)
	if err != nil {
		slog.Error("Database SELECT query failed", "sql", qResult.SQL, "error", err)
		h.respondError(w, http.StatusInternalServerError, "Database Query Error", err.Error())
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

// Post handles single/bulk INSERT operations.
func (h *RESTHandler) Post(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "table")
	if !h.Registry.IsValidTable(tableName) {
		h.respondError(w, http.StatusNotFound, "Table not found", fmt.Sprintf("Table '%s' does not exist in schema", tableName))
		return
	}

	tableSchema, _ := h.Registry.GetTableSchema(tableName)

	// Decode body dynamically. Could be single map or array of maps
	var rawBody json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawBody); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON Request", err.Error())
		return
	}

	var payloads []map[string]interface{}
	if rawBody[0] == '[' {
		// Bulk insert payload
		if err := json.Unmarshal(rawBody, &payloads); err != nil {
			h.respondError(w, http.StatusBadRequest, "Malformed bulk payload", err.Error())
			return
		}
	} else {
		// Single insert payload
		var single map[string]interface{}
		if err := json.Unmarshal(rawBody, &single); err != nil {
			h.respondError(w, http.StatusBadRequest, "Malformed single payload", err.Error())
			return
		}
		payloads = append(payloads, single)
	}

	if len(payloads) == 0 {
		h.respondError(w, http.StatusBadRequest, "Empty Payload", "At least one row must be provided for insertion")
		return
	}

	// Begin atomic transaction to secure integrity of the writes
	tx, err := h.DB.Begin()
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Transaction Initiation Failed", err.Error())
		return
	}
	defer tx.Rollback()

	insertedRows := []map[string]interface{}{}

	// Validate columns and build insert statements for each payload
	for _, payload := range payloads {
		var cols []string
		var valPlaceholders []string
		var args []interface{}
		argCounter := 1

		for k, v := range payload {
			if _, exists := tableSchema.Columns[k]; !exists {
				continue // Skip non-existent columns to avoid errors
			}
			cols = append(cols, fmt.Sprintf(`"%s"`, k))
			valPlaceholders = append(valPlaceholders, fmt.Sprintf("$%d", argCounter))
			args = append(args, v)
			argCounter++
		}

		if len(cols) == 0 {
			continue
		}

		insertSQL := fmt.Sprintf(
			`INSERT INTO "%s" (%s) VALUES (%s) RETURNING *`,
			tableName,
			strings.Join(cols, ", "),
			strings.Join(valPlaceholders, ", "),
		)

		singleRow, err := h.insertAndScan(tx, insertSQL, args)
		if err != nil {
			h.respondError(w, http.StatusInternalServerError, "Failed to scan inserted row", err.Error())
			return
		}
		insertedRows = append(insertedRows, singleRow)
	}

	if err := tx.Commit(); err != nil {
		h.respondError(w, http.StatusInternalServerError, "Transaction Commit Failed", err.Error())
		return
	}

	if len(insertedRows) == 1 {
		h.respondJSON(w, http.StatusCreated, insertedRows[0])
	} else {
		h.respondJSON(w, http.StatusCreated, insertedRows)
	}
}

// Patch handles UPDATE operations with filters.
func (h *RESTHandler) Patch(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "table")
	if !h.Registry.IsValidTable(tableName) {
		h.respondError(w, http.StatusNotFound, "Table not found", fmt.Sprintf("Table '%s' does not exist in schema", tableName))
		return
	}

	tableSchema, _ := h.Registry.GetTableSchema(tableName)

	var updateData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		h.respondError(w, http.StatusBadRequest, "Malformed JSON Body", err.Error())
		return
	}

	qResult, err := query.BuildUpdate(tableName, tableSchema, r.URL.Query(), updateData)
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid Update Request", err.Error())
		return
	}

	res, err := h.DB.Exec(qResult.SQL, qResult.Args...)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Update execution failed", err.Error())
		return
	}

	rowsAffected, _ := res.RowsAffected()
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"rows_affected": rowsAffected,
		"message":       "Records successfully updated",
	})
}

// Delete handles DELETE operations with filters.
func (h *RESTHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tableName := chi.URLParam(r, "table")
	if !h.Registry.IsValidTable(tableName) {
		h.respondError(w, http.StatusNotFound, "Table not found", fmt.Sprintf("Table '%s' does not exist in schema", tableName))
		return
	}

	tableSchema, _ := h.Registry.GetTableSchema(tableName)

	qResult, err := query.BuildDelete(tableName, tableSchema, r.URL.Query())
	if err != nil {
		h.respondError(w, http.StatusBadRequest, "Invalid Delete Request", err.Error())
		return
	}

	res, err := h.DB.Exec(qResult.SQL, qResult.Args...)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "Delete execution failed", err.Error())
		return
	}

	rowsAffected, _ := res.RowsAffected()
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"rows_affected": rowsAffected,
		"message":       "Records successfully deleted",
	})
}

// scanRows converts sql.Rows dynamically to slice of maps
func (h *RESTHandler) scanRows(rows *sql.Rows) ([]map[string]interface{}, error) {
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
				// Safely parse bytes as string
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

// scanRowFromRow scans a single SQL returning row into a map dynamically using schema columns.
func (h *RESTHandler) scanRowFromRow(row *sql.Row, schema db.Table) (map[string]interface{}, error) {
	// Query to retrieve columns directly to match Scan order
	// Create order array of columns
	var cols []string
	for name := range schema.Columns {
		cols = append(cols, name)
	}

	columns := make([]interface{}, len(cols))
	columnPointers := make([]interface{}, len(cols))
	for i := range columns {
		columnPointers[i] = &columns[i]
	}

	// Warning: sql.Row doesn't expose column mappings, we must execute a select or inspect order.
	// However, we can run Query instead of QueryRow to retrieve columns dynamically and get exact scan order!
	// Yes! Using Query for returning rows is much safer than Row because it dynamically maps columns.
	return nil, fmt.Errorf("use scanRows helper with Query instead")
}

// Safe Insert returning mapper using query execution.
func (h *RESTHandler) insertAndScan(tx *sql.Tx, insertSQL string, args []interface{}) (map[string]interface{}, error) {
	rows, err := tx.Query(insertSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results, err := h.scanRows(rows)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no row returned")
	}
	return results[0], nil
}

func (h *RESTHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *RESTHandler) respondError(w http.ResponseWriter, status int, errType, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   errType,
		"message": msg,
	})
}
