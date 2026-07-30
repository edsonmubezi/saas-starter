# EkklesiaHR Observability Stack

## Overview

This directory contains configuration for the observability and audit infrastructure:

- **Loki**: Log aggregation and storage
- **Tempo**: Distributed tracing backend
- **Grafana**: Dashboards and visualization
- **Fluent Bit**: Log shipping and forwarding

## Quick Start

### Local Development

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f app

# Access Grafana
open http://localhost:3000
# Username: admin
# Password: admin
```

### Production Deployment

```bash
# Start production stack
docker-compose -f docker-compose.prod.yml up -d

# Check service health
docker-compose -f docker-compose.prod.yml ps

# View application logs in Grafana
# Navigate to Explore → Loki
```

## Services

### Grafana (Port 3000)
- **URL**: http://localhost:3000
- **Default Credentials**: admin / admin
- **Dashboards**: Pre-provisioned in `/grafana/dashboards/`
- **Datasources**: Auto-configured (Loki, Tempo, PostgreSQL)

### Loki (Port 3100)
- **URL**: http://localhost:3100
- **Retention**: 14 days for operational logs
- **Config**: `loki-config.yaml`

### Tempo (Port 3200, 4317, 4318)
- **HTTP**: http://localhost:3200
- **OTLP gRPC**: localhost:4317
- **OTLP HTTP**: localhost:4318
- **Retention**: 7 days for traces
- **Config**: `tempo-config.yaml`

### Fluent Bit
- **Collects**: Docker container logs
- **Forwards to**: Loki
- **Config**: `fluent-bit/fluent-bit.conf`

## Querying Logs

### LogQL Examples (Loki)

```logql
# All logs for a specific tenant
{job="ekklesiahr"} | json | tenant_id="123"

# Failed HTTP requests
{job="ekklesiahr"} | json | status >= 400

# Slow requests (>1 second)
{job="ekklesiahr"} | json | latency_ms > 1000

# Specific user's actions
{job="ekklesiahr"} | json | user_id="456"

# Errors in the last hour
{job="ekklesiahr"} | json | level="error" | __timestamp__ > now() - 1h

# Search for specific text
{job="ekklesiahr"} |= "database connection"

# Aggregate error rate
rate({job="ekklesiahr"} | json | level="error"[5m])
```

### TraceQL Examples (Tempo)

```traceql
# Find slow operations
{ duration > 1s }

# Find operations with errors
{ status = error }

# Find specific service
{ service.name = "ekklesiahr-api" }

# Complex query
{ service.name = "ekklesiahr-api" && duration > 500ms && status = error }
```

## Querying Audit Events

### PostgreSQL Queries

```sql
-- Recent high-severity events
SELECT
    occurred_at,
    actor_id,
    action,
    resource_type,
    resource_id,
    severity
FROM audit.events
WHERE severity IN ('HIGH', 'CRITICAL')
ORDER BY occurred_at DESC
LIMIT 50;

-- User activity timeline
SELECT
    occurred_at,
    action,
    resource_type,
    resource_id,
    ip
FROM audit.events
WHERE actor_id = 123
  AND tenant_id = 456
ORDER BY occurred_at DESC;

-- Failed login attempts by IP
SELECT
    ip,
    COUNT(*) as attempt_count,
    MAX(occurred_at) as last_attempt
FROM audit.events
WHERE action = 'LOGIN_FAILED'
  AND occurred_at > NOW() - INTERVAL '24 hours'
GROUP BY ip
HAVING COUNT(*) >= 5
ORDER BY attempt_count DESC;

-- Role modifications
SELECT
    occurred_at,
    actor_id,
    after_json->>'role_name' as role_granted,
    resource_id as user_id
FROM audit.events
WHERE action IN ('ROLE_GRANTED', 'ROLE_REVOKED')
  AND tenant_id = 456
ORDER BY occurred_at DESC;

-- Export activity
SELECT
    occurred_at,
    actor_id,
    action,
    after_json->>'record_count' as records_exported
FROM audit.events
WHERE action LIKE 'EXPORT_%'
  AND tenant_id = 456
