package importlog

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Repository defines the data access interface for import logs.
type Repository interface {
	CreateSession(ctx context.Context, s *ImportSession) (int64, error)
	CreateFailedRecords(ctx context.Context, records []ImportFailedRecord) error
	GetSessions(ctx context.Context, orgID int64, importType string, limit, offset int) ([]ImportSession, int, error)
	GetFailedRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error)
	GetDuplicateRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error)
	MarkRecordResolved(ctx context.Context, recordID, orgID int64, resolution string) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL-backed import log repository.
func NewPostgresRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{db: db}
}

// CreateSession inserts a new import session and returns its ID.
func (r *PostgresRepository) CreateSession(ctx context.Context, s *ImportSession) (int64, error) {
	query := `
		INSERT INTO import_sessions
			(import_type, status, total_rows, successful, failed, duplicates, error_file, created_by, created_by_name, organization_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`
	var id int64
	err := r.db.QueryRow(ctx, query,
		s.ImportType, s.Status, s.TotalRows, s.Successful, s.Failed, s.Duplicates,
		nilIfEmpty(s.ErrorFile), s.CreatedBy, s.CreatedByName, s.OrgID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert import session: %w", err)
	}
	return id, nil
}

// CreateFailedRecords batch-inserts failed records using a single query with multiple VALUES.
func (r *PostgresRepository) CreateFailedRecords(ctx context.Context, records []ImportFailedRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Build batch INSERT: INSERT INTO ... VALUES ($1,$2,...), ($N+1,$N+2,...), ...
	const colCount = 9
	valueStrings := make([]string, len(records))
	args := make([]interface{}, 0, len(records)*colCount)

	for i, rec := range records {
		base := i * colCount
		valueStrings[i] = fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9)
		args = append(args,
			rec.ImportSessionID,
			rec.RowNumber,
			rec.RecordName,
			rec.Reason,
			rec.RecordType,
			nilIfEmpty(rec.ColumnName),
			nilIfEmpty(rec.Value),
			rec.ExistingRecordID, // nil for non-duplicate records
			rec.OrgID,
		)
	}

	query := `
		INSERT INTO import_failed_records
			(import_session_id, row_number, record_name, reason, record_type, column_name, value, existing_employee_id, organization_id)
		VALUES ` + strings.Join(valueStrings, ", ")

	_, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("batch insert failed records: %w", err)
	}
	return nil
}

// GetSessions returns paginated import sessions for an organization, optionally filtered by import_type.
func (r *PostgresRepository) GetSessions(ctx context.Context, orgID int64, importType string, limit, offset int) ([]ImportSession, int, error) {
	// Count query
	countQuery := `SELECT COUNT(*) FROM import_sessions WHERE organization_id = $1`
	countArgs := []interface{}{orgID}
	argIdx := 2

	if importType != "" {
		countQuery += fmt.Sprintf(" AND import_type = $%d", argIdx)
		countArgs = append(countArgs, importType)
		argIdx++
	}

	var total int
	if err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count import sessions: %w", err)
	}

	// Data query
	dataQuery := `
		SELECT id, import_type, status, total_rows, successful, failed, COALESCE(duplicates, 0),
		       COALESCE(error_file, '') AS error_file,
		       created_by, created_by_name, organization_id,
		       TO_CHAR(created_at, 'YYYY-MM-DD HH24:MI') AS created_at
		FROM import_sessions
		WHERE organization_id = $1
	`
	dataArgs := []interface{}{orgID}
	dataArgIdx := 2

	if importType != "" {
		dataQuery += fmt.Sprintf(" AND import_type = $%d", dataArgIdx)
		dataArgs = append(dataArgs, importType)
		dataArgIdx++
	}

	dataQuery += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", dataArgIdx, dataArgIdx+1)
	dataArgs = append(dataArgs, limit, offset)

	rows, err := r.db.Query(ctx, dataQuery, dataArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("query import sessions: %w", err)
	}
	defer rows.Close()

	var sessions []ImportSession
	for rows.Next() {
		var s ImportSession
		if err := rows.Scan(
			&s.ID, &s.ImportType, &s.Status, &s.TotalRows, &s.Successful, &s.Failed, &s.Duplicates,
			&s.ErrorFile, &s.CreatedBy, &s.CreatedByName, &s.OrgID, &s.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan import session: %w", err)
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// GetFailedRecords returns all failed records for a given import session.
func (r *PostgresRepository) GetFailedRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error) {
	query := `
		SELECT id, import_session_id, row_number, record_name, reason, record_type,
		       COALESCE(column_name, '') AS column_name,
		       COALESCE(value, '') AS value,
		       organization_id
		FROM import_failed_records
		WHERE import_session_id = $1 AND organization_id = $2
		ORDER BY row_number ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID, orgID)
	if err != nil {
		return nil, fmt.Errorf("query failed records: %w", err)
	}
	defer rows.Close()

	var records []ImportFailedRecord
	for rows.Next() {
		var rec ImportFailedRecord
		if err := rows.Scan(
			&rec.ID, &rec.ImportSessionID, &rec.RowNumber, &rec.RecordName,
			&rec.Reason, &rec.RecordType, &rec.ColumnName, &rec.Value, &rec.OrgID,
		); err != nil {
			return nil, fmt.Errorf("scan failed record: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

// GetDuplicateRecords returns unresolved duplicate records for a given import session.
func (r *PostgresRepository) GetDuplicateRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error) {
	query := `
		SELECT ifr.id, ifr.import_session_id, ifr.row_number, ifr.record_name, ifr.reason, ifr.record_type,
		       COALESCE(ifr.column_name, '') AS column_name,
		       COALESCE(ifr.value, '') AS value,
		       ifr.existing_employee_id,
		       ifr.organization_id
		FROM import_failed_records ifr
		WHERE ifr.import_session_id = $1
		  AND ifr.organization_id = $2
		  AND ifr.record_type = 'duplicate'
		  AND ifr.resolved_at IS NULL
		ORDER BY ifr.row_number ASC
	`
	rows, err := r.db.Query(ctx, query, sessionID, orgID)
	if err != nil {
		return nil, fmt.Errorf("query duplicate records: %w", err)
	}
	defer rows.Close()

	var records []ImportFailedRecord
	for rows.Next() {
		var rec ImportFailedRecord
		if err := rows.Scan(
			&rec.ID, &rec.ImportSessionID, &rec.RowNumber, &rec.RecordName,
			&rec.Reason, &rec.RecordType, &rec.ColumnName, &rec.Value,
			&rec.ExistingRecordID, &rec.OrgID,
		); err != nil {
			return nil, fmt.Errorf("scan duplicate record: %w", err)
		}
		records = append(records, rec)
	}

	return records, nil
}

// MarkRecordResolved marks a duplicate record as resolved.
func (r *PostgresRepository) MarkRecordResolved(ctx context.Context, recordID, orgID int64, resolution string) error {
	query := `
		UPDATE import_failed_records
		SET resolved_at = NOW(), resolution = $3
		WHERE id = $1 AND organization_id = $2 AND record_type = 'duplicate'
	`
	tag, err := r.db.Exec(ctx, query, recordID, orgID, resolution)
	if err != nil {
		return fmt.Errorf("mark record resolved: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("record not found or already resolved")
	}
	return nil
}

// nilIfEmpty returns nil if the string is empty, otherwise returns the string pointer.
func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
