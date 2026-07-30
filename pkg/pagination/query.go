package pagination

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type Options[T any] struct {
	Table      string
	Fields     []string
	SearchCols []string
	Pagination Pagination
	SortBy     string
	ScanFunc   func(pgx.Row) (T, error)
	Args       []any // <-- Add this line
}

func QueryWithPagination[T any](ctx context.Context, db *pgxpool.Pool, opts Options[T]) (Result[T], error) {
	// Validate all field names to prevent SQL injection
	for _, field := range opts.Fields {
		if !isValidSQLIdentifier(field) {
			return Result[T]{}, fmt.Errorf("invalid field name: %s", field)
		}
	}

	// Validate search columns
	for _, col := range opts.SearchCols {
		if !isValidSQLIdentifier(col) {
			return Result[T]{}, fmt.Errorf("invalid search column name: %s", col)
		}
	}

	// Validate sort by field
	if opts.Pagination.SortBy != "" && !isValidSQLIdentifier(opts.Pagination.SortBy) {
		return Result[T]{}, fmt.Errorf("invalid sort by field: %s", opts.Pagination.SortBy)
	}

	args := []interface{}{}
	where := "WHERE 1=1"

	if opts.Pagination.Search != "" {
		filter, searchArgs := BuildSearchFilter(opts.SearchCols, 1, opts.Pagination.Search)
		where += " AND " + filter
		args = append(args, searchArgs...)
	}

	offset := opts.Pagination.Offset()
	args = append(args, opts.Pagination.PageSize, offset)

	orderBy := "id"
	sortOrder := "ASC"

	// Developer-set default sort (may include direction, e.g. "date_created DESC")
	// Supports single-column ("date_created DESC") and multi-column ("col1 DESC, col2 ASC")
	if opts.SortBy != "" {
		if strings.Contains(opts.SortBy, ",") {
			// Multi-column sort expression — use as-is (developer-set, trusted)
			orderBy = opts.SortBy
			sortOrder = "" // direction already embedded
		} else {
			parts := strings.Fields(opts.SortBy)
			candidate := parts[0]
			if contains(opts.Fields, candidate) {
				orderBy = candidate
				if len(parts) > 1 {
					dir := strings.ToUpper(parts[1])
					if dir == "DESC" || dir == "ASC" {
						sortOrder = dir
					}
				}
			}
		}
	}

	// URL-provided sort overrides developer default
	if opts.Pagination.SortBy != "" && contains(opts.Fields, opts.Pagination.SortBy) {
		orderBy = opts.Pagination.SortBy
		sortOrder = "ASC" // reset direction for user-chosen field
	}
	if opts.Pagination.SortOrder == "desc" {
		sortOrder = "DESC"
	} else if opts.Pagination.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	// Build ORDER BY clause
	orderClause := orderBy
	if sortOrder != "" {
		orderClause = orderBy + " " + sortOrder
	}

	// Final query (fields and table are now validated)
	query := fmt.Sprintf(`
		SELECT %s FROM %s
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, strings.Join(opts.Fields, ","), opts.Table, where, orderClause, len(args)-1, len(args))

	rows, err := db.Query(ctx, query, args...)
	if err != nil {
		return Result[T]{}, err
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		item, err := opts.ScanFunc(rows)
		if err != nil {
			return Result[T]{}, err
		}
		results = append(results, item)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s", opts.Table, where)
	err = db.QueryRow(ctx, countQuery, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return Result[T]{}, err
	}

	totalPages := 0
	if opts.Pagination.PageSize > 0 {
		totalPages = (total + opts.Pagination.PageSize - 1) / opts.Pagination.PageSize
	}

	return Result[T]{
		Data:       results,
		Page:       opts.Pagination.Page,
		PageSize:   opts.Pagination.PageSize,
		TotalCount: total,
		TotalPages: totalPages,
		HasNext:    opts.Pagination.Page < totalPages,
		HasPrev:    opts.Pagination.Page > 1,
	}, nil
}

func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// isValidSQLIdentifier validates that a string is a safe SQL identifier
// (alphanumeric, underscore, and dots for qualified names)
func isValidSQLIdentifier(identifier string) bool {
	// Allow empty strings (will be handled by caller)
	if identifier == "" {
		return true
	}

	// Pattern: alphanumeric, underscores, and dots (for qualified names like "table.column")
	// Must start with a letter or underscore
	// Can contain: letters, numbers, underscores, dots
	pattern := `^[a-zA-Z_][a-zA-Z0-9_\.]*$`
	matched, _ := regexp.MatchString(pattern, identifier)
	return matched
}