ORDER BY occurred_at DESC;
```

## Monitoring & Alerts

### Pre-configured Alerts

1. **Failed Login Attempts**: ≥5 failures from same IP in 10 minutes
2. **Role Escalation**: Any ROLE_GRANTED event with severity HIGH
3. **Mass Delete**: DELETE operations affecting >100 records
4. **Export Activity**: Any EXPORT action for sensitive data
5. **System Errors**: Error rate >1% over 5 minutes

### Alert Destinations

Configure alert notifications in Grafana:
1. Navigate to **Alerting** → **Contact points**
2. Add webhook/email/Slack integration
3. Configure routing rules

## Maintenance

### Log Retention

**Operational Logs (Loki)**:
- Retention: 14 days
- Auto-cleanup: Enabled
- Config: `loki-config.yaml` → `limits_config.retention_period`

**Traces (Tempo)**:
- Retention: 7 days
- Auto-cleanup: Enabled
- Config: `tempo-config.yaml` → `compactor.compaction.block_retention`

**Audit Events (PostgreSQL)**:
- Hot storage: 90 days (indexed)
- Archive: 7 years (S3 WORM)
- Managed by: audit archiver service

### Disk Space Management

```bash
# Check disk usage
docker exec ekklesia-loki du -sh /loki/chunks
docker exec ekklesia-tempo du -sh /var/tempo/traces

# Clean old data (if needed)
docker-compose down
docker volume prune
docker-compose up -d
```

### Backup

```bash
# Backup Grafana dashboards
docker exec ekklesia-grafana tar czf - /var/lib/grafana/dashboards > grafana-dashboards-backup.tar.gz

# Backup Loki data
docker exec ekklesia-loki tar czf - /loki/chunks > loki-backup.tar.gz

# Restore
docker exec -i ekklesia-grafana tar xzf - -C / < grafana-dashboards-backup.tar.gz
```

## Troubleshooting

### Logs Not Appearing in Grafana

1. Check Fluent Bit is running:
   ```bash
   docker-compose ps fluent-bit
   docker-compose logs fluent-bit
   ```

2. Check Loki is receiving logs:
   ```bash
   curl http://localhost:3100/ready
   curl 'http://localhost:3100/loki/api/v1/query?query={job="ekklesiahr"}'
   ```

3. Check Grafana datasource:
   - Go to **Configuration** → **Data sources** → **Loki**
   - Click **Save & test**

### Traces Not Appearing

1. Check Tempo is running:
   ```bash
   docker-compose ps tempo
   curl http://localhost:3200/ready
   ```

2. Check OTLP endpoint is accessible:
   ```bash
   telnet tempo 4317
   ```

3. Check application environment:
   ```bash
   echo $OTEL_ENABLED
   echo $OTEL_EXPORTER_OTLP_ENDPOINT
   ```

### High Disk Usage

```bash
# Check volume sizes
docker system df -v

# Clean up unused data
docker volume prune

# Reduce retention periods in configs
# Edit loki-config.yaml and tempo-config.yaml
```

### Slow Queries

1. Add indexes to audit.events:
   ```sql
   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_audit_tenant_action
   ON audit.events(tenant_id, action, occurred_at DESC);
   ```

2. Use query filters:
   - Always include `tenant_id`
   - Limit time ranges
   - Use indexes (check `EXPLAIN`)

## Environment Variables

### Application

```bash
# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_SAMPLING=true

# Tracing
OTEL_ENABLED=true
OTEL_EXPORTER_OTLP_ENDPOINT=tempo:4317
OTEL_SERVICE_NAME=ekklesiahr-api

# Audit
AUDIT_WORKER_ENABLED=true
AUDIT_ARCHIVE_ENABLED=false
```

### Grafana

```bash
GF_SECURITY_ADMIN_PASSWORD=admin
GF_USERS_ALLOW_SIGN_UP=false
GF_AUTH_ANONYMOUS_ENABLED=false
```

## Performance Tuning

### Loki

```yaml
# loki-config.yaml
limits_config:
  ingestion_rate_mb: 20  # Increase for high-volume
  max_query_parallelism: 64  # Increase for faster queries
```

### Tempo

```yaml
# tempo-config.yaml
ingester:
  max_block_bytes: 5_000_000  # Larger blocks for more traces
storage:
  trace:
    pool:
      max_workers: 200  # More workers for throughput
```

## Security

### Access Control

- **Grafana**: Only authenticated users
- **Loki API**: Internal network only
- **Tempo API**: Internal network only
- **Audit database**: Read-only role for Grafana

### Audit Log Access

Every query to `audit.events` should be logged:

```sql
-- Add audit-the-audit trigger
-- See migrations/100_create_audit_schema.sql
```

## Support

- **Documentation**: `docs/LOGGING_IMPLEMENTATION.md`
- **Issues**: Check application logs first
- **Dashboards**: Import from `/grafana/dashboards/`

---

**Last Updated**: 2025-10-20
**Maintainer**: DevOps Team
