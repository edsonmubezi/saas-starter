package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/edsonmubezi/myapp/pkg/resilience"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	// DefaultQueryTimeout is the default timeout for database queries
	DefaultQueryTimeout = 30 * time.Second
	// SlowQueryThreshold is the duration above which queries are logged as slow
	SlowQueryThreshold = 2 * time.Second
)

// QueryOptions holds options for resilient query execution
type QueryOptions struct {
	// Timeout for the query (defaults to DefaultQueryTimeout)
	Timeout time.Duration
	// EnableRetry enables retry logic for transient failures
	EnableRetry bool
	// RetryConfig custom retry configuration (uses default if nil)
	RetryConfig *resilience.RetryConfig
	// LogSlowQueries enables logging of queries exceeding SlowQueryThreshold
	LogSlowQueries bool
}

// DefaultQueryOptions returns sensible defaults for query options
func DefaultQueryOptions() *QueryOptions {
	return &QueryOptions{
		Timeout:        DefaultQueryTimeout,
		EnableRetry:    true,
		RetryConfig:    nil, // Will use resilience.DefaultRetryConfig()
		LogSlowQueries: true,
	}
}

// QueryWithTimeout executes a query with automatic timeout and optional retry
func QueryWithTimeout(ctx context.Context, pool *pgxpool.Pool, opts *QueryOptions, sql string, args ...interface{}) (pgx.Rows, error) {
	if opts == nil {
		opts = DefaultQueryOptions()
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultQueryTimeout
	}

	// Create context with timeout
	queryCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	start := time.Now()
	var rows pgx.Rows
	var err error

	operation := func() error {
		rows, err = pool.Query(queryCtx, sql, args...)
		return err
	}

	// Execute with or without retry
	if opts.EnableRetry {
		err = resilience.WithRetry(queryCtx, opts.RetryConfig, operation)
	} else {
		err = operation()
	}

	duration := time.Since(start)

	// Log slow queries
	if opts.LogSlowQueries && duration > SlowQueryThreshold {
		logSlowQuery(sql, duration, args...)
	}

	// Wrap timeout errors with more context
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("query timeout after %v: %w", opts.Timeout, err)
		}
		return nil, err
	}

	return rows, nil
}

// QueryRowWithTimeout executes a query expecting a single row with timeout
func QueryRowWithTimeout(ctx context.Context, pool *pgxpool.Pool, opts *QueryOptions, sql string, args ...interface{}) pgx.Row {
	if opts == nil {
		opts = DefaultQueryOptions()
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultQueryTimeout
	}

	queryCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	// Note: We can't defer cancel() here because the row needs to be scanned
	// The caller must ensure the context is properly handled

	start := time.Now()
	row := pool.QueryRow(queryCtx, sql, args...)
	duration := time.Since(start)

	if opts.LogSlowQueries && duration > SlowQueryThreshold {
		logSlowQuery(sql, duration, args...)
	}

	// Return a wrapper that will cancel the context after scanning
	return &timeoutRow{row: row, cancel: cancel}
}

// timeoutRow wraps pgx.Row to handle context cancellation
type timeoutRow struct {
	row    pgx.Row
	cancel context.CancelFunc
}

func (r *timeoutRow) Scan(dest ...interface{}) error {
	defer r.cancel()
	return r.row.Scan(dest...)
}

// ExecWithTimeout executes a statement with automatic timeout and optional retry
func ExecWithTimeout(ctx context.Context, pool *pgxpool.Pool, opts *QueryOptions, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if opts == nil {
		opts = DefaultQueryOptions()
	}
	if opts.Timeout == 0 {
		opts.Timeout = DefaultQueryTimeout
	}

	execCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	start := time.Now()
	var tag pgconn.CommandTag
	var err error

	operation := func() error {
		tag, err = pool.Exec(execCtx, sql, args...)
		return err
	}

	if opts.EnableRetry {
		err = resilience.WithRetry(execCtx, opts.RetryConfig, operation)
	} else {
		err = operation()
	}

	duration := time.Since(start)

	if opts.LogSlowQueries && duration > SlowQueryThreshold {
		logSlowQuery(sql, duration, args...)
	}

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return tag, fmt.Errorf("exec timeout after %v: %w", opts.Timeout, err)
		}
		return tag, err
	}

	return tag, nil
}

// TxWithTimeout begins a transaction with timeout
func TxWithTimeout(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration) (pgx.Tx, context.Context, context.CancelFunc, error) {
	if timeout == 0 {
		timeout = DefaultQueryTimeout
	}

	txCtx, cancel := context.WithTimeout(ctx, timeout)

	tx, err := pool.Begin(txCtx)
	if err != nil {
		cancel()
		return nil, nil, nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	return tx, txCtx, cancel, nil
}

// ExecuteInTransaction executes a function within a transaction with automatic commit/rollback
func ExecuteInTransaction(ctx context.Context, pool *pgxpool.Pool, timeout time.Duration, fn func(tx pgx.Tx) error) error {
	tx, txCtx, cancel, err := TxWithTimeout(ctx, pool, timeout)
	if err != nil {
		return err
	}
	defer cancel()

	// Execute the function
	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(txCtx); rbErr != nil {
			log.Printf("⚠ Transaction rollback failed: %v (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(txCtx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// logSlowQuery logs a slow query with truncated SQL for security
func logSlowQuery(sql string, duration time.Duration, args ...interface{}) {
	// Truncate SQL for logging (don't log full queries which may contain sensitive data)
	truncatedSQL := sql
	if len(sql) > 100 {
		truncatedSQL = sql[:100] + "..."
	}

	log.Printf("⚠ SLOW QUERY: %v | SQL: %s | Args count: %d", duration, truncatedSQL, len(args))
}

// WithCircuitBreaker wraps a database operation with circuit breaker
func WithCircuitBreaker(ctx context.Context, fn func() error) error {
	cb := resilience.GetBreaker(resilience.ServiceDatabase)
	if cb == nil {
		return fn()
	}
	return cb.Execute(ctx, fn)
}

// QueryWithCircuitBreaker combines query with circuit breaker protection
func QueryWithCircuitBreaker(ctx context.Context, pool *pgxpool.Pool, opts *QueryOptions, sql string, args ...interface{}) (pgx.Rows, error) {
	cb := resilience.GetBreaker(resilience.ServiceDatabase)
	if cb == nil {
		return QueryWithTimeout(ctx, pool, opts, sql, args...)
	}

	var rows pgx.Rows
	err := cb.Execute(ctx, func() error {
		var innerErr error
		rows, innerErr = QueryWithTimeout(ctx, pool, opts, sql, args...)
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return nil, fmt.Errorf("database circuit breaker open: %w", err)
	}

	return rows, err
}

// ExecWithCircuitBreaker combines exec with circuit breaker protection
func ExecWithCircuitBreaker(ctx context.Context, pool *pgxpool.Pool, opts *QueryOptions, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	cb := resilience.GetBreaker(resilience.ServiceDatabase)
	if cb == nil {
		return ExecWithTimeout(ctx, pool, opts, sql, args...)
	}

	var tag pgconn.CommandTag
	err := cb.Execute(ctx, func() error {
		var innerErr error
		tag, innerErr = ExecWithTimeout(ctx, pool, opts, sql, args...)
		return innerErr
	})

	if errors.Is(err, resilience.ErrCircuitOpen) {
		return tag, fmt.Errorf("database circuit breaker open: %w", err)
	}

	return tag, err
}
