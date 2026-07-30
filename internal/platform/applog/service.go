package applog

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// ensurePlatformTables creates the platform schema and tables if they don't exist.
func ensurePlatformTables(db *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	ddl := []string{
		"CREATE SCHEMA IF NOT EXISTS platform",

		`CREATE TABLE IF NOT EXISTS platform.application_logs (
			id          BIGSERIAL PRIMARY KEY,
			timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			level       VARCHAR(10) NOT NULL DEFAULT 'INFO',
			category    VARCHAR(20) NOT NULL DEFAULT 'system',
			message     TEXT NOT NULL,
			service     VARCHAR(50) NOT NULL DEFAULT 'saas-api',
			trace_id    VARCHAR(100),
			span_id     VARCHAR(100),
			tenant_id   BIGINT,
			user_id     BIGINT,
			request_id  VARCHAR(100),
			method      VARCHAR(10),
			path        TEXT,
			status_code INT,
			duration_ns BIGINT,
			ip_address  VARCHAR(45),
			user_agent  TEXT,
			error_type  VARCHAR(255),
			error_msg   TEXT,
			error_stack TEXT,
			error_code  VARCHAR(50),
			fields      JSONB,
			stack       TEXT
		)`,

		`CREATE TABLE IF NOT EXISTS platform.access_logs (
			id              BIGSERIAL PRIMARY KEY,
			timestamp       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			method          VARCHAR(10) NOT NULL,
			path            TEXT NOT NULL,
			query           TEXT,
			status_code     INT NOT NULL,
			duration_ns     BIGINT NOT NULL,
			request_size    BIGINT,
			response_size   BIGINT,
			ip_address      VARCHAR(45) NOT NULL,
			user_agent      TEXT,
			referer         TEXT,
			tenant_id       BIGINT,
			user_id         BIGINT,
			request_id      VARCHAR(100),
			trace_id        VARCHAR(100),
			span_id         VARCHAR(100),
			error           TEXT,
			content_type    VARCHAR(100),
			protocol        VARCHAR(20),
			tls_version     VARCHAR(20),
			geo_location    JSONB
		)`,

		`CREATE TABLE IF NOT EXISTS platform.security_events (
			id                BIGSERIAL PRIMARY KEY,
			timestamp         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			tenant_id         BIGINT,
			severity          VARCHAR(20) NOT NULL DEFAULT 'INFO',
			category          VARCHAR(30) NOT NULL DEFAULT 'authentication',
			event_type        VARCHAR(60) NOT NULL,
			actor_id          BIGINT,
			actor_email       VARCHAR(255),
			ip_address        VARCHAR(45),
			user_agent        TEXT,
			geo_location      JSONB,
			details           JSONB,
			threat_indicators TEXT[],
			alert_sent        BOOLEAN NOT NULL DEFAULT FALSE,
			alert_channels    TEXT[],
			created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)`,
	}

	for _, q := range ddl {
		if _, err := db.Exec(ctx, q); err != nil {
			log.Printf("WARNING [applog] ensurePlatformTables: %v", err)
		}
	}

	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_app_logs_tenant_ts ON platform.application_logs (tenant_id, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_app_logs_level_ts ON platform.application_logs (level, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_app_logs_category_ts ON platform.application_logs (category, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_tenant_ts ON platform.access_logs (tenant_id, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_access_logs_status_ts ON platform.access_logs (status_code, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sec_events_tenant_ts ON platform.security_events (tenant_id, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sec_events_type_ts ON platform.security_events (event_type, timestamp DESC)",
		"CREATE INDEX IF NOT EXISTS idx_sec_events_severity_ts ON platform.security_events (severity, timestamp DESC)",
	}

	for _, q := range indexes {
		if _, err := db.Exec(ctx, q); err != nil {
			log.Printf("WARNING [applog] ensurePlatformTables index: %v", err)
		}
	}

	log.Println("[applog] Platform tables ensured (application_logs, access_logs, security_events)")
}

// Service provides application logging functionality
type Service struct {
	db            *pgxpool.Pool
	repo          *PostgresRepository
	serviceName   string
	bufferSize    int
	logBuffer     chan *ApplicationLog
	accessBuffer  chan *HTTPAccessLog
	flushInterval time.Duration
}

// Config for the logging service
type Config struct {
	ServiceName   string
	BufferSize    int
	FlushInterval time.Duration
}

// NewService creates a new application logging service
func NewService(db *pgxpool.Pool, cfg Config) *Service {
	if cfg.BufferSize == 0 {
		cfg.BufferSize = 1000
	}
	if cfg.FlushInterval == 0 {
		cfg.FlushInterval = 5 * time.Second
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "saas-api"
	}

	var repo *PostgresRepository
	if db != nil {
		ensurePlatformTables(db)
		repo = NewPostgresRepository(db)
	}

	svc := &Service{
		db:            db,
		repo:          repo,
		serviceName:   cfg.ServiceName,
		bufferSize:    cfg.BufferSize,
		logBuffer:     make(chan *ApplicationLog, cfg.BufferSize),
		accessBuffer:  make(chan *HTTPAccessLog, cfg.BufferSize),
		flushInterval: cfg.FlushInterval,
	}

	go svc.flushWorker()
	go svc.accessFlushWorker()

	return svc
}

// LogBuilder provides a fluent interface for building log entries
type LogBuilder struct {
	log *ApplicationLog
	svc *Service
}

// NewLog creates a new log builder
func (s *Service) NewLog() *LogBuilder {
	return &LogBuilder{
		log: &ApplicationLog{
			Timestamp: time.Now().UTC(),
			Service:   s.serviceName,
			Level:     LevelInfo,
			Category:  CategorySystem,
			Fields:    make(map[string]interface{}),
		},
		svc: s,
	}
}

func (b *LogBuilder) Level(level LogLevel) *LogBuilder {
	b.log.Level = level
	return b
}

func (b *LogBuilder) Debug() *LogBuilder {
	b.log.Level = LevelDebug
	return b
}

func (b *LogBuilder) Info() *LogBuilder {
	b.log.Level = LevelInfo
	return b
}

func (b *LogBuilder) Warn() *LogBuilder {
	b.log.Level = LevelWarn
	return b
}

func (b *LogBuilder) Error() *LogBuilder {
	b.log.Level = LevelError
	return b
}

func (b *LogBuilder) Fatal() *LogBuilder {
	b.log.Level = LevelFatal
	return b
}

func (b *LogBuilder) Category(cat LogCategory) *LogBuilder {
	b.log.Category = cat
	return b
}

func (b *LogBuilder) HTTP() *LogBuilder {
	b.log.Category = CategoryHTTP
	return b
}

func (b *LogBuilder) Database() *LogBuilder {
	b.log.Category = CategoryDatabase
	return b
}

func (b *LogBuilder) Message(msg string) *LogBuilder {
	b.log.Message = msg
	return b
}

func (b *LogBuilder) Messagef(format string, args ...interface{}) *LogBuilder {
	b.log.Message = fmt.Sprintf(format, args...)
	return b
}

func (b *LogBuilder) Tenant(tenantID int64) *LogBuilder {
	b.log.TenantID = &tenantID
	return b
}

func (b *LogBuilder) User(userID int64) *LogBuilder {
	b.log.UserID = &userID
	return b
}

func (b *LogBuilder) Request(requestID, method, path string) *LogBuilder {
	b.log.RequestID = requestID
	b.log.Method = method
	b.log.Path = path
	return b
}

func (b *LogBuilder) Trace(traceID, spanID string) *LogBuilder {
	b.log.TraceID = traceID
	b.log.SpanID = spanID
	return b
}

func (b *LogBuilder) Duration(d time.Duration) *LogBuilder {
	b.log.Duration = d
	return b
}

func (b *LogBuilder) Status(code int) *LogBuilder {
	b.log.StatusCode = code
	return b
}

func (b *LogBuilder) IP(ip string) *LogBuilder {
	b.log.IPAddress = ip
	return b
}

func (b *LogBuilder) UserAgent(ua string) *LogBuilder {
	b.log.UserAgent = ua
	return b
}

func (b *LogBuilder) Field(key string, value interface{}) *LogBuilder {
	b.log.Fields[key] = value
	return b
}

func (b *LogBuilder) Fields(fields map[string]interface{}) *LogBuilder {
	for k, v := range fields {
		b.log.Fields[k] = v
	}
	return b
}

func (b *LogBuilder) Err(err error) *LogBuilder {
	if err != nil {
		b.log.Error = &ErrorDetails{
			Type:    fmt.Sprintf("%T", err),
			Message: err.Error(),
		}
		b.log.Stack = captureStackTrace(3)
	}
	return b
}

func (b *LogBuilder) ErrWithCode(err error, code string) *LogBuilder {
	if err != nil {
		b.log.Error = &ErrorDetails{
			Type:    fmt.Sprintf("%T", err),
			Message: err.Error(),
			Code:    code,
		}
		b.log.Stack = captureStackTrace(3)
	}
	return b
}

// Send queues the log for asynchronous writing
func (b *LogBuilder) Send() {
	select {
	case b.svc.logBuffer <- b.log:
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		b.svc.writeLog(ctx, b.log)
	}
}

// SendSync writes the log synchronously
func (b *LogBuilder) SendSync(ctx context.Context) error {
	return b.svc.writeLog(ctx, b.log)
}

// LogHTTPAccess logs an HTTP access entry
func (s *Service) LogHTTPAccess(l *HTTPAccessLog) {
	l.Timestamp = time.Now().UTC()
	select {
	case s.accessBuffer <- l:
	default:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		s.writeAccessLog(ctx, l)
	}
}

func (s *Service) writeLog(ctx context.Context, l *ApplicationLog) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.InsertBatch(ctx, []ApplicationLog{*l})
}

func (s *Service) writeAccessLog(ctx context.Context, l *HTTPAccessLog) error {
	if s.repo == nil {
		return nil
	}
	return s.repo.InsertAccessBatch(ctx, []HTTPAccessLog{*l})
}

func (s *Service) flushWorker() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	batch := make([]*ApplicationLog, 0, 100)

	for {
		select {
		case l := <-s.logBuffer:
			batch = append(batch, l)
			if len(batch) >= 100 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (s *Service) accessFlushWorker() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	batch := make([]*HTTPAccessLog, 0, 100)

	for {
		select {
		case l := <-s.accessBuffer:
			batch = append(batch, l)
			if len(batch) >= 100 {
				s.flushAccessBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flushAccessBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (s *Service) flushBatch(logs []*ApplicationLog) {
	if s.repo == nil || len(logs) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	docs := make([]ApplicationLog, len(logs))
	for i, l := range logs {
		docs[i] = *l
	}
	if err := s.repo.InsertBatch(ctx, docs); err != nil {
		log.Printf("ERROR [applog] Failed to flush application log batch (%d entries): %v", len(docs), err)
	}
}

func (s *Service) flushAccessBatch(logs []*HTTPAccessLog) {
	if s.repo == nil || len(logs) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	docs := make([]HTTPAccessLog, len(logs))
	for i, l := range logs {
		docs[i] = *l
	}
	if err := s.repo.InsertAccessBatch(ctx, docs); err != nil {
		log.Printf("ERROR [applog] Failed to flush access log batch (%d entries): %v", len(docs), err)
	}
}

// GetLogs retrieves application logs with filters and pagination
func (s *Service) GetLogs(ctx context.Context, filters LogFilters, page, pageSize int) ([]ApplicationLog, int64, error) {
	if s.repo == nil {
		return []ApplicationLog{}, 0, nil
	}
	return s.repo.GetLogs(ctx, filters, page, pageSize)
}

// GetAccessLogs retrieves HTTP access logs
func (s *Service) GetAccessLogs(ctx context.Context, filters LogFilters, page, pageSize int) ([]HTTPAccessLog, int64, error) {
	if s.repo == nil {
		return []HTTPAccessLog{}, 0, nil
	}
	accessFilters := AccessLogFilters{
		TenantID:  filters.TenantID,
		UserID:    filters.UserID,
		TraceID:   filters.TraceID,
		RequestID: filters.RequestID,
		Method:    filters.Method,
		Path:      filters.Path,
		FromDate:  filters.FromDate,
		ToDate:    filters.ToDate,
	}
	return s.repo.GetAccessLogs(ctx, accessFilters, page, pageSize)
}

// GetLogByID retrieves a single log by ID
func (s *Service) GetLogByID(ctx context.Context, id int64) (*ApplicationLog, error) {
	if s.repo == nil {
		return nil, nil
	}
	return s.repo.GetLogByID(ctx, id)
}

// GetLogStats returns aggregated log statistics
func (s *Service) GetLogStats(ctx context.Context, tenantID *int64, hours int) (*LogStats, error) {
	if s.repo == nil {
		return &LogStats{
			ByLevel:    make(map[string]int64),
			ByCategory: make(map[string]int64),
		}, nil
	}

	data, err := s.repo.GetLogStats(ctx, tenantID, hours)
	if err != nil {
		return nil, err
	}

	stats := &LogStats{
		ByLevel:       data.ByLevel,
		ByCategory:    data.ByCategory,
		TopErrorPaths: data.TopErrorPaths,
		ErrorTrend:    data.ErrorTrend,
	}

	var errorCount int64
	for level, count := range data.ByLevel {
		stats.TotalLogs += count
		if level == string(LevelError) || level == string(LevelFatal) {
			errorCount += count
		}
	}

	if stats.TotalLogs > 0 {
		stats.ErrorRate = float64(errorCount) / float64(stats.TotalLogs) * 100
	}

	return stats, nil
}

func captureStackTrace(skip int) string {
	var pcs [32]uintptr
	n := runtime.Callers(skip, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var sb strings.Builder
	for {
		frame, more := frames.Next()
		if strings.Contains(frame.File, "runtime/") {
			if !more {
				break
			}
			continue
		}
		sb.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}
	return sb.String()
}

// Convenience methods

func (s *Service) Debug(msg string) {
	s.NewLog().Debug().Message(msg).Send()
}

func (s *Service) Debugf(format string, args ...interface{}) {
	s.NewLog().Debug().Messagef(format, args...).Send()
}

func (s *Service) Info(msg string) {
	s.NewLog().Info().Message(msg).Send()
}

func (s *Service) Infof(format string, args ...interface{}) {
	s.NewLog().Info().Messagef(format, args...).Send()
}

func (s *Service) Warn(msg string) {
	s.NewLog().Warn().Message(msg).Send()
}

func (s *Service) Warnf(format string, args ...interface{}) {
	s.NewLog().Warn().Messagef(format, args...).Send()
}

func (s *Service) ErrorMsg(msg string, err error) {
	s.NewLog().Error().Message(msg).Err(err).Send()
}

func (s *Service) Errorf(format string, args ...interface{}) {
	s.NewLog().Error().Messagef(format, args...).Send()
}
