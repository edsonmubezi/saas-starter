package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/edsonmubezi/myapp/internal/platform/logging"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

const (
	DefaultBatchSize      = 500
	DefaultPollInterval   = 5 * time.Second
	MaxRetries            = 10
	InitialBackoff        = 1 * time.Second
)

// OutboxWorker processes events from outbox to events table
type OutboxWorker struct {
	db           *pgxpool.Pool
	logger       *zap.Logger
	pollInterval time.Duration
	batchSize    int
}

// NewOutboxWorker creates a new worker instance
func NewOutboxWorker(db *pgxpool.Pool, pollInterval time.Duration, batchSize int) *OutboxWorker {
	if pollInterval == 0 {
		pollInterval = DefaultPollInterval
	}
	if batchSize == 0 {
		batchSize = DefaultBatchSize
	}

	return &OutboxWorker{
		db:           db,
		logger:       logging.GetLogger(),
		pollInterval: pollInterval,
		batchSize:    batchSize,
	}
}

// Start begins processing outbox events
func (w *OutboxWorker) Start(ctx context.Context) error {
	w.logger.Info("Starting audit outbox worker",
		zap.Duration("poll_interval", w.pollInterval),
		zap.Int("batch_size", w.batchSize),
	)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// Process once immediately
	if err := w.processBatch(ctx); err != nil {
		w.logger.Error("Failed to process initial batch", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Stopping audit outbox worker")
			return ctx.Err()
		case <-ticker.C:
			if err := w.processBatch(ctx); err != nil {
				w.logger.Error("Failed to process outbox batch", zap.Error(err))
			}
		}
	}
}

// processBatch fetches and processes a batch of outbox entries
func (w *OutboxWorker) processBatch(ctx context.Context) error {
	// Fetch undelivered entries
	query := `
		SELECT id, event_json, retry_count
		FROM audit.outbox
		WHERE delivered_at IS NULL
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := w.db.Query(ctx, query, w.batchSize)
	if err != nil {
		return fmt.Errorf("failed to fetch outbox entries: %w", err)
	}
	defer rows.Close()

	processed := 0
	failed := 0

	for rows.Next() {
		var entry OutboxEntry
		if err := rows.Scan(&entry.ID, &entry.EventJSON, &entry.RetryCount); err != nil {
			w.logger.Error("Failed to scan outbox entry", zap.Error(err))
			failed++
			continue
		}

		if err := w.processEntry(ctx, &entry); err != nil {
			w.logger.Error("Failed to process outbox entry",
				zap.Int64("id", entry.ID),
				zap.Int("retry_count", entry.RetryCount),
				zap.Error(err),
			)
			failed++
			w.handleFailure(ctx, &entry, err)
		} else {
			processed++
		}
	}

	if processed > 0 || failed > 0 {
		w.logger.Info("Processed outbox batch",
			zap.Int("processed", processed),
			zap.Int("failed", failed),
		)
	}

	return nil
}

// processEntry processes a single outbox entry
func (w *OutboxWorker) processEntry(ctx context.Context, entry *OutboxEntry) error {
	// Unmarshal event
	var event AuditEvent
	if err := json.Unmarshal(entry.EventJSON, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Compute signature
	sig, err := ComputeSignature(&event)
	if err != nil {
		return fmt.Errorf("failed to compute signature: %w", err)
	}
	event.SigSHA256 = sig

	// Begin transaction
	tx, err := w.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Insert into events table (idempotent on audit_id)
	insertQuery := `
		INSERT INTO audit.events (
			audit_id, occurred_at, actor_type, actor_id, tenant_id,
			action, resource_type, resource_id, request_id, ip, user_agent,
			origin_service, severity, before_json, after_json,
			schema_version, sig_sha256
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
		)
		ON CONFLICT (audit_id) DO NOTHING
	`

	_, err = tx.Exec(ctx, insertQuery,
		event.AuditID, event.OccurredAt, event.ActorType, event.ActorID, event.TenantID,
		event.Action, event.ResourceType, event.ResourceID, event.RequestID, event.IP,
		event.UserAgent, event.OriginService, event.Severity, event.BeforeJSON,
		event.AfterJSON, event.SchemaVersion, event.SigSHA256,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into events: %w", err)
	}

	// Mark as delivered
	updateQuery := `
		UPDATE audit.outbox
		SET delivered_at = NOW()
		WHERE id = $1
	`

	_, err = tx.Exec(ctx, updateQuery, entry.ID)
	if err != nil {
		return fmt.Errorf("failed to mark as delivered: %w", err)
	}

	// Commit
	return tx.Commit(ctx)
}

// handleFailure increments retry count or moves to dead letter
func (w *OutboxWorker) handleFailure(ctx context.Context, entry *OutboxEntry, processingErr error) {
	entry.RetryCount++

	if entry.RetryCount >= MaxRetries {
		// Move to dead letter queue
		w.moveToDeadLetter(ctx, entry, processingErr)
		return
	}

	// Increment retry count with exponential backoff
	errorMsg := processingErr.Error()
	updateQuery := `
		UPDATE audit.outbox
		SET retry_count = retry_count + 1,
		    last_error = $2
		WHERE id = $1
	`

	_, err := w.db.Exec(ctx, updateQuery, entry.ID, errorMsg)
	if err != nil {
		w.logger.Error("Failed to update retry count", zap.Error(err))
	}
}

// moveToDeadLetter moves an entry to the dead letter queue
func (w *OutboxWorker) moveToDeadLetter(ctx context.Context, entry *OutboxEntry, err error) {
	tx, txErr := w.db.Begin(ctx)
	if txErr != nil {
		w.logger.Error("Failed to begin transaction for dead letter", zap.Error(txErr))
		return
	}
	defer tx.Rollback(ctx)

	// Insert into dead letter
	insertQuery := `
		INSERT INTO audit.dead_letter (event_json, error_message, retry_count, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, insertErr := tx.Exec(ctx, insertQuery, entry.EventJSON, err.Error(), entry.RetryCount, entry.CreatedAt)
	if insertErr != nil {
		w.logger.Error("Failed to insert into dead letter", zap.Error(insertErr))
		return
	}

	// Delete from outbox
	deleteQuery := `DELETE FROM audit.outbox WHERE id = $1`
	_, deleteErr := tx.Exec(ctx, deleteQuery, entry.ID)
	if deleteErr != nil {
		w.logger.Error("Failed to delete from outbox", zap.Error(deleteErr))
		return
	}

	if commitErr := tx.Commit(ctx); commitErr != nil {
		w.logger.Error("Failed to commit dead letter transaction", zap.Error(commitErr))
		return
	}

	w.logger.Warn("Moved event to dead letter queue",
		zap.Int64("outbox_id", entry.ID),
		zap.Int("retries", entry.RetryCount),
		zap.Error(err),
	)
}
