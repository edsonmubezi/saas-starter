# SaaS Starter

A production-ready multitenant SaaS backend in Go. Drop this into any project and own a fully-featured authentication, RBAC, audit, and organization management layer from day one.

## What's included

| Feature | Details |
|---|---|
| **Multitenancy** | Full data isolation per org via `organization_id` on every query |
| **Auth** | JWT + refresh tokens, 2FA (TOTP / Email / SMS), account lockout |
| **RBAC** | `admin.*` (SuperAdmin) and `tenant.*` (OrgAdmin) permission scopes |
| **Audit trail** | Immutable events with SHA-256 integrity, stored in PostgreSQL |
| **Alerting** | Slack, Teams, Email, SMS, HMAC-signed webhooks per org |
| **AI / Chat** | Per-org OpenAI key, knowledge base with pgvector embeddings |
| **Email** | Per-org SMTP config with `.env` fallback, branded templates |
| **File storage** | Local (dev) or AWS S3 (prod), per-tenant isolation |
| **Import/Export** | Generic data import with column mapping and draft preview |
| **Observability** | OpenTelemetry tracing, PostgreSQL app/access/security logs, Grafana dashboards |
| **Security** | CSP, HSTS, rate limiting, body size limits, input sanitization |

## Tech stack

- **Go 1.23** · Gorilla Mux
- **PostgreSQL 15** (+ pgvector) · **Redis 7**
- **Docker Compose** for local dev, **Kubernetes** manifests in `ops/k8s/`

---

## Quick start (Docker Compose)

```bash
cp .env.example .env        # fill in required values (see below)
docker compose up -d        # starts postgres, redis, api
```

The API will be available at `http://localhost:8080`.

Default login after first run:
- **Email:** `admin@example.com`
- **Password:** `Admin@1234`

> Change the default credentials immediately after first login.

---

## Environment variables

### Required

| Variable | Description |
|---|---|
| `DB_HOST` / `DB_PORT` / `DB_USER` / `DB_PASSWORD` / `DB_NAME` | PostgreSQL connection |
| `JWT_SECRET` | Minimum 64 random bytes |
| `SECRET_KEY` | Exactly 32 bytes (AES-256 key) |

### Common optional

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP listen port |
| `ENV` | `development` | `development` or `production` |
| `REDIS_ENABLED` | `false` | Enable Redis caching & rate limiting |
| `STORAGE_TYPE` | `local` | `local` or `s3` |
| `SMTP_ENABLED` | `false` | Enable real email sending |
| `CORS_ALLOWED_ORIGIN` | `*` | Restrict CORS in production |

See `.env.example` for the full list.

---

## API versioning

All API routes are prefixed with `/api/v1/`. Example:

```
POST /api/v1/login
GET  /api/v1/users/me
GET  /api/v1/admin/organizations
GET  /api/v1/org/org-users
```

Health checks are unversioned: `GET /health`, `GET /health/detailed`.

Swagger UI is at `http://localhost:8080/swagger/`.

---

## Project structure

```
├── api/
│   ├── handler/        # HTTP handlers (22 handlers)
│   ├── middleware/     # Auth, RBAC, rate limiting, CSRF, body limit, etc.
│   └── router/         # Route registration (Super/, Orgs/)
├── cmd/                # Entrypoints (server, migrate, seed)
├── db/migrations/      # SQL migration files (up/down)
├── docs/               # Swagger docs (auto-generated)
├── frontend/           # React + Vite SPA
├── internal/
│   ├── auth/           # Login attempts, refresh tokens, TOTP, password reset
│   ├── organization/   # Org lifecycle + branding
│   ├── permission/     # Permission entity and repository
│   ├── role/           # Role management
│   ├── seeder/         # Auto-seeding on startup
│   ├── user/           # User management
│   └── platform/       # Audit, alerting, app logs, security events, email, tracing
├── ops/
│   ├── k8s/            # Kubernetes manifests
│   └── grafana/        # Grafana dashboards and provisioning
└── pkg/
    ├── apierror/       # Structured error responses
    ├── auth/           # JWT, token blacklist, fingerprints
    ├── cache/          # Redis caching
    ├── featureflags/   # Feature flags for microservices migration
    └── resilience/     # Circuit breaker
```

---

## Make targets

```bash
make build-server         # Build the API binary
make fmt                  # Format Go code
make vet                  # Run go vet
make lint                 # Run golangci-lint
make security-scan        # Run govulncheck
make test                 # Run all tests
make coverage             # Run tests + generate HTML coverage report
make docker-build         # Build Docker image
make docker-push          # Push to registry
make migrate-up           # Apply pending DB migrations
make migrate-create MIGRATION_NAME=add_foo  # Create a new migration
```

---

## Adding your domain

1. Create `internal/yourfeature/` with `entity.go`, `repository.go`, `usecase.go`
2. Register the use case in `internal/container/container.go`
3. Add handler in `api/handler/yourfeature_handler.go`
4. Add routes in `api/router/Orgs/` or `api/router/Super/`
5. Add permissions to `internal/seeder/permissions.go`
6. Create migrations in `db/migrations/`

All queries must filter by `organization_id` to maintain tenant isolation.
