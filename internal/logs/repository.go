// internal/logs/repository.go
package logs

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

// LogRepository defines the methods for saving logs
type LogRepository interface {
	SaveLog(ctx context.Context, log *Log) error
}

// PostgresLogRepository is the implementation of LogRepository using PostgreSQL
type PostgresLogRepository struct {
	db *pgxpool.Pool
}

// NewPostgresLogRepository creates a new instance of PostgresLogRepository
func NewPostgresLogRepository(db *pgxpool.Pool) *PostgresLogRepository {
	return &PostgresLogRepository{db: db}
}

// SaveLog saves a log entry to the database
func (r *PostgresLogRepository) SaveLog(ctx context.Context, log *Log) error {
	// SQL query to insert a new log entry into the logs table
	sql := `INSERT INTO logs (user_id, action, ip_address, user_agent, created_at) 
	        VALUES ($1, $2, $3, $4, $5) RETURNING id`

	// Execute the query
	err := r.db.QueryRow(ctx, sql, log.UserID, log.Action, log.IPAddress, log.UserAgent, log.CreatedAt).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("could not save log: %v", err)
	}
	return nil
}
