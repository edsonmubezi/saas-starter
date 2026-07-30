package pagination

import (
	"log"
	"net/http"
)

// LogRequest logs basic pagination + search + sort info
func LogRequest(r *http.Request, pg Pagination) {
	log.Printf(
		"[Pagination] Path: %s | Page: %d | PageSize: %d | Search: %q | SortBy: %q %s",
		r.URL.Path, pg.Page, pg.PageSize, pg.Search, pg.SortBy, pg.SortOrder,
	)
}
