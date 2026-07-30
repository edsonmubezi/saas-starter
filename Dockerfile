# ---------------- 1) Build stage ----------------
FROM golang:1.23-alpine AS builder

# Environment for reproducible builds
ENV CGO_ENABLED=0 GO111MODULE=on GOTOOLCHAIN=auto

# Install git for modules
RUN apk add --no-cache git

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

# Copy source
COPY . .

# Ensure optional asset dirs exist (avoids COPY errors later)
RUN mkdir -p /app/templates /app/static

# Build binaries with version info
ARG VERSION=dev
ARG COMMIT=unknown
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT}" \
      -o /out/myapp ./cmd/server.go && \
    go build -trimpath -ldflags="-s -w" \
      -o /out/migrate ./cmd/migrate/main.go

# ---------------- 2) Runtime stage ----------------
FROM alpine:3.20

# Runtime dependencies (fonts useful for PDF/image rendering, bash and su-exec for entrypoint)
RUN apk add --no-cache ca-certificates ttf-dejavu fontconfig freetype bash su-exec

# Create non-root user
RUN addgroup -g 10001 app && adduser -D -u 10001 -G app app

WORKDIR /app

# Copy binaries + assets from builder
COPY --from=builder --chown=10001:10001 /out/myapp /app/myapp
COPY --from=builder --chown=10001:10001 /out/migrate /app/migrate
COPY --from=builder --chown=10001:10001 /app/templates /app/templates
COPY --from=builder --chown=10001:10001 /app/static /app/static
COPY --from=builder --chown=10001:10001 /app/db /app/db

# Copy entrypoint script (sed removes Windows CRLF line endings)
COPY scripts/entrypoint.sh /app/scripts/
RUN sed -i 's/\r$//' /app/scripts/*.sh && chmod +x /app/scripts/*.sh

# Ensure app user owns the app directory
RUN chown -R app:app /app

# Run as root initially (entrypoint drops privileges after setup)
USER root

EXPOSE 8080
ENV PORT=8080

ENTRYPOINT ["/app/scripts/entrypoint.sh"]
