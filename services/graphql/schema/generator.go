package schema

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/graphql-go/graphql"

	"github.com/novabase/novabase/services/graphql/db"
)

// GenerateSchema dynamically builds a GraphQL schema from PG database introspection.
func GenerateSchema(dbConn *sql.DB, registry *db.Registry, pubsub *PubSub) (*graphql.Schema, error) {
	tableTypes := make(map[string]*graphql.Object)
	queryFields := graphql.Fields{}
	mutationFields := graphql.Fields{}

	// --------------------------------------------------------------------------
	// 1. Pass 1: Build Type Objects for all tables
	// --------------------------------------------------------------------------
	for tableName, tableSchema := range registry.Tables {
		fields := graphql.Fields{}

		for colName, col := range tableSchema.Columns {
			var gType graphql.Type
			switch col.Type {
			case "integer", "serial", "bigint":
				gType = graphql.Int
			case "numeric", "double precision", "real":
				gType = graphql.Float
			case "boolean":
				gType = graphql.Boolean
			default:
				gType = graphql.String
			}

			// Capture column detail inside closure safely
			fields[colName] = &graphql.Field{
				Type: gType,
			}
		}

		tableTypes[tableName] = graphql.NewObject(graphql.ObjectConfig{
			Name:        tableName,
			Description: fmt.Sprintf("Auto-generated GraphQL type for database table '%s'", tableName),
			Fields:      fields,
		})
	}

	// --------------------------------------------------------------------------
	// 2. Pass 2: Connect Foreign Key Relationships
	// --------------------------------------------------------------------------
	for tableName, tableSchema := range registry.Tables {
		tblType := tableTypes[tableName]

		for targetTable, fk := range tableSchema.ForeignKeys {
			targetType, exists := tableTypes[targetTable]
			if !exists {
				continue
			}

			sourceCol := fk.SourceColumn
			tTable := fk.TargetTable
			tCol := fk.TargetColumn

			tblType.AddFieldConfig(targetTable, &graphql.Field{
				Type: targetType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					// Extract parent field value
					parentMap, ok := p.Source.(map[string]interface{})
					if !ok {
						return nil, nil
					}

					fkValue, exists := parentMap[sourceCol]
					if !exists || fkValue == nil {
						return nil, nil
					}

					// Query the related table
					query := fmt.Sprintf(`SELECT * FROM "%s" WHERE "%s" = $1 LIMIT 1`, tTable, tCol)
					rows, err := queryRows(dbConn, query, []interface{}{fkValue})
					if err != nil {
						return nil, err
					}
					if len(rows) == 0 {
						return nil, nil
					}
					return rows[0], nil
				},
			})
		}
	}

	// --------------------------------------------------------------------------
	// 3. Build Root Queries & Mutations
	// --------------------------------------------------------------------------
	for tableName, tableSchema := range registry.Tables {
		tblType := tableTypes[tableName]
		tName := tableName // local copy for closure

		// Singular Fetch Query (e.g. profile(id: 1))
		queryFields[tName] = &graphql.Field{
			Type: tblType,
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{Type: graphql.Int},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, hasID := p.Args["id"]
				if !hasID {
					return nil, fmt.Errorf("id argument is required")
				}

				query := fmt.Sprintf(`SELECT * FROM "%s" WHERE "id" = $1 LIMIT 1`, tName)
				rows, err := queryRows(dbConn, query, []interface{}{id})
				if err != nil {
					return nil, err
				}
				if len(rows) == 0 {
					return nil, nil
				}
				return rows[0], nil
			},
		}

		// Plural List Query with Filtering, Sorting, & Pagination (e.g. profiles(limit: 10, offset: 0))
		listArgs := graphql.FieldConfigArgument{
			"limit":  &graphql.ArgumentConfig{Type: graphql.Int},
			"offset": &graphql.ArgumentConfig{Type: graphql.Int},
			"order":  &graphql.ArgumentConfig{Type: graphql.String},
		}
		// Add column filter fields dynamically
		for colName, col := range tableSchema.Columns {
			var gType graphql.Type
			switch col.Type {
			case "integer", "serial", "bigint":
				gType = graphql.Int
			case "boolean":
				gType = graphql.Boolean
			default:
				gType = graphql.String
			}
			listArgs[colName] = &graphql.ArgumentConfig{Type: gType}
		}

		queryFields[tName+"s"] = &graphql.Field{
			Type: graphql.NewList(tblType),
			Args: listArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var (
					whereClauses = []string{}
					args         = []interface{}{}
					argCounter   = 1
					limitStr     = ""
					offsetStr    = ""
					orderBy      = ""
				)

				// Parse limit/offset pagination arguments
				if limit, ok := p.Args["limit"].(int); ok {
					limitStr = fmt.Sprintf(" LIMIT %d", limit)
				}
				if offset, ok := p.Args["offset"].(int); ok {
					offsetStr = fmt.Sprintf(" OFFSET %d", offset)
				}
				if order, ok := p.Args["order"].(string); ok {
					parts := strings.Split(order, ".")
					col := parts[0]
					// check column exists in database schema to avoid injection
					if _, exists := registry.Tables[tName].Columns[col]; exists {
						direction := "ASC"
						if len(parts) > 1 && strings.ToUpper(parts[1]) == "DESC" {
							direction = "DESC"
						}
						orderBy = fmt.Sprintf(` ORDER BY "%s" %s`, col, direction)
					}
				}

				// Parse column filters
				for colName := range registry.Tables[tName].Columns {
					if filterVal, exists := p.Args[colName]; exists {
						whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, colName, argCounter))
						args = append(args, filterVal)
						argCounter++
					}
				}

				sqlBuilder := strings.Builder{}
				sqlBuilder.WriteString(fmt.Sprintf(`SELECT * FROM "%s"`, tName))
				if len(whereClauses) > 0 {
					sqlBuilder.WriteString(" WHERE ")
					sqlBuilder.WriteString(strings.Join(whereClauses, " AND "))
				}
				sqlBuilder.WriteString(orderBy)
				sqlBuilder.WriteString(limitStr)
				sqlBuilder.WriteString(offsetStr)

				return queryRows(dbConn, sqlBuilder.String(), args)
			},
		}

		// Create Mutation (e.g. create_profile)
		mutationArgs := graphql.FieldConfigArgument{}
		for colName, col := range tableSchema.Columns {
			if colName == "id" || colName == "created_at" {
				continue
			}
			var gType graphql.Type
			switch col.Type {
			case "integer", "bigint":
				gType = graphql.Int
			case "boolean":
				gType = graphql.Boolean
			default:
				gType = graphql.String
			}
			mutationArgs[colName] = &graphql.ArgumentConfig{Type: gType}
		}

		mutationFields["create_"+tName] = &graphql.Field{
			Type: tblType,
			Args: mutationArgs,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				var cols []string
				var placeHolders []string
				var args []interface{}
				argCounter := 1

				for colName := range registry.Tables[tName].Columns {
					if colName == "id" || colName == "created_at" {
						continue
					}
					if val, ok := p.Args[colName]; ok {
						cols = append(cols, fmt.Sprintf(`"%s"`, colName))
						placeHolders = append(placeHolders, fmt.Sprintf("$%d", argCounter))
						args = append(args, val)
						argCounter++
					}
				}

				if len(cols) == 0 {
					return nil, fmt.Errorf("no arguments provided for insert")
				}

				insertSQL := fmt.Sprintf(
					`INSERT INTO "%s" (%s) VALUES (%s) RETURNING *`,
					tName,
					strings.Join(cols, ", "),
					strings.Join(placeHolders, ", "),
				)

				rows, err := queryRows(dbConn, insertSQL, args)
				if err != nil {
					return nil, err
				}
				if len(rows) == 0 {
					return nil, fmt.Errorf("insert failed to return row")
				}

				insertedRow := rows[0]

				// Publish created event to NATS for dynamic websocket subscription streaming
				pubsub.Publish(tName+".created", insertedRow)

				return insertedRow, nil
			},
		}
	}

	// Dynamic Root Query & Mutation configuration
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Query",
		Fields: queryFields,
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name:   "Mutation",
		Fields: mutationFields,
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		return nil, err
	}

	return &schema, nil
}

// Database helper to execute queries and format output to map slice
func queryRows(dbConn *sql.DB, query string, args []interface{}) ([]map[string]interface{}, error) {
	slog.Debug("Executing database GraphQL query", "sql", query, "args", args)
	rows, err := dbConn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

	return results, nil
}
