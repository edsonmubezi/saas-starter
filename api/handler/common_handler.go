package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/pkg/customdatatype"
	"github.com/edsonmubezi/myapp/pkg/database"
	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	"github.com/edsonmubezi/myapp/pkg/pagination"

	"github.com/joho/godotenv"
)

func loadEnv() {
	err := godotenv.Load(".env") // this will load .env from root automatically
	if err != nil {
		log.Fatalf("%v", err)
	}
}

// @Summary      Health Check
// @Description  Checks the health status of the application and database connection. Returns healthy status if database is reachable.
// @Tags         System
// @Produce      json
// @Success      200 {object} map[string]string "Application is healthy"
// @Failure      503 {object} map[string]string "Service unavailable - database connection failed"
// @Router       /health [get]
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Check database connection
	if database.DB == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"reason": "database connection not initialized",
		})
		return
	}

	// Ping database to verify connection is alive
	if err := database.DB.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "unhealthy",
			"reason": "database ping failed",
		})
		return
	}

	// Everything is healthy
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}
// @Summary      Test ID Encryption
// @Description  Development/testing endpoint to encrypt an ID value. Returns both original and encrypted ID.
// @Tags         System
// @Produce      json
// @Param        id query string true "ID to encrypt" example("12345")
// @Success      200 {object} map[string]string "Encryption successful"
// @Failure      400 {object} map[string]string "Missing id parameter"
// @Failure      500 {object} map[string]string "Encryption failed"
// @Router       /api/test/encrypt [get]
func EncryptTestHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing `id` query parameter", http.StatusBadRequest)
		return
	}
	loadEnv()

	// secret := os.Getenv("SECRET_KEY")
	// fmt.Println("Loaded SECRET_KEY:", secret) // Debug output
	encryptedID, err := secureid.EncryptID(id)
	if err != nil {
		log.Printf("Encryption error: %v", err)
		http.Error(w, "Failed to encrypt ID", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"original_id":  id,
		"encrypted_id": encryptedID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type APIResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"` // omit if nil
}

func SendJSONResponse(w http.ResponseWriter, status int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := APIResponse{
		Status:  status,
		Message: message,
		Data:    data,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func SendValidationErrors(w http.ResponseWriter, validationErrors []middleware.FieldError) {
	SendJSONResponse(w, http.StatusBadRequest, "Validation failed", map[string]interface{}{
		"errors": validationErrors,
	})
}

func ConvertMapToFieldErrors(errMap map[string]string) []middleware.FieldError {
	var fieldErrors []middleware.FieldError
	for field, message := range errMap {
		fieldErrors = append(fieldErrors, middleware.FieldError{
			Field:   field,
			Message: message,
		})
	}
	return fieldErrors
}

func getLastPathSegment(r *http.Request) string {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(segments) == 0 {
		return ""
	}
	return segments[len(segments)-1]
}

func computeTax(amount float64) float64 {
	switch {
	case amount <= 270000:
		return 0
	case amount > 270000 && amount <= 520000:
		return 0.08 * (amount - 270000)
	case amount > 520000 && amount <= 760000:
		return 20000 + 0.20*(amount-520000)
	case amount > 760000 && amount <= 1000000:
		return 68000 + 0.25*(amount-760000)
	case amount > 1000000:
		return 128000 + 0.30*(amount-1000000)
	default:
		return 0
	}
}

func PaginationQeury(r *http.Request) pagination.Pagination {
	pag := pagination.Parse(r)
	if v := r.URL.Query().Get("sort_by"); v != "" {
		pag.SortBy = v
	}
	if v := r.URL.Query().Get("sort_order"); v != "" {
		pag.SortOrder = v
	}
	if q := r.URL.Query().Get("q"); q != "" {
		pag.Search = q
	}
	return pag
}

// helper: "1, 2,3" -> []int64{1,2,3}; ignores invalid/empty parts
func parseInt64CSV(s string) []int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	seen := make(map[int64]struct{})
	out := make([]int64, 0, 4)
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		v, err := strconv.ParseInt(p, 10, 64)
		if err != nil {
			continue // or return nil / error if you prefer strict
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

// parseEncryptedInt64CSV handles CSV values that may be encrypted IDs.
// It tries to decrypt each value first; if that fails, falls back to plain int parsing.
func parseEncryptedInt64CSV(s string) []int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	seen := make(map[int64]struct{})
	out := make([]int64, 0, 4)
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var v int64
		// Try decrypting first
		decrypted, err := secureid.DecryptID(p)
		if err == nil {
			v, err = strconv.ParseInt(decrypted, 10, 64)
			if err != nil {
				continue
			}
		} else {
			// Fallback to plain int
			v, err = strconv.ParseInt(p, 10, 64)
			if err != nil {
				continue
			}
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func int64p(v int64) *int64 { return &v }

// parseEncryptedInt64 decrypts a single encrypted ID string and returns the int64 value.
// Falls back to plain int parsing if decryption fails. Returns 0 on any error.
func parseEncryptedInt64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	if decrypted, err := secureid.DecryptID(s); err == nil {
		if v, err := strconv.ParseInt(decrypted, 10, 64); err == nil {
			return v
		}
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// parseFlexDate parses a date string in YYYY-MM-DD or RFC3339 format.
func parseFlexDate(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

// e.g. type DateOnly time.Time
func ParseDateRangeFromURL(r *http.Request) (customdatatype.DateOnly, customdatatype.DateOnly) {
	q := r.URL.Query()
	const layout = "2006-01-02"

	// Load location with fallback to UTC if Africa/Dar_es_Salaam is not available
	loc, err := time.LoadLocation("Africa/Dar_es_Salaam")
	if err != nil {
		loc = time.UTC
	}

	now := time.Now().In(loc)

	defFrom := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	// Preserve location when truncating to date-only
	y, m, d := now.Date()
	defTo := time.Date(y, m, d, 0, 0, 0, 0, loc)

	from := defFrom
	to := defTo

	if fromStr := strings.TrimSpace(q.Get("start_date")); fromStr != "" {
		if t, err := time.ParseInLocation(layout, fromStr, loc); err == nil {
			from = t
		}
	}
	if toStr := strings.TrimSpace(q.Get("end_date")); toStr != "" {
		if t, err := time.ParseInLocation(layout, toStr, loc); err == nil {
			to = t
		}
	}

	if from.After(to) {
		from, to = to, from
	}

	return customdatatype.NewDateOnly(from), customdatatype.NewDateOnly(to)
}

func escapeForJSON(s string) string {
	b, _ := json.Marshal(s)
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return string(b[1 : len(b)-1])
	}
	return s
}
