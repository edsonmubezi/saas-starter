package applog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// PostgresRepository provides PostgreSQL storage for application and access logs.
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// InsertBatch inserts a batch of application logs using a multi-row INSERT.
func (r *PostgresRepository) InsertBatch(ctx context.Context, logs []ApplicationLog) error {
	if len(logs) == 0 {
		return nil
	}

	const cols = 22
	placeholders := make([]string, 0, len(logs))
	args := make([]interface{}, 0, len(logs)*cols)

	for i, l := range logs {
		base := i * cols

		var fieldsJSON []byte
		if len(l.Fields) > 0 {
			fieldsJSON, _ = json.Marshal(l.Fields)
		}

		var errorType, errorMsg, errorStack, errorCode *string
		if l.Error != nil {
			errorType = &l.Error.Type
			errorMsg = &l.Error.Message
			errorStack = &l.Error.StackTrace
			errorCode = &l.Error.Code
		}

		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
			base+11, base+12, base+13, base+14, base+15, base+16, base+17, base+18, base+19, base+20,
			base+21, base+22,
		))

		var fieldsArg interface{}
		if fieldsJSON != nil {
			fieldsArg = string(fieldsJSON)
		}

		args = append(args,
			l.Timestamp,              // 1  timestamp
			string(l.Level),          // 2  level
			string(l.Category),       // 3  category
			l.Message,                // 4  message
			l.Service,                // 5  service
			l.TraceID,                // 6  trace_id
			l.SpanID,                 // 7  span_id
			l.TenantID,               // 8  tenant_id
			l.UserID,                 // 9  user_id
			l.RequestID,              // 10 request_id
			l.Method,                 // 11 method
			l.Path,                   // 12 path
			l.StatusCode,             // 13 status_code
			l.Duration.Nanoseconds(), // 14 duration_ns
			l.IPAddress,              // 15 ip_address
			l.UserAgent,              // 16 user_agent
			errorType,                // 17 error_type
			errorMsg,                 // 18 error_msg
			errorStack,               // 19 error_stack
			errorCode,                // 20 error_code
			fieldsArg,                // 21 fields (JSONB)
			l.Stack,                  // 22 stack
		)
	}

	query := "INSERT INTO platform.application_logs " +
		"(timestamp, level, category, message, service, trace_id, span_id, " +
		"tenant_id, user_id, request_id, method, path, status_code, duration_ns, " +
		"ip_address, user_agent, error_type, error_msg, error_stack, error_code, " +
		"fields, stack) VALUES " + strings.Join(placeholders, ",")

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// InsertAccessBatch inserts a batch of HTTP access logs using a multi-row INSERT.
func (r *PostgresRepository) InsertAccessBatch(ctx context.Context, logs []HTTPAccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	const cols = 21
	placeholders := make([]string, 0, len(logs))
	args := make([]interface{}, 0, len(logs)*cols)

	for i, l := range logs {
		base := i * cols

		var geoJSON interface{}
		if l.GeoLocation != nil {
			b, _ := json.Marshal(l.GeoLocation)
			geoJSON = string(b)
		}

		placeholders = append(placeholders, fmt.Sprintf(
			"($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			base+1, base+2, base+3, base+4, base+5, base+6, base+7, base+8, base+9, base+10,
			base+11, base+12, base+13, base+14, base+15, base+16, base+17, base+18, base+19, base+20,
			base+21,
		))

		args = append(args,
			l.Timestamp,              // 1  timestamp
			l.Method,                 // 2  method
			l.Path,                   // 3  path
			l.Query,                  // 4  query
			l.StatusCode,             // 5  status_code
			l.Duration.Nanoseconds(), // 6  duration_ns
			l.RequestSize,            // 7  request_size
			l.ResponseSize,           // 8  response_size
			l.IPAddress,              // 9  ip_address
			l.UserAgent,              // 10 user_agent
			l.Referer,                // 11 referer
			l.TenantID,               // 12 tenant_id
			l.UserID,                 // 13 user_id
			l.RequestID,              // 14 request_id
			l.TraceID,                // 15 trace_id
			l.SpanID,                 // 16 span_id
			l.Error,                  // 17 error
			l.ContentType,            // 18 content_type
			l.Protocol,               // 19 protocol
			l.TLSVersion,             // 20 tls_version
			geoJSON,                  // 21 geo_location (JSONB)
		)
	}

	query := "INSERT INTO platform.access_logs " +
		"(timestamp, method, path, query, status_code, duration_ns, " +
		"request_size, response_size, ip_address, user_agent, referer, " +
		"tenant_id, user_id, request_id, trace_id, span_id, " +
		"error, content_type, protocol, tls_version, geo_location) " +
		"VALUES " + strings.Join(placeholders, ",")

	_, err := r.db.Exec(ctx, query, args...)
	return err
}

