.PHONY: all build run test test-unit test-cover lint fmt vet \
        docker-build docker-up docker-down migrate-up migrate-down \
        swagger clean deps tidy quickstart

# ── Variables ────────────────────────────────────────────────────────────────
APP_NAME    := marketpay
MODULE      := github.com/marketpay/backend
CONFIG      := configs/config.yaml
DOCKER_TAG  := latest
GOFLAGS     := -ldflags="-w -s"
GOFMT_FILES := $(shell find . -name "*.go" -not -path "./vendor/*")

# ── Default ──────────────────────────────────────────────────────────────────
all: fmt vet build

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	@echo "Building binaries..."
	go build $(GOFLAGS) -o bin/api     ./cmd/api
	go build $(GOFLAGS) -o bin/worker  ./cmd/worker
	go build $(GOFLAGS) -o bin/migrate ./cmd/migrate
	@echo "Build complete → bin/"

# ── Run (local, requires running postgres + redis) ────────────────────────────
run: build
	./bin/api -config $(CONFIG)

run-worker: build
	./bin/worker -config $(CONFIG)

# ── Tests ─────────────────────────────────────────────────────────────────────
test: test-unit

test-unit:
	@echo "Running unit tests..."
	go test ./internal/... ./pkg/... -v -count=1 -timeout 60s

test-cover:
	@echo "Running tests with coverage..."
	go test ./internal/... ./pkg/... -coverprofile=coverage.out -covermode=atomic
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	go test ./... -race -count=1

# ── Code quality ──────────────────────────────────────────────────────────────
fmt:
	gofmt -w $(GOFMT_FILES)

vet:
	go vet ./...

lint:
	golangci-lint run ./...

tidy:
	go mod tidy

deps:
	go mod download

# ── Docker ────────────────────────────────────────────────────────────────────
docker-build:
	docker build -f deployments/Dockerfile -t $(APP_NAME):$(DOCKER_TAG) .

docker-up:
	docker compose up --build -d
	@echo "Services starting. API will be available at http://localhost:8080"
	@echo "Frontend:   http://localhost:3000"
	@echo "Swagger UI: http://localhost:8080/swagger/index.html"

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api worker frontend

docker-clean:
	docker compose down -v --remove-orphans

# ── Migrations ────────────────────────────────────────────────────────────────
migrate-up: build
	./bin/migrate -config $(CONFIG) -direction up

migrate-down: build
	./bin/migrate -config $(CONFIG) -direction down

migrate-status:
	@echo "Checking migration status..."
	docker compose exec postgres psql -U marketpay -d marketpay \
	  -c "SELECT version, dirty FROM schema_migrations ORDER BY version;"

# ── Swagger ───────────────────────────────────────────────────────────────────
swagger:
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/api/main.go -o docs/ --parseDependency --parseInternal
	@echo "Swagger docs generated → docs/"

# ── Proto (requires protoc + plugins) ────────────────────────────────────────
proto:
	@which protoc > /dev/null || (echo "protoc not found. Install from https://grpc.io/docs/protoc-installation/" && exit 1)
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       proto/loan/loan.proto \
	       proto/vendor/vendor.proto \
	       proto/payment/payment.proto
	@echo "Proto files compiled."

# ── Clean ─────────────────────────────────────────────────────────────────────
clean:
	rm -rf bin/ docs/ coverage.out coverage.html

# ── Quick start ───────────────────────────────────────────────────────────────
quickstart:
	@echo "╔══════════════════════════════════════╗"
	@echo "║      MarketPay Quick Start           ║"
	@echo "╚══════════════════════════════════════╝"
	docker compose up --build -d
	@echo ""
	@echo "✓ Postgres  → localhost:5432"
	@echo "✓ Redis     → localhost:6379"
	@echo "✓ API       → http://localhost:8080"
	@echo "✓ Frontend  → http://localhost:3000"
	@echo "✓ Swagger   → http://localhost:8080/swagger/index.html"
	@echo ""
	@echo "Default credentials:"
	@echo "  Email:    superadmin@marketpay.sl"
	@echo "  Password: password"
