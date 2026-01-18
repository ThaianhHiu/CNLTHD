# Makefile for URL Shortener

.PHONY: help build run test clean docker-up docker-down docker-logs

# Default target
help:
	@echo "URL Shortener - Available Commands:"
	@echo ""
	@echo "  make build       - Build the application"
	@echo "  make run         - Run the application locally"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make docker-up   - Start all services with Docker Compose"
	@echo "  make docker-down - Stop all Docker services"
	@echo "  make docker-logs - View Docker logs"
	@echo "  make deps        - Download dependencies"
	@echo ""

# Download dependencies
deps:
	go mod download
	go mod tidy

# Build the application
build:
	CGO_ENABLED=0 go build -o bin/url-shortener .

# Run locally
run:
	go run main.go

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Docker commands
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v --rmi all

# Development with hot reload (requires air)
dev:
	air

# Lint code
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