// GetLogs retrieves application logs with filters and pagination.
func (r *PostgresRepository) GetLogs(ctx context.Context, filters LogFilters, page, pageSize int) ([]ApplicationLog, int64, error) {
	where, args := buildAppLogWhere(filters)

	countQuery := "SELECT COUNT(*) FROM platform.application_logs" + where
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %w", err)
	}

	if total == 0 {
		return []ApplicationLog{}, 0, nil
	}

	offset := (page - 1) * pageSize
	nextArg := len(args) + 1
	dataQuery := "SELECT id, timestamp, level, category, message, service, trace_id, span_id, " +
		"tenant_id, user_id, request_id, method, path, status_code, duration_ns, " +
		"ip_address, user_agent, error_type, error_msg, error_stack, error_code, " +
		"fields, stack FROM platform.application_logs" + where +
		fmt.Sprintf(" ORDER BY timestamp DESC LIMIT $%d OFFSET $%d", nextArg, nextArg+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query logs: %w", err)
	}
	defer rows.Close()

	logs, err := scanApplicationLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetAccessLogs retrieves HTTP access logs with filters and pagination.
func (r *PostgresRepository) GetAccessLogs(ctx context.Context, filters AccessLogFilters, page, pageSize int) ([]HTTPAccessLog, int64, error) {
	where, args := buildAccessLogWhere(filters)

	countQuery := "SELECT COUNT(*) FROM platform.access_logs al LEFT JOIN users u ON u.id = al.user_id" + where
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count access logs: %w", err)
	}

	if total == 0 {
		return []HTTPAccessLog{}, 0, nil
	}

	offset := (page - 1) * pageSize
	nextArg := len(args) + 1
	dataQuery := "SELECT al.id, al.timestamp, al.method, al.path, al.query, al.status_code, al.duration_ns, " +
		"al.request_size, al.response_size, al.ip_address, al.user_agent, al.referer, " +
		"al.tenant_id, al.user_id, al.request_id, al.trace_id, al.span_id, " +
		"al.error, al.content_type, al.protocol, al.tls_version, al.geo_location, " +
		"u.email AS user_email " +
		"FROM platform.access_logs al LEFT JOIN users u ON u.id = al.user_id" + where +
		fmt.Sprintf(" ORDER BY al.timestamp DESC LIMIT $%d OFFSET $%d", nextArg, nextArg+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query access logs: %w", err)
	}
	defer rows.Close()

	logs, err := scanAccessLogs(rows)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogByID retrieves a single application log by its ID.
func (r *PostgresRepository) GetLogByID(ctx context.Context, id int64) (*ApplicationLog, error) {
	query := "SELECT id, timestamp, level, category, message, service, trace_id, span_id, " +
		"tenant_id, user_id, request_id, method, path, status_code, duration_ns, " +
		"ip_address, user_agent, error_type, error_msg, error_stack, error_code, " +
		"fields, stack FROM platform.application_logs WHERE id = $1"

	row := r.db.QueryRow(ctx, query, id)

	var (
		l          ApplicationLog
		durationNs int64
		errorType  *string
		errorMsg   *string
		errorStack *string
		errorCode  *string
		fieldsJSON []byte
	)

	err := row.Scan(
		&l.ID, &l.Timestamp, &l.Level, &l.Category, &l.Message, &l.Service,
		&l.TraceID, &l.SpanID, &l.TenantID, &l.UserID, &l.RequestID,
		&l.Method, &l.Path, &l.StatusCode, &durationNs,
		&l.IPAddress, &l.UserAgent, &errorType, &errorMsg, &errorStack, &errorCode,
		&fieldsJSON, &l.Stack,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan log: %w", err)
	}

	l.Duration = time.Duration(durationNs)

	if errorType != nil && *errorType != "" {
		l.Error = &ErrorDetails{Type: *errorType}
		if errorMsg != nil {
			l.Error.Message = *errorMsg
		}
		if errorStack != nil {
			l.Error.StackTrace = *errorStack
		}
		if errorCode != nil {
			l.Error.Code = *errorCode
		}
	}

	if len(fieldsJSON) > 0 {
		l.Fields = make(map[string]interface{})
		_ = json.Unmarshal(fieldsJSON, &l.Fields)
	}

	return &l, nil
}

