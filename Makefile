.PHONY: help build run test clean deps fmt lint tidy

help:
	@echo "MarketPay USSD Service - Available Commands"
	@echo ""
	@echo "  make build       - Build the MarketPay USSD service"
	@echo "  make run         - Run the MarketPay USSD service"
	@echo "  make test        - Run all tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make tidy        - Tidy go.mod and go.sum"
	@echo "  make fmt         - Format code"
	@echo "  make lint        - Run linter"
	@echo ""

build: deps
	@echo "Building MarketPay USSD service..."
	go build -o bin/marketpay-ussd main.go examples.go
	@echo "Build complete: bin/marketpay-ussd"

run: build
	@echo "Running MarketPay USSD service..."
	./bin/marketpay-ussd

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	go clean

deps:
	@echo "Downloading dependencies..."
	go mod download

tidy:
	@echo "Tidying go.mod..."
	go mod tidy

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Running linter..."
	golangci-lint run ./...

examples: build
	@echo "Running examples..."
	./bin/marketpay-ussd --examples
