package importlog

import (
	"context"
	"fmt"
	"log"
)

// UseCase defines the business logic for import logging.
type UseCase interface {
	SaveImportResult(ctx context.Context, session *ImportSession, records []ImportFailedRecord) error
	GetImportHistory(ctx context.Context, orgID int64, importType string, page, pageSize int) ([]ImportSession, int, error)
	GetSessionFailedRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error)
	GetDuplicateRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error)
	MarkRecordResolved(ctx context.Context, recordID, orgID int64, resolution string) error
}

type useCaseImpl struct {
	repo Repository
}

// NewUseCase creates a new import log use case.
func NewUseCase(repo Repository) UseCase {
	return &useCaseImpl{repo: repo}
}

// SaveImportResult persists an import session and its failed records.
// Errors are logged but not propagated — import logging must never break the import flow.
func (u *useCaseImpl) SaveImportResult(ctx context.Context, session *ImportSession, records []ImportFailedRecord) error {
	sessionID, err := u.repo.CreateSession(ctx, session)
	if err != nil {
		log.Printf("[ImportLog] Failed to save import session: %v", err)
		return fmt.Errorf("save import session: %w", err)
	}

	if len(records) > 0 {
		// Set the session ID on all records
		for i := range records {
			records[i].ImportSessionID = sessionID
		}

		if err := u.repo.CreateFailedRecords(ctx, records); err != nil {
			log.Printf("[ImportLog] Failed to save %d failed records for session %d: %v", len(records), sessionID, err)
			return fmt.Errorf("save failed records: %w", err)
		}
	}

	return nil
}

// GetImportHistory returns paginated import sessions for an organization.
func (u *useCaseImpl) GetImportHistory(ctx context.Context, orgID int64, importType string, page, pageSize int) ([]ImportSession, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * pageSize
	return u.repo.GetSessions(ctx, orgID, importType, pageSize, offset)
}

// GetSessionFailedRecords returns all failed records for a specific import session.
func (u *useCaseImpl) GetSessionFailedRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error) {
	return u.repo.GetFailedRecords(ctx, sessionID, orgID)
}

// GetDuplicateRecords returns unresolved duplicate records for a specific import session.
func (u *useCaseImpl) GetDuplicateRecords(ctx context.Context, sessionID, orgID int64) ([]ImportFailedRecord, error) {
	return u.repo.GetDuplicateRecords(ctx, sessionID, orgID)
}

// MarkRecordResolved marks a duplicate record as resolved with the given resolution.
func (u *useCaseImpl) MarkRecordResolved(ctx context.Context, recordID, orgID int64, resolution string) error {
	return u.repo.MarkRecordResolved(ctx, recordID, orgID, resolution)
}
