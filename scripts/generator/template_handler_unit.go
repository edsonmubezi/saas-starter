package generator

import "text/template"

var HandlerUnitTestTemplate = template.Must(template.New("handler_unit_test").Parse(`
package tests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/gorilla/mux"

	"{{ .ModulePath }}/api/handler"
	"{{ .ModulePath }}/api/middleware"
	"{{ .ModulePath }}/internal/{{ .EntityLower }}"
	"{{ .ModulePath }}/pkg/pagination"
)

type mockUseCase struct {
	mock.Mock
	{{ .EntityLower }}.{{ .Entity }}UseCase
}

func (m *mockUseCase) Create{{ .Entity }}(ctx context.Context, input *{{ .EntityLower }}.Add{{ .Entity }}, orgID int64) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, input, orgID)
	if v := args.Get(0); v != nil {
		return v.(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUseCase) Update{{ .Entity }}(ctx context.Context, id int64, input *{{ .EntityLower }}.Update{{ .Entity }}, orgID int64) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, id, input, orgID)
	if v := args.Get(0); v != nil {
		return v.(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUseCase) Get{{ .Entity }}ByID(ctx context.Context, id int64) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, id)
	if v := args.Get(0); v != nil {
		return v.(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUseCase) GetAll{{ .EntityPlural }}(ctx context.Context, pag pagination.Pagination) (pagination.Result[{{ .EntityLower }}.{{ .Entity }}], error) {
	args := m.Called(ctx, pag)
	if v := args.Get(0); v != nil {
		return v.(pagination.Result[{{ .EntityLower }}.{{ .Entity }}]), args.Error(1)
	}
	return pagination.Result[{{ .EntityLower }}.{{ .Entity }}]{}, args.Error(1)
}

func (m *mockUseCase) SoftDelete{{ .Entity }}(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func TestCreate{{ .Entity }}Handler_Success(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	authCtx := middleware.AuthContext{OrganizationID: 1}
	ctx := middleware.WithAuthContext(context.Background(), authCtx)

	expected := &{{ .EntityLower }}.{{ .Entity }}{ID: 1, {{ .NameField }}: "Example", OrganizationID: 1}

	mockUC.On("Create{{ .Entity }}", mock.Anything, mock.AnythingOfType("*{{ .EntityLower }}.Add{{ .Entity }}"), int64(1)).
		Return(expected, nil)

	reqBody := ` + "`" + `{"{{ .NameFieldSnake }}": "Example"}` + "`" + `
	req := httptest.NewRequest(http.MethodPost, "/{{ .EntityLower }}s", strings.NewReader(reqBody)).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	require.Contains(t, rr.Body.String(), "Example")
}

func TestCreate{{ .Entity }}Handler_ValidationError(t *testing.T) {
	reqBody := ` + "`" + `{"invalid_json":}` + "`" + `
	req := httptest.NewRequest(http.MethodPost, "/{{ .EntityLower }}s", strings.NewReader(reqBody))
	rr := httptest.NewRecorder()

	handler.Create{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid JSON format")
}

func TestCreate{{ .Entity }}Handler_UseCaseFails(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{OrganizationID: 1})
	req := httptest.NewRequest(http.MethodPost, "/{{ .EntityLower }}s", strings.NewReader(` + "`" + `{"{{ .NameFieldSnake }}": "Example"}` + "`" + `)).WithContext(ctx)
	rr := httptest.NewRecorder()

	mockUC.On("Create{{ .Entity }}", mock.Anything, mock.AnythingOfType("*{{ .EntityLower }}.Add{{ .Entity }}"), int64(1)).Return(nil, fmt.Errorf("simulated error"))

	handler.Create{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	require.Contains(t, rr.Body.String(), "Failed to create {{ .EntityLower }}")
}

func TestCreate{{ .Entity }}Handler_MissingOrg(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/{{ .EntityLower }}s", strings.NewReader(` + "`" + `{"{{ .NameFieldSnake }}": "Example"}` + "`" + `))
	ctx := middleware.WithAuthContext(req.Context(), middleware.AuthContext{})
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.Create{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "organization_id is required")
}

func TestGet{{ .Entity }}ByIDHandler_Success(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	expected := &{{ .EntityLower }}.{{ .Entity }}{ID: 1, {{ .NameField }}: "Found", OrganizationID: 1}

	req := httptest.NewRequest(http.MethodGet, "/{{ .EntityLower }}s/1", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	mockUC.On("Get{{ .Entity }}ByID", mock.Anything, int64(1)).Return(expected, nil)

	handler.Get{{ .Entity }}ByIDHandler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "Found")
}

func TestGet{{ .Entity }}ByIDHandler_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/{{ .EntityLower }}s/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.Get{{ .Entity }}ByIDHandler(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid")
}

func TestUpdate{{ .Entity }}Handler_Success(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{OrganizationID: 1})
	input := &{{ .EntityLower }}.Update{{ .Entity }}{ {{ .NameField }}: "Updated" }
	expected := &{{ .EntityLower }}.{{ .Entity }}{ID: 3, {{ .NameField }}: "Updated", OrganizationID: 1}

	handler.ParseUpdate{{ .Entity }}Body = func(r *http.Request) (*{{ .EntityLower }}.Update{{ .Entity }}, []middleware.FieldError, error) {
		return input, nil, nil
	}

	req := httptest.NewRequest(http.MethodPut, "/{{ .EntityLower }}s/3", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "3"})
	req = req.WithContext(ctx)

	mockUC.On("Update{{ .Entity }}", mock.Anything, int64(3), input, int64(1)).Return(expected, nil)

	rr := httptest.NewRecorder()
	handler.Update{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "updated")
}

func TestUpdate{{ .Entity }}Handler_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/{{ .EntityLower }}s/abc", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	handler.Update{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid")
}

func TestGetAll{{ .EntityPlural }}Handler_Success(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	authCtx := middleware.AuthContext{OrganizationID: 1}
	ctx := middleware.WithAuthContext(context.Background(), authCtx)

	expected := pagination.Result[{{ .EntityLower }}.{{ .Entity }}]{
		Data: []{{ .EntityLower }}.{{ .Entity }}{
			{ID: 1, {{ .NameField }}: "One"},
			{ID: 2, {{ .NameField }}: "Two"},
		},
		TotalCount: 2,
	}

	mockUC.On("GetAll{{ .EntityPlural }}", mock.Anything, mock.Anything).Return(expected, nil)

	req := httptest.NewRequest(http.MethodGet, "/{{ .EntityLower }}s?page=1&page_size=10", nil).WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetAll{{ .EntityPlural }}Handler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "One")
	require.Contains(t, rr.Body.String(), "Two")
}

func TestSoftDelete{{ .Entity }}Handler_Success(t *testing.T) {
	mockUC := new(mockUseCase)
	handler.Set{{ .Entity }}UseCase(mockUC)

	req := httptest.NewRequest(http.MethodDelete, "/{{ .EntityLower }}s/5", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "5"})
	req = req.WithContext(middleware.WithAuthContext(req.Context(), middleware.AuthContext{OrganizationID: 1}))

	mockUC.On("SoftDelete{{ .Entity }}", mock.Anything, int64(5)).Return(nil)

	rr := httptest.NewRecorder()
	handler.SoftDelete{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "deleted")
}

func TestSoftDelete{{ .Entity }}Handler_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/{{ .EntityLower }}s/invalid", nil)
	req = mux.SetURLVars(req, map[string]string{"id": "invalid"})
	rr := httptest.NewRecorder()

	handler.SoftDelete{{ .Entity }}Handler(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
	require.Contains(t, rr.Body.String(), "Invalid")
}
`))
