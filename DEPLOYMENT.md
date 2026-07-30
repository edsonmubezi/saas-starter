# Deployment Guide

## Pre-deployment checklist

- [ ] `JWT_SECRET` set to 64+ random bytes (`openssl rand -hex 64`)
- [ ] `SECRET_KEY` set to exactly 32 bytes (`openssl rand -hex 16`)
- [ ] `DB_PASSWORD` set to a strong unique password
- [ ] `SMTP_*` configured for real email sending
- [ ] `CORS_ALLOWED_ORIGIN` restricted to your frontend domain
- [ ] `ENV=production` in environment
- [ ] `REDIS_ENABLED=true` with a real Redis instance
- [ ] TLS termination configured (Nginx, Ingress, or load balancer)

---

## Docker Compose (single server)

```bash
# 1. Clone and configure
cp .env.example .env
# Fill in all required values in .env

# 2. Start everything
docker compose up -d

# 3. Verify
curl http://localhost:8080/health/detailed
```

### Profiles

```bash
# With observability stack (Grafana + Prometheus + Loki + Tempo)
docker compose --profile observability up -d

# With ClamAV malware scanning
docker compose --profile security up -d
```

---

## Kubernetes

### Prerequisites

- `kubectl` configured for your cluster
- An Ingress controller (nginx recommended)
- `cert-manager` for TLS (optional but recommended)
- A container registry your cluster can pull from

### Deploy

```bash
# 1. Build and push the image
make docker-build DOCKER_REGISTRY=your-registry.io VERSION=v1.0.0
make docker-push  DOCKER_REGISTRY=your-registry.io VERSION=v1.0.0

# 2. Apply namespace
kubectl apply -f ops/k8s/namespace.yaml

# 3. Create secrets (fill in ops/k8s/secret.template.yaml first, then apply)
# WARNING: never commit this file with real values
cp ops/k8s/secret.template.yaml ops/k8s/secret.yaml
# edit ops/k8s/secret.yaml
kubectl apply -f ops/k8s/secret.yaml

# 4. Apply config and workloads
kubectl apply -f ops/k8s/configmap.yaml
kubectl apply -f ops/k8s/deployment.yaml
kubectl apply -f ops/k8s/service.yaml
kubectl apply -f ops/k8s/ingress.yaml
kubectl apply -f ops/k8s/hpa.yaml

# 5. Verify
kubectl rollout status deployment/saas-starter-api -n saas-starter
kubectl get pods -n saas-starter
```

### Update image

```bash
kubectl set image deployment/saas-starter-api \
  api=your-registry.io/saas-starter:v1.1.0 \
  -n saas-starter
kubectl rollout status deployment/saas-starter-api -n saas-starter
```

### Rollback

```bash
kubectl rollout undo deployment/saas-starter-api -n saas-starter
```

---

## Database migrations

Migrations run automatically on server startup. To run manually:

```bash
# Apply all pending
make migrate-up

# Check status
make migrate-status

# Roll back last migration
make migrate-down
```

---

## Post-deploy verification

```bash
BASE_URL=https://api.your-domain.com

# Health
curl $BASE_URL/health
curl $BASE_URL/health/detailed

# Login
curl -X POST $BASE_URL/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Admin@1234"}'

# Swagger
open $BASE_URL/swagger/
```

---

## Secrets management

For production, prefer:
- **AWS Secrets Manager** + External Secrets Operator (K8s)
- **HashiCorp Vault** with agent injection
- **Sealed Secrets** (Bitnami) for GitOps-safe secret storage

Never commit `ops/k8s/secret.yaml` with real values.

---

## Backup

### PostgreSQL

```bash
# Manual snapshot
./scripts/create-base-backup.sh

# Test recovery
./scripts/test-recovery.sh
```

For automated backups, use your cloud provider's managed database snapshots or configure WAL archiving (`scripts/wal-archive.sh`).

Audit events, security events, and application logs are stored in PostgreSQL under the `platform` and `audit` schemas and are covered by the PostgreSQL backup above.
