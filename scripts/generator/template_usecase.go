package generator

import "text/template"

var UsecaseTestTemplate = template.Must(template.New("usecase_test").Parse(`
package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"{{ .ModulePath }}/internal/{{ .EntityLower }}"
	"{{ .ModulePath }}/internal/shared"
	"{{ .ModulePath }}/api/middleware"
	"{{ .ModulePath }}/pkg/pagination"
)

// === MOCKS ===

type mockRepo struct {
	mock.Mock
	{{ .EntityLower }}.{{ .InterfaceName }}
}

type mockChecker struct {
	mock.Mock
	shared.NameExistenceChecker
}

func (m *mockRepo) Create{{ .Entity }}(ctx context.Context, obj *{{ .EntityLower }}.{{ .Entity }}) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, obj)
	return args.Get(0).(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
}

func (m *mockRepo) Get{{ .Entity }}ByID(ctx context.Context, id int64) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
}

func (m *mockRepo) Update{{ .Entity }}(ctx context.Context, obj *{{ .EntityLower }}.{{ .Entity }}) (*{{ .EntityLower }}.{{ .Entity }}, error) {
	args := m.Called(ctx, obj)
	return args.Get(0).(*{{ .EntityLower }}.{{ .Entity }}), args.Error(1)
}

func (m *mockRepo) SoftDelete{{ .Entity }}(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockRepo) GetAll{{ .EntityPlural }}(ctx context.Context, pag pagination.Pagination, orgID int64) (pagination.Result[{{ .EntityLower }}.{{ .Entity }}], error) {
	args := m.Called(ctx, pag, orgID)
	return args.Get(0).(pagination.Result[{{ .EntityLower }}.{{ .Entity }}]), args.Error(1)
}

func (c *mockChecker) ExistsByNameInOrg(ctx context.Context, table, field, value string, orgID int64, excludeIDs ...int64) (bool, error) {
	args := c.Called(ctx, table, field, value, orgID, excludeIDs)
	return args.Bool(0), args.Error(1)
}

// === TESTS ===

func TestCreate{{ .Entity }}_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepo)
	mockChecker := new(mockChecker)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, mockChecker)

	input := &{{ .EntityLower }}.Add{{ .Entity }}{
		{{ .AddInputExample }}
	}

	expected := &{{ .EntityLower }}.{{ .Entity }}{
		ID:             123,
		{{ .AddInputExample }}
	}

	mockChecker.On("ExistsByNameInOrg", ctx, "{{ .TableName }}", "{{ .NameField }}", input.{{ .NameFieldGo }}, int64(1), []int64(nil)).Return(false, nil)
	mockRepo.On("Create{{ .Entity }}", ctx, mock.AnythingOfType("*{{ .EntityLower }}.{{ .Entity }}")).Return(expected, nil)

	result, err := usecase.Create{{ .Entity }}(ctx, input, 1)
	require.NoError(t, err)
	require.Equal(t, expected.ID, result.ID)
}

func TestCreate{{ .Entity }}_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepo)
	mockChecker := new(mockChecker)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, mockChecker)

	input := &{{ .EntityLower }}.Add{{ .Entity }}{
		{{ .AddInputExample }}
	}

	mockChecker.On("ExistsByNameInOrg", ctx, "{{ .TableName }}", "{{ .NameField }}", input.{{ .NameFieldGo }}, int64(1), []int64(nil)).Return(true, nil)

	_, err := usecase.Create{{ .Entity }}(ctx, input, 1)
	require.ErrorContains(t, err, "already exists")
}

func TestGet{{ .Entity }}ByID_Unauthorized(t *testing.T) {
	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{OrganizationID: 2})
	mockRepo := new(mockRepo)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, nil)

	dept := &{{ .EntityLower }}.{{ .Entity }}{ID: 1, OrganizationID: 1}
	mockRepo.On("Get{{ .Entity }}ByID", ctx, int64(1)).Return(dept, nil)

	_, err := usecase.Get{{ .Entity }}ByID(ctx, 1)
	require.ErrorContains(t, err, "unauthorized")
}

func TestGetAll{{ .EntityPlural }}_Forbidden(t *testing.T) {
	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{})

	mockRepo := new(mockRepo)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, nil)

	_, err := usecase.GetAll{{ .EntityPlural }}(ctx, pagination.Pagination{})
	require.ErrorContains(t, err, "unauthorized")
}

func TestUpdate{{ .Entity }}_Success(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockRepo)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, nil)

	orgID := int64(1)
	existing := &{{ .EntityLower }}.{{ .Entity }}{ID: 5, OrganizationID: orgID}
	updated := &{{ .EntityLower }}.{{ .Entity }}{ID: 5, {{ .UpdateInputExample }} OrganizationID: orgID}

	mockRepo.On("Get{{ .Entity }}ByID", ctx, int64(5)).Return(existing, nil)
	mockRepo.On("Update{{ .Entity }}", ctx, mock.AnythingOfType("*{{ .EntityLower }}.{{ .Entity }}")).Return(updated, nil)

	input := &{{ .EntityLower }}.Update{{ .Entity }}{
		{{ .UpdateInputExample }}
	}
	result, err := usecase.Update{{ .Entity }}(ctx, 5, input, orgID)

	require.NoError(t, err)
	require.Equal(t, updated.ID, result.ID)
}

func TestSoftDelete{{ .Entity }}_Unauthorized(t *testing.T) {
	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{OrganizationID: 999})

	mockRepo := new(mockRepo)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, nil)

	mockRepo.On("Get{{ .Entity }}ByID", ctx, int64(10)).Return(&{{ .EntityLower }}.{{ .Entity }}{ID: 10, OrganizationID: 1}, nil)

	err := usecase.SoftDelete{{ .Entity }}(ctx, 10)
	require.ErrorContains(t, err, "unauthorized")
}

func TestSoftDelete{{ .Entity }}_Success(t *testing.T) {
	ctx := middleware.WithAuthContext(context.Background(), middleware.AuthContext{OrganizationID: 1})

	mockRepo := new(mockRepo)
	usecase := {{ .EntityLower }}.New{{ .Entity }}UseCase(mockRepo, nil)

	mockRepo.On("Get{{ .Entity }}ByID", ctx, int64(10)).Return(&{{ .EntityLower }}.{{ .Entity }}{ID: 10, OrganizationID: 1}, nil)
	mockRepo.On("SoftDelete{{ .Entity }}", ctx, int64(10)).Return(nil)

	err := usecase.SoftDelete{{ .Entity }}(ctx, 10)
	require.NoError(t, err)
}`))
