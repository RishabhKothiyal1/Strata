package db

import (
	"database/sql"
	"log/slog"
	"sync"
)

type Column struct {
	Name       string
	Type       string
	IsNullable bool
}

type ForeignKey struct {
	SourceColumn string
	TargetTable  string
	TargetColumn string
}

type Table struct {
	Name        string
	Columns     map[string]Column
	ForeignKeys map[string]ForeignKey // maps target_table -> ForeignKey detail
}

type Registry struct {
	mu     sync.RWMutex
	Tables map[string]Table
}

func NewRegistry() *Registry {
	return &Registry{
		Tables: make(map[string]Table),
	}
}

func (r *Registry) Introspect(db *sql.DB) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	slog.Info("Introspecting database schema...")

	// 1. Fetch all tables
	tableQuery := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE';
	`
	rows, err := db.Query(tableQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	tables := make(map[string]Table)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}
		tables[name] = Table{
			Name:        name,
			Columns:     make(map[string]Column),
			ForeignKeys: make(map[string]ForeignKey),
		}
	}

	// 2. Fetch all columns
	columnQuery := `
		SELECT table_name, column_name, data_type, is_nullable 
		FROM information_schema.columns 
		WHERE table_schema = 'public';
	`
	colRows, err := db.Query(columnQuery)
	if err != nil {
		return err
	}
	defer colRows.Close()

	for colRows.Next() {
		var tableName, columnName, dataType, isNullable string
		if err := colRows.Scan(&tableName, &columnName, &dataType, &isNullable); err != nil {
			return err
		}

		if t, exists := tables[tableName]; exists {
			t.Columns[columnName] = Column{
				Name:       columnName,
				Type:       dataType,
				IsNullable: isNullable == "YES",
			}
		}
	}

	// 3. Fetch all foreign keys
	fkQuery := `
		SELECT
			tc.table_name AS source_table,
			kcu.column_name AS source_column,
			ccu.table_name AS target_table,
			ccu.column_name AS target_column
		FROM
			information_schema.table_constraints AS tc
			JOIN information_schema.key_column_usage AS kcu 
				ON tc.constraint_name = kcu.constraint_name AND tc.table_schema = kcu.table_schema
			JOIN information_schema.constraint_column_usage AS ccu 
				ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_schema = 'public';
	`
	fkRows, err := db.Query(fkQuery)
	if err != nil {
		return err
	}
	defer fkRows.Close()

	for fkRows.Next() {
		var sourceTable, sourceColumn, targetTable, targetColumn string
		if err := fkRows.Scan(&sourceTable, &sourceColumn, &targetTable, &targetColumn); err != nil {
			return err
		}

		if t, exists := tables[sourceTable]; exists {
			t.ForeignKeys[targetTable] = ForeignKey{
				SourceColumn: sourceColumn,
				TargetTable:  targetTable,
				TargetColumn: targetColumn,
			}
		}
	}

	r.Tables = tables
	slog.Info("Introspection complete. Loaded schemas for tables.", "count", len(r.Tables))
	return nil
}

func (r *Registry) IsValidTable(table string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.Tables[table]
	return exists
}

func (r *Registry) IsValidColumn(table, column string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, exists := r.Tables[table]
	if !exists {
		return false
	}
	_, exists = t.Columns[column]
	return exists
}

func (r *Registry) GetTableSchema(table string) (Table, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, exists := r.Tables[table]
	return t, exists
}
