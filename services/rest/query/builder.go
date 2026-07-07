package query

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/strata/strata/services/rest/db"
)

type BuildResult struct {
	SQL  string
	Args []interface{}
}

// BuildSelect constructs a safe SELECT query with parameterized filters, ordering, and pagination.
func BuildSelect(tableName string, tableSchema db.Table, queryParams url.Values) (*BuildResult, error) {
	var (
		columns      = "*"
		whereClauses = []string{}
		args         = []interface{}{}
		argCounter   = 1
		orderBy      = ""
		limitStr     = ""
		offsetStr    = ""
	)

	// 1. Parse 'select' projection
	if sel := queryParams.Get("select"); sel != "" && sel != "*" {
		parts := strings.Split(sel, ",")
		var validCols []string
		for _, col := range parts {
			col = strings.TrimSpace(col)
			// Check if column is valid in introspection schema
			if _, exists := tableSchema.Columns[col]; exists {
				validCols = append(validCols, fmt.Sprintf(`"%s"`, col))
			}
		}
		if len(validCols) > 0 {
			columns = strings.Join(validCols, ", ")
		}
	}

	// 2. Parse general column filters (e.g. name=eq.Alice or age=gt.21)
	for key, values := range queryParams {
		if key == "select" || key == "order" || key == "limit" || key == "offset" {
			continue
		}

		// Only filter on valid columns
		if _, exists := tableSchema.Columns[key]; !exists {
			continue
		}

		for _, val := range values {
			parts := strings.SplitN(val, ".", 2)
			if len(parts) < 2 {
				// Fallback to equality filter if no operator specified
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, key, argCounter))
				args = append(args, val)
				argCounter++
				continue
			}

			op := parts[0]
			filterVal := parts[1]

			switch op {
			case "eq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "neq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" != $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "gt":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" > $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "gte":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" >= $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "lt":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" < $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "lte":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" <= $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "like":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" LIKE $%d`, key, argCounter))
				// Translate * wildcard to % for SQL LIKE operations
				args = append(args, strings.ReplaceAll(filterVal, "*", "%"))
				argCounter++
			case "is":
				if filterVal == "null" {
					whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NULL`, key))
				} else if filterVal == "notnull" {
					whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IS NOT NULL`, key))
				}
			case "in":
				// Example format: in.(1,2,3)
				if strings.HasPrefix(filterVal, "(") && strings.HasSuffix(filterVal, ")") {
					itemsStr := filterVal[1 : len(filterVal)-1]
					items := strings.Split(itemsStr, ",")
					var placeHolders []string
					for _, item := range items {
						placeHolders = append(placeHolders, fmt.Sprintf("$%d", argCounter))
						args = append(args, strings.TrimSpace(item))
						argCounter++
					}
					if len(placeHolders) > 0 {
						whereClauses = append(whereClauses, fmt.Sprintf(`"%s" IN (%s)`, key, strings.Join(placeHolders, ", ")))
					}
				}
			}
		}
	}

	// 3. Parse 'order' sorting parameter
	if order := queryParams.Get("order"); order != "" {
		parts := strings.Split(order, ".")
		col := parts[0]
		if _, exists := tableSchema.Columns[col]; exists {
			direction := "ASC"
			if len(parts) > 1 && strings.ToLower(parts[1]) == "desc" {
				direction = "DESC"
			}
			orderBy = fmt.Sprintf(` ORDER BY "%s" %s`, col, direction)
		}
	}

	// 4. Parse 'limit' pagination parameter
	if limit := queryParams.Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil && val >= 0 {
			limitStr = fmt.Sprintf(" LIMIT %d", val)
		}
	}

	// 5. Parse 'offset' pagination parameter
	if offset := queryParams.Get("offset"); offset != "" {
		if val, err := strconv.Atoi(offset); err == nil && val >= 0 {
			offsetStr = fmt.Sprintf(" OFFSET %d", val)
		}
	}

	// Assemble final SQL query
	sqlBuilder := strings.Builder{}
	sqlBuilder.WriteString(fmt.Sprintf(`SELECT %s FROM "%s"`, columns, tableName))

	if len(whereClauses) > 0 {
		sqlBuilder.WriteString(" WHERE ")
		sqlBuilder.WriteString(strings.Join(whereClauses, " AND "))
	}

	sqlBuilder.WriteString(orderBy)
	sqlBuilder.WriteString(limitStr)
	sqlBuilder.WriteString(offsetStr)

	return &BuildResult{
		SQL:  sqlBuilder.String(),
		Args: args,
	}, nil
}

// BuildUpdate constructs a parameterized UPDATE query based on filters.
func BuildUpdate(tableName string, tableSchema db.Table, queryParams url.Values, updateData map[string]interface{}) (*BuildResult, error) {
	if len(updateData) == 0 {
		return nil, fmt.Errorf("no update fields provided")
	}

	var (
		setClauses   = []string{}
		whereClauses = []string{}
		args         = []interface{}{}
		argCounter   = 1
	)

	// 1. Populate Set fields
	for k, v := range updateData {
		if _, exists := tableSchema.Columns[k]; !exists {
			continue
		}
		setClauses = append(setClauses, fmt.Sprintf(`"%s" = $%d`, k, argCounter))
		args = append(args, v)
		argCounter++
	}

	if len(setClauses) == 0 {
		return nil, fmt.Errorf("no valid columns to update")
	}

	// 2. Parse where filters from query params
	for key, values := range queryParams {
		if key == "select" || key == "order" || key == "limit" || key == "offset" {
			continue
		}

		if _, exists := tableSchema.Columns[key]; !exists {
			continue
		}

		for _, val := range values {
			parts := strings.SplitN(val, ".", 2)
			op := "eq"
			filterVal := val
			if len(parts) == 2 {
				op = parts[0]
				filterVal = parts[1]
			}

			switch op {
			case "eq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "neq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" != $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			}
		}
	}

	if len(whereClauses) == 0 {
		return nil, fmt.Errorf("updating all records without filters is restricted for safety")
	}

	sql := fmt.Sprintf(`UPDATE "%s" SET %s WHERE %s`, tableName, strings.Join(setClauses, ", "), strings.Join(whereClauses, " AND "))
	return &BuildResult{
		SQL:  sql,
		Args: args,
	}, nil
}

// BuildDelete constructs a parameterized DELETE query based on filters.
func BuildDelete(tableName string, tableSchema db.Table, queryParams url.Values) (*BuildResult, error) {
	var (
		whereClauses = []string{}
		args         = []interface{}{}
		argCounter   = 1
	)

	// Parse where filters
	for key, values := range queryParams {
		if key == "select" || key == "order" || key == "limit" || key == "offset" {
			continue
		}

		if _, exists := tableSchema.Columns[key]; !exists {
			continue
		}

		for _, val := range values {
			parts := strings.SplitN(val, ".", 2)
			op := "eq"
			filterVal := val
			if len(parts) == 2 {
				op = parts[0]
				filterVal = parts[1]
			}

			switch op {
			case "eq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" = $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			case "neq":
				whereClauses = append(whereClauses, fmt.Sprintf(`"%s" != $%d`, key, argCounter))
				args = append(args, filterVal)
				argCounter++
			}
		}
	}

	if len(whereClauses) == 0 {
		return nil, fmt.Errorf("deleting all records without filters is restricted for safety")
	}

	sql := fmt.Sprintf(`DELETE FROM "%s" WHERE %s`, tableName, strings.Join(whereClauses, " AND "))
	return &BuildResult{
		SQL:  sql,
		Args: args,
	}, nil
}
