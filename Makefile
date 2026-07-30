# === CONFIGURATION ===
CLI_NAME=testgen.exe
SRC_DIR=scripts
TEST_DIR=tests
APP_NAME=saas-starter
DOCKER_REGISTRY ?= your-registry
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Required input:
REPO ?=
BLUEPRINT ?=
SOURCE ?=

# Migration configuration
MIGRATION_NAME ?=

# === APP BUILD ===
build:
	go build -o $(CLI_NAME) ./$(SRC_DIR)

build-server:
	go build -ldflags="-w -s -X main.Version=$(VERSION) -X main.Commit=$(COMMIT)" -o bin/server ./cmd/server.go

# === DOCKER ===
docker-build:
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg COMMIT=$(COMMIT) \
		-t $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION) \
		-t $(DOCKER_REGISTRY)/$(APP_NAME):latest \
		.

docker-push:
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):$(VERSION)
	docker push $(DOCKER_REGISTRY)/$(APP_NAME):latest

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# === CODE QUALITY ===
fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null 2>&1 || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

security-scan:
	@which govulncheck > /dev/null 2>&1 || go install golang.org/x/vuln/cmd/govulncheck@latest
	govulncheck ./...

# === COVERAGE ===
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report written to coverage.html"

# === TEST GENERATION ===
gen: validate-input build
	./$(CLI_NAME) repo --source=$(REPO) --blueprint=$(BLUEPRINT)

# === USECASE TEST GENERATION ===
gen-usecase: validate-usecase build
	./$(CLI_NAME) usecase --source=$(REPO)

# === HANDLER UNIT TEST GENERATION ===
gen-handler: validate-handler-input build
	./$(CLI_NAME) handler --source=$(SOURCE)

# === HANDLER INTEGRATION TEST GENERATION ===
gen-handler-integration: validate-handler-input build
	./$(CLI_NAME) handler-integration --source=$(SOURCE)

# === VALIDATIONS ===
validate-input:
ifeq ($(strip $(REPO)),)
	$(error Missing required REPO=... path to repository file)
endif
ifeq ($(strip $(BLUEPRINT)),)
	$(error Missing required BLUEPRINT=... path to blueprint JSON)
endif

validate-usecase:
ifeq ($(strip $(REPO)),)
	$(error Missing required REPO=... path to usecase file)
endif

validate-handler-input:
ifeq ($(strip $(SOURCE)),)
	$(error Missing required SOURCE=... path to handler file)
endif

# === RUN TESTS ===
test:
	go test ./$(TEST_DIR)/...

testv:
	go test -v ./$(TEST_DIR)/...

clean:
	del /Q $(CLI_NAME)
	del /Q $(TEST_DIR)\*\_test.go

# === DATABASE MIGRATIONS ===
migrate-up:
	go run cmd/migrate/main.go -command=up

migrate-down:
	go run cmd/migrate/main.go -command=down -steps=1

migrate-down-all:
	go run cmd/migrate/main.go -command=down -steps=0

migrate-status:
	go run cmd/migrate/main.go -command=status

migrate-create:
ifeq ($(strip $(MIGRATION_NAME)),)
	$(error Missing required MIGRATION_NAME=... name for migration)
endif
	go run cmd/migrate/main.go -command=create -name="$(MIGRATION_NAME)"

help:
	@echo ""
	@echo "Build:"
	@echo "  make build-server                            - Build the API server binary"
	@echo "  make build                                   - Build the testgen CLI"
	@echo ""
	@echo "Code Quality:"
	@echo "  make fmt                                     - Format Go source files"
	@echo "  make vet                                     - Run go vet"
	@echo "  make lint                                    - Run golangci-lint (requires installation)"
	@echo "  make security-scan                           - Run govulncheck"
	@echo "  make coverage                                - Run tests with coverage report"
	@echo ""
	@echo "Testing:"
	@echo "  make test                                    - Run all tests"
	@echo "  make testv                                   - Run tests verbosely"
	@echo "  make gen REPO=<file> BLUEPRINT=<file>        - Generate repository test"
	@echo "  make gen-usecase REPO=<file>                 - Generate usecase unit test"
	@echo "  make gen-handler SOURCE=<file>               - Generate handler unit test"
	@echo "  make gen-handler-integration SOURCE=<file>   - Generate handler integration test"
	@echo "  make clean                                   - Clean CLI and test files"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-build                            - Build Docker image"
	@echo "  make docker-push                             - Push image to registry"
	@echo "  make docker-up                               - Start all services (docker compose)"
	@echo "  make docker-down                             - Stop all services"
	@echo ""
	@echo "Database Migrations:"
	@echo "  make migrate-up                              - Apply all pending migrations"
	@echo "  make migrate-down                            - Rollback last migration"
	@echo "  make migrate-down-all                        - Rollback all migrations"
	@echo "  make migrate-status                          - Show migration status"
	@echo "  make migrate-create MIGRATION_NAME=<name>    - Create new migration files"
	@echo ""
