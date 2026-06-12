# MarketPay - Windows PowerShell build script
# Usage: .\scripts\make.ps1 <command>
# Example: .\scripts\make.ps1 quickstart

param(
    [Parameter(Position=0)]
    [string]$Command = "help"
)

$ErrorActionPreference = "Stop"

function Build {
    Write-Host "Building binaries..." -ForegroundColor Cyan
    New-Item -ItemType Directory -Force -Path bin | Out-Null
    go build -ldflags="-w -s" -o bin/api.exe     ./cmd/api
    go build -ldflags="-w -s" -o bin/worker.exe  ./cmd/worker
    go build -ldflags="-w -s" -o bin/migrate.exe ./cmd/migrate
    Write-Host "Build complete -> bin/" -ForegroundColor Green
}

function RunApi {
    Build
    ./bin/api.exe -config configs/config.yaml
}

function Test-Unit {
    Write-Host "Running unit tests..." -ForegroundColor Cyan
    go test ./internal/... ./pkg/... -v -count=1 -timeout 60s
}

function Test-Cover {
    go test ./internal/... ./pkg/... -coverprofile=coverage.out -covermode=atomic
    go tool cover -html=coverage.out -o coverage.html
    Write-Host "Coverage report: coverage.html" -ForegroundColor Green
    Start-Process coverage.html
}

function Docker-Up {
    Write-Host "Starting Docker Compose..." -ForegroundColor Cyan
    docker compose up --build -d
    Write-Host ""
    Write-Host "Services started:" -ForegroundColor Green
    Write-Host "  Frontend -> http://localhost:3000"
    Write-Host "  API      -> http://localhost:8080"
    Write-Host "  Swagger  -> http://localhost:8080/swagger/index.html"
    Write-Host "  Health   -> http://localhost:8080/health"
    Write-Host "  Postgres -> localhost:5432"
    Write-Host "  Redis    -> localhost:6379"
    Write-Host ""
    Write-Host "Default credentials:"
    Write-Host "  superadmin@marketpay.sl / password"
}

function Docker-Down {
    docker compose down
}

function Docker-Logs {
    docker compose logs -f api worker frontend
}

function Docker-Clean {
    docker compose down -v --remove-orphans
}

function Migrate-Up {
    Build
    ./bin/migrate.exe -config configs/config.yaml -direction up
}

function Migrate-Down {
    Build
    ./bin/migrate.exe -config configs/config.yaml -direction down
}

function Tidy {
    Write-Host "Running go mod tidy..." -ForegroundColor Cyan
    go mod tidy
    Write-Host "Done." -ForegroundColor Green
}

function Clean {
    Remove-Item -Recurse -Force -ErrorAction SilentlyContinue bin/, coverage.out, coverage.html
    Write-Host "Cleaned." -ForegroundColor Green
}

function Quickstart {
    Write-Host "========================================" -ForegroundColor Yellow
    Write-Host "   MarketPay Quick Start (Windows)     " -ForegroundColor Yellow
    Write-Host "========================================" -ForegroundColor Yellow
    Docker-Up
}

function Show-Help {
    Write-Host ""
    Write-Host "MarketPay PowerShell Build Script" -ForegroundColor Yellow
    Write-Host "Usage: .\scripts\make.ps1 <command>" -ForegroundColor White
    Write-Host ""
    Write-Host "Commands:" -ForegroundColor Cyan
    Write-Host "  quickstart    Start everything via Docker Compose (recommended)"
    Write-Host "  build         Build all Go binaries"
    Write-Host "  run           Build and run the API locally"
    Write-Host "  test          Run unit tests"
    Write-Host "  test-cover    Run tests with HTML coverage report"
    Write-Host "  docker-up     docker compose up --build"
    Write-Host "  docker-down   docker compose down"
    Write-Host "  docker-logs   Follow API + worker logs"
    Write-Host "  docker-clean  Remove containers and volumes"
    Write-Host "  migrate-up    Run database migrations"
    Write-Host "  migrate-down  Rollback database migrations"
    Write-Host "  tidy          go mod tidy"
    Write-Host "  clean         Remove build artifacts"
    Write-Host ""
    Write-Host "Prerequisites:" -ForegroundColor Cyan
    Write-Host "  - Docker Desktop (https://www.docker.com/products/docker-desktop)"
    Write-Host "  - Go 1.24+      (https://go.dev/dl/)"
    Write-Host ""
}

switch ($Command.ToLower()) {
    "quickstart"   { Quickstart }
    "build"        { Build }
    "run"          { RunApi }
    "test"         { Test-Unit }
    "test-unit"    { Test-Unit }
    "test-cover"   { Test-Cover }
    "docker-up"    { Docker-Up }
    "docker-down"  { Docker-Down }
    "docker-logs"  { Docker-Logs }
    "docker-clean" { Docker-Clean }
    "migrate-up"   { Migrate-Up }
    "migrate-down" { Migrate-Down }
    "tidy"         { Tidy }
    "clean"        { Clean }
    default        { Show-Help }
}
