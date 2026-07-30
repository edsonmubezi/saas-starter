package pagination

import (
	"net/http"
	"strconv"
	"strings"
)

type Pagination struct {
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
	Search    string `json:"search,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"` // "asc" or "desc"
}

// MaxPageSize is the maximum allowed page size to prevent memory exhaustion
const MaxPageSize = 100

func Parse(r *http.Request) Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	// Support both "limit" and "page_size" query params for flexibility
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit, _ = strconv.Atoi(r.URL.Query().Get("page_size"))
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	sortOrder := strings.ToLower(r.URL.Query().Get("sort_order"))
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "" // leave empty so developer-set defaults take effect
	}

	// Support both "search" and "q" query params for flexibility
	search := r.URL.Query().Get("search")
	if search == "" {
		search = r.URL.Query().Get("q")
	}

	return Pagination{
		Page:      page,
		PageSize:  limit,
		Search:    search,
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: sortOrder,
	}
}

func (p Pagination) Offset() int {
	return (p.Page - 1) * p.PageSize
}

func FromValues(page, pageSize int) Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return Pagination{
		Page:     page,
		PageSize: pageSize,
	}
}