// GetLogStats returns aggregate log statistics for the given time window.
func (r *PostgresRepository) GetLogStats(ctx context.Context, tenantID *int64, hours int) (*LogStatsData, error) {
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	tenantFilter := ""
	args := []interface{}{since}
	if tenantID != nil {
		tenantFilter = " AND tenant_id = $2"
		args = append(args, *tenantID)
	}

	levelQuery := "SELECT level, COUNT(*) FROM platform.application_logs " +
		"WHERE timestamp >= $1" + tenantFilter + " GROUP BY level"

	levelRows, err := r.db.Query(ctx, levelQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query level stats: %w", err)
	}
	defer levelRows.Close()

	byLevel := make(map[string]int64)
	for levelRows.Next() {
		var level string
		var count int64
		if err := levelRows.Scan(&level, &count); err != nil {
			return nil, fmt.Errorf("failed to scan level stat: %w", err)
		}
		byLevel[level] = count
	}
	if err := levelRows.Err(); err != nil {
		return nil, fmt.Errorf("level rows error: %w", err)
	}

	catQuery := "SELECT category, COUNT(*) FROM platform.application_logs " +
		"WHERE timestamp >= $1" + tenantFilter + " GROUP BY category"

	catRows, err := r.db.Query(ctx, catQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query category stats: %w", err)
	}
	defer catRows.Close()

	byCategory := make(map[string]int64)
	for catRows.Next() {
		var category string
		var count int64
		if err := catRows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category stat: %w", err)
		}
		byCategory[category] = count
	}
	if err := catRows.Err(); err != nil {
		return nil, fmt.Errorf("category rows error: %w", err)
	}

	pathQuery := "SELECT COALESCE(path, ''), COUNT(*) AS cnt FROM platform.application_logs " +
		"WHERE timestamp >= $1 AND level IN ('ERROR', 'FATAL') AND COALESCE(path, '') != ''" +
		tenantFilter + " GROUP BY path ORDER BY cnt DESC LIMIT 5"
	pathRows, err := r.db.Query(ctx, pathQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query top error paths: %w", err)
	}
	defer pathRows.Close()
	var topErrorPaths []PathErrorStat
	for pathRows.Next() {
		var p PathErrorStat
		if err := pathRows.Scan(&p.Path, &p.Count); err != nil {
			return nil, fmt.Errorf("failed to scan path stat: %w", err)
		}
		topErrorPaths = append(topErrorPaths, p)
	}
	if topErrorPaths == nil {
		topErrorPaths = []PathErrorStat{}
	}

	trendTenantJoin := ""
	var trendArgs []interface{}
	if tenantID != nil {
		trendTenantJoin = "AND l.tenant_id = $1"
		trendArgs = []interface{}{*tenantID}
	}
	trendQuery := fmt.Sprintf(`
		WITH hours AS (
			SELECT generate_series(
				date_trunc('hour', NOW() - INTERVAL '23 hours'),
				date_trunc('hour', NOW()),
				'1 hour'::interval
			) AS hr
		)
		SELECT h.hr,
		       COUNT(l.id) FILTER (WHERE l.level IN ('ERROR','FATAL')) AS errors,
		       COUNT(l.id) FILTER (WHERE l.level = 'WARN') AS warns
		FROM hours h
		LEFT JOIN platform.application_logs l
		    ON date_trunc('hour', l.timestamp) = h.hr %s
		GROUP BY h.hr
		ORDER BY h.hr ASC
	`, trendTenantJoin)
	trendRows, err := r.db.Query(ctx, trendQuery, trendArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to query error trend: %w", err)
	}
	defer trendRows.Close()
	var errorTrend []HourlyErrorStat
	for trendRows.Next() {
		var h HourlyErrorStat
		if err := trendRows.Scan(&h.Hour, &h.Errors, &h.Warns); err != nil {
			return nil, fmt.Errorf("failed to scan trend stat: %w", err)
		}
		errorTrend = append(errorTrend, h)
	}
	if errorTrend == nil {
		errorTrend = []HourlyErrorStat{}
	}

	return &LogStatsData{
		ByLevel:       byLevel,
		ByCategory:    byCategory,
		TopErrorPaths: topErrorPaths,
		ErrorTrend:    errorTrend,
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

func buildAppLogWhere(f LogFilters) (string, []interface{}) {
	var clauses []string
	var args []interface{}
	n := 1

	if f.Level != nil {
		clauses = append(clauses, fmt.Sprintf("level = $%d", n))
		args = append(args, string(*f.Level))
		n++
	}
	if f.Category != nil {
		clauses = append(clauses, fmt.Sprintf("category = $%d", n))
		args = append(args, string(*f.Category))
		n++
	}
	if f.TenantID != nil {
		clauses = append(clauses, fmt.Sprintf("tenant_id = $%d", n))
		args = append(args, *f.TenantID)
		n++
	}
	if f.UserID != nil {
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", n))
		args = append(args, *f.UserID)
		n++
	}
	if f.TraceID != nil {
		clauses = append(clauses, fmt.Sprintf("trace_id = $%d", n))
		args = append(args, *f.TraceID)
		n++
	}
	if f.RequestID != nil {
		clauses = append(clauses, fmt.Sprintf("request_id = $%d", n))
		args = append(args, *f.RequestID)
		n++
	}
	if f.Search != nil && *f.Search != "" {
		clauses = append(clauses, fmt.Sprintf("message ILIKE $%d", n))
		args = append(args, "%"+*f.Search+"%")
		n++
	}
	if f.FromDate != nil {
		clauses = append(clauses, fmt.Sprintf("timestamp >= $%d", n))
		args = append(args, *f.FromDate)
		n++
	}
	if f.ToDate != nil {
		clauses = append(clauses, fmt.Sprintf("timestamp <= $%d", n))
		args = append(args, *f.ToDate)
		n++
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func buildAccessLogWhere(f AccessLogFilters) (string, []interface{}) {
	var clauses []string
	var args []interface{}
	n := 1

	if f.TenantID != nil {
		clauses = append(clauses, fmt.Sprintf("al.tenant_id = $%d", n))
		args = append(args, *f.TenantID)
		n++
	}
	if f.UserID != nil {
		clauses = append(clauses, fmt.Sprintf("al.user_id = $%d", n))
		args = append(args, *f.UserID)
		n++
	}
	if f.TraceID != nil {
		clauses = append(clauses, fmt.Sprintf("al.trace_id = $%d", n))
		args = append(args, *f.TraceID)
		n++
	}
	if f.RequestID != nil {
		clauses = append(clauses, fmt.Sprintf("al.request_id = $%d", n))
		args = append(args, *f.RequestID)
		n++
	}
	if f.Method != nil {
		clauses = append(clauses, fmt.Sprintf("al.method = $%d", n))
		args = append(args, *f.Method)
		n++
	}
	if f.Path != nil {
		clauses = append(clauses, fmt.Sprintf("al.path ILIKE $%d", n))
		args = append(args, "%"+*f.Path+"%")
		n++
	}
	if f.FromDate != nil {
		clauses = append(clauses, fmt.Sprintf("al.timestamp >= $%d", n))
		args = append(args, *f.FromDate)
		n++
	}
	if f.ToDate != nil {
		clauses = append(clauses, fmt.Sprintf("al.timestamp <= $%d", n))
		args = append(args, *f.ToDate)
		n++
	}

	if len(clauses) == 0 {
		return "", nil
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

func scanApplicationLogs(rows pgx.Rows) ([]ApplicationLog, error) {
	var logs []ApplicationLog

	for rows.Next() {
		var (
			l          ApplicationLog
			durationNs int64
			errorType  *string
			errorMsg   *string
			errorStack *string
			errorCode  *string
			fieldsJSON []byte
		)

		err := rows.Scan(
			&l.ID, &l.Timestamp, &l.Level, &l.Category, &l.Message, &l.Service,
			&l.TraceID, &l.SpanID, &l.TenantID, &l.UserID, &l.RequestID,
			&l.Method, &l.Path, &l.StatusCode, &durationNs,
			&l.IPAddress, &l.UserAgent, &errorType, &errorMsg, &errorStack, &errorCode,
			&fieldsJSON, &l.Stack,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan application log row: %w", err)
		}

		l.Duration = time.Duration(durationNs)

		if errorType != nil && *errorType != "" {
			l.Error = &ErrorDetails{Type: *errorType}
			if errorMsg != nil {
				l.Error.Message = *errorMsg
			}
			if errorStack != nil {
				l.Error.StackTrace = *errorStack
			}
			if errorCode != nil {
				l.Error.Code = *errorCode
			}
		}

		if len(fieldsJSON) > 0 {
			l.Fields = make(map[string]interface{})
			_ = json.Unmarshal(fieldsJSON, &l.Fields)
		}

		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("application log rows iteration error: %w", err)
	}

	return logs, nil
}

func scanAccessLogs(rows pgx.Rows) ([]HTTPAccessLog, error) {
	var logs []HTTPAccessLog

	for rows.Next() {
		var (
			l          HTTPAccessLog
			durationNs int64
			geoJSON    []byte
		)

		err := rows.Scan(
			&l.ID, &l.Timestamp, &l.Method, &l.Path, &l.Query,
			&l.StatusCode, &durationNs, &l.RequestSize, &l.ResponseSize,
			&l.IPAddress, &l.UserAgent, &l.Referer,
			&l.TenantID, &l.UserID, &l.RequestID, &l.TraceID, &l.SpanID,
			&l.Error, &l.ContentType, &l.Protocol, &l.TLSVersion,
			&geoJSON, &l.UserEmail,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan access log row: %w", err)
		}

		l.Duration = time.Duration(durationNs)

		if len(geoJSON) > 0 {
			l.GeoLocation = &GeoLocation{}
			_ = json.Unmarshal(geoJSON, l.GeoLocation)
		}

		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("access log rows iteration error: %w", err)
	}

	return logs, nil
}
