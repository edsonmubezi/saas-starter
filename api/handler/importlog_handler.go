package handler

import (
	"net/http"
	"strconv"

	"github.com/edsonmubezi/myapp/internal/identity"
	"github.com/edsonmubezi/myapp/internal/importlog"
	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
)

var importLogUseCase importlog.UseCase

// SetImportLogUseCase sets the import log use case.
func SetImportLogUseCase(uc importlog.UseCase) {
	importLogUseCase = uc
}

// GetImportSessionsHandler returns paginated import sessions for the org.
// GET /api/orgs/import-logs?page=1&page_size=20&import_type=<your_type>
func GetImportSessionsHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	page := 1
	pageSize := 20
	if pg, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil && pg > 0 {
		page = pg
	}
	if ps, err := strconv.Atoi(r.URL.Query().Get("page_size")); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	importType := r.URL.Query().Get("import_type")

	sessions, total, err := importLogUseCase.GetImportHistory(r.Context(), orgID, importType, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch import history", nil)
		return
	}

	if sessions == nil {
		sessions = []importlog.ImportSession{}
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Import history retrieved",
		"data": map[string]any{
			"sessions":    sessions,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + pageSize - 1) / pageSize,
		},
	})
}

// GetImportFailedRecordsHandler returns failed records for a specific import session.
// GET /api/orgs/import-logs/{id}/records
func GetImportFailedRecordsHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := mux.Vars(r)["id"]

	// Try encrypted ID first, fall back to plain int
	var sessionID int64
	if decrypted, err := secureid.DecryptID(idStr); err == nil {
		sessionID, _ = strconv.ParseInt(decrypted, 10, 64)
	}
	if sessionID == 0 {
		sessionID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			SendJSONResponse(w, http.StatusBadRequest, "Invalid session ID", nil)
			return
		}
	}

	records, err := importLogUseCase.GetSessionFailedRecords(r.Context(), sessionID, orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch failed records", nil)
		return
	}

	if records == nil {
		records = []importlog.ImportFailedRecord{}
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Failed records retrieved",
		"data":    records,
	})
}

// GetDuplicateRecordsHandler returns unresolved duplicate records for a specific import session.
// GET /api/orgs/import-logs/{id}/duplicates
func GetDuplicateRecordsHandler(w http.ResponseWriter, r *http.Request) {
	orgID, err := identity.OrgID(r.Context())
	if err != nil {
		SendJSONResponse(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	idStr := mux.Vars(r)["id"]

	var sessionID int64
	if decrypted, err := secureid.DecryptID(idStr); err == nil {
		sessionID, _ = strconv.ParseInt(decrypted, 10, 64)
	}
	if sessionID == 0 {
		sessionID, err = strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			SendJSONResponse(w, http.StatusBadRequest, "Invalid session ID", nil)
			return
		}
	}

	records, err := importLogUseCase.GetDuplicateRecords(r.Context(), sessionID, orgID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to fetch duplicate records", nil)
		return
	}

	if records == nil {
		records = []importlog.ImportFailedRecord{}
	}

	securejson.JSON(w, http.StatusOK, map[string]any{
		"status":  http.StatusOK,
		"message": "Duplicate records retrieved",
		"data":    records,
	})
}
