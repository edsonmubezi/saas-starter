package generator

import "text/template"

var RepoTestTemplate = template.Must(template.New("repo_test").Parse(`
package {{ .EntityLower }}

import (
	"context"
	"testing"
	"time"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/edsonmubezi/myapp/pkg/database"
	"github.com/edsonmubezi/myapp/internal/{{ .EntityLower }}"
	"github.com/edsonmubezi/myapp/pkg/pagination"
)

func truncateTable(ctx context.Context, db *pgxpool.Pool, tableName string) {
	_, _ = db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", tableName))
}

func setupTestRepo(t *testing.T) (*{{ .EntityLower }}.Postgres{{ .Entity }}Repository, func()) {
	db, err := database.SetupTestDB()
	require.NoError(t, err)

	cleanup := func() {
		truncateTable(context.Background(), db, "{{ .TableName }}")
	}

	return {{ .EntityLower }}.NewPostgres{{ .Entity }}Repository(db), cleanup
}

func generateTest{{ .Entity }}() *{{ .EntityLower }}.{{ .Entity }} {
	return &{{ .EntityLower }}.{{ .Entity }}{
		{{- range .Fields }}
		{{ .Name }}: {{ .SampleValue }},
		{{- end }}
	}
}

func Test{{ .Entity }}Repository_Create{{ .Entity }}(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	input := generateTest{{ .Entity }}()
	result, err := repo.Create{{ .Entity }}(context.Background(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, input.ID, result.ID)
}

func Test{{ .Entity }}Repository_Get{{ .Entity }}ByID(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	created, _ := repo.Create{{ .Entity }}(context.Background(), generateTest{{ .Entity }}())
	found, err := repo.Get{{ .Entity }}ByID(context.Background(), created.ID)

	require.NoError(t, err)
	require.NotNil(t, found)
	require.Equal(t, created.ID, found.ID)
}

func Test{{ .Entity }}Repository_Update{{ .Entity }}(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	created, _ := repo.Create{{ .Entity }}(context.Background(), generateTest{{ .Entity }}())

	// modify sample update field
	{{- range .UpdatableFields }}
	created.{{ .Name }} = {{ .UpdatedSampleValue }}
	{{- end }}

	updated, err := repo.Update{{ .Entity }}(context.Background(), created)
	require.NoError(t, err)
	require.Equal(t, "Updated Department", updated.DepartmentName)
	{{- range .UpdatableFields }}
	require.Equal(t, {{ .UpdatedSampleValue }}, updated.{{ .Name }})
	{{- end }}
}

func Test{{ .Entity }}Repository_SoftDelete{{ .Entity }}(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	created, _ := repo.Create{{ .Entity }}(context.Background(), generateTest{{ .Entity }}())
	err := repo.SoftDelete{{ .Entity }}(context.Background(), created.ID)

	require.NoError(t, err)

	deleted, err := repo.Get{{ .Entity }}ByID(context.Background(), created.ID)
	require.Error(t, err)
	require.Nil(t, deleted)
}

func Test{{ .Entity }}Repository_GetAll{{ .Entity }}s(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	for i := 0; i < 2; i++ {
		_, _ = repo.Create{{ .Entity }}(context.Background(), generateTest{{ .Entity }}())
	}

	result, err := repo.GetAll{{ .Entity }}s(context.Background(), pagination.Pagination{
		Page: 1,
		PageSize: 10,
	}, 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(result.Data), 1)
}

func Test{{ .Entity }}Repository_Get{{ .Entity }}sByOrganization(t *testing.T) {
	repo, cleanup := setupTestRepo(t)
	defer cleanup()

	_, _ = repo.Create{{ .Entity }}(context.Background(), generateTest{{ .Entity }}())
	list, err := repo.Get{{ .Entity }}sByOrganization(context.Background(), 1)

	require.NoError(t, err)
	require.GreaterOrEqual(t, len(list), 1)
}

`))
