package generator

import "text/template"

var HandlerIntegrationTestTemplate = template.Must(template.New("handler_integration_test").Parse(`package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"{{ .ModulePath }}/api/handler"
	"{{ .ModulePath }}/api/middleware"
	"{{ .ModulePath }}/internal/{{ .EntityLower }}"
	"{{ .ModulePath }}/pkg/database"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupTestRouter(db *pgxpool.Pool) http.Handler {
	repo := {{ .EntityLower }}.NewPostgres{{ .Entity }}Repository(db)
	checker := &fakeChecker{}
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(repo, checker)
	handler.Set{{ .Entity }}UseCase(usecase)

	r := mux.NewRouter()
	r.HandleFunc("/{{ .EntityLower }}s", handler.Create{{ .Entity }}Handler).Methods("POST")
	r.HandleFunc("/{{ .EntityLower }}s/{id}", handler.Get{{ .Entity }}ByIDHandler).Methods("GET")
	r.HandleFunc("/{{ .EntityLower }}s/{id}", handler.Update{{ .Entity }}Handler).Methods("PUT")
	r.HandleFunc("/{{ .EntityLower }}s", handler.GetAll{{ .EntityPlural }}Handler).Methods("GET")
	r.HandleFunc("/{{ .EntityLower }}s/{id}", handler.SoftDelete{{ .Entity }}Handler).Methods("DELETE")
	return r
}

type fakeChecker struct{}
func (f *fakeChecker) ExistsByNameInOrg(ctx context.Context, table, field, value string, orgID int64, extra ...int64) (bool, error) {
	return false, nil
}

func intToStr(id int64) string {
	return fmt.Sprintf("%d", id)
}

func TestIntegration_Get{{ .Entity }}ByIDHandler_Success(t *testing.T) {
	db, cleanup := database.SetupTestDB()
	defer cleanup()
	router := setupTestRouter(db)

	req := httptest.NewRequest("POST", "/{{ .EntityLower }}s", bytes.NewBufferString(` + "`{{ .CreateJSONPayload }}`" + `))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(), middleware.AuthContext{OrganizationID: 1}))
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, req)
	require.Equal(t, http.StatusCreated, createRR.Code)

	var created map[string]any
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	id := int64(created["data"].(map[string]any)["id"].(float64))

	getReq := httptest.NewRequest("GET", "/{{ .EntityLower }}s/"+intToStr(id), nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"id": intToStr(id)})
	getReq = getReq.WithContext(middleware.WithAuthContext(getReq.Context(), middleware.AuthContext{OrganizationID: 1}))
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	require.Equal(t, http.StatusOK, getRR.Code)
	require.Contains(t, getRR.Body.String(), {{ printf "%q" .CreateJSONPayload }})
}

func TestIntegration_Update{{ .Entity }}Handler_Success(t *testing.T) {
	db, cleanup := database.SetupTestDB()
	defer cleanup()
	router := setupTestRouter(db)

	req := httptest.NewRequest("POST", "/{{ .EntityLower }}s", bytes.NewBufferString(` + "`{{ .CreateJSONPayload }}`" + `))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(), middleware.AuthContext{OrganizationID: 1}))
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, req)
	require.Equal(t, http.StatusCreated, createRR.Code)

	var created map[string]any
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	id := int64(created["data"].(map[string]any)["id"].(float64))

	updateReq := httptest.NewRequest("PUT", "/{{ .EntityLower }}s/"+intToStr(id), bytes.NewBufferString(` + "`{{ .UpdateJSONPayload }}`" + `))
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq = mux.SetURLVars(updateReq, map[string]string{"id": intToStr(id)})
	updateReq = updateReq.WithContext(middleware.WithAuthContext(updateReq.Context(), middleware.AuthContext{OrganizationID: 1}))
	updateRR := httptest.NewRecorder()
	router.ServeHTTP(updateRR, updateReq)

	require.Equal(t, http.StatusOK, updateRR.Code)
	require.Contains(t, updateRR.Body.String(), {{ printf "%q" .UpdateJSONPayload }})
}

func TestIntegration_GetAll{{ .EntityPlural }}Handler_Success(t *testing.T) {
	db, cleanup := database.SetupTestDB()
	defer cleanup()
	router := setupTestRouter(db)

	payloads := []string{
		` + "`{{ .CreateJSONPayload }}`" + `,
	}

	for _, p := range payloads {
		req := httptest.NewRequest("POST", "/{{ .EntityLower }}s", bytes.NewBufferString(p))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithAuthContext(req.Context(), middleware.AuthContext{OrganizationID: 1}))
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		require.Equal(t, http.StatusCreated, rr.Code)
	}

	getReq := httptest.NewRequest("GET", "/{{ .EntityLower }}s", nil)
	getReq = getReq.WithContext(middleware.WithAuthContext(getReq.Context(), middleware.AuthContext{OrganizationID: 1}))
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	require.Equal(t, http.StatusOK, getRR.Code)
	require.Contains(t, getRR.Body.String(), {{ printf "%q" .NameField }})
}

func TestIntegration_SoftDelete{{ .Entity }}Handler_Success(t *testing.T) {
	db, cleanup := database.SetupTestDB()
	defer cleanup()
	router := setupTestRouter(db)

	req := httptest.NewRequest("POST", "/{{ .EntityLower }}s", bytes.NewBufferString(` + "`{{ .CreateJSONPayload }}`" + `))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(middleware.WithAuthContext(req.Context(), middleware.AuthContext{OrganizationID: 1}))
	createRR := httptest.NewRecorder()
	router.ServeHTTP(createRR, req)
	require.Equal(t, http.StatusCreated, createRR.Code)

	var created map[string]any
	_ = json.Unmarshal(createRR.Body.Bytes(), &created)
	id := int64(created["data"].(map[string]any)["id"].(float64))

	delReq := httptest.NewRequest("DELETE", "/{{ .EntityLower }}s/"+intToStr(id), nil)
	delReq = mux.SetURLVars(delReq, map[string]string{"id": intToStr(id)})
	delReq = delReq.WithContext(middleware.WithAuthContext(delReq.Context(), middleware.AuthContext{OrganizationID: 1}))
	delRR := httptest.NewRecorder()
	router.ServeHTTP(delRR, delReq)

	require.Equal(t, http.StatusOK, delRR.Code)

	getReq := httptest.NewRequest("GET", "/{{ .EntityLower }}s/"+intToStr(id), nil)
	getReq = mux.SetURLVars(getReq, map[string]string{"id": intToStr(id)})
	getReq = getReq.WithContext(middleware.WithAuthContext(getReq.Context(), middleware.AuthContext{OrganizationID: 1}))
	getRR := httptest.NewRecorder()
	router.ServeHTTP(getRR, getReq)

	require.NotEqual(t, http.StatusOK, getRR.Code)
}
`))
