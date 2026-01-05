.PHONY: build test lint install clean snapshot release fmt vet tidy

# Build variables
BINARY_NAME=costdiff
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-s -w -X github.com/hserkanyilmaz/costdiff/cmd.Version=$(VERSION) -X github.com/hserkanyilmaz/costdiff/cmd.Commit=$(COMMIT) -X github.com/hserkanyilmaz/costdiff/cmd.BuildDate=$(BUILD_DATE)"

# Default target
all: build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) main.go

# Build for multiple platforms
build-all:
	@echo "Building for all platforms..."
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 main.go
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy

# Install the binary
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -rf dist/
	rm -f coverage.out coverage.html

# Create a snapshot release (for testing)
snapshot:
	@echo "Creating snapshot release..."
	goreleaser release --snapshot --clean

# Create a release (requires GITHUB_TOKEN)
release:
	@echo "Creating release..."
	goreleaser release --clean

# Run the tool (for development)
run:
	go run main.go $(ARGS)

# Development helpers
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/goreleaser/goreleaser@latest

# Check everything before commit
check: fmt vet lint test
	@echo "All checks passed!"

# Help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  build-all    - Build for all platforms"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  lint         - Run linter"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  tidy         - Tidy dependencies"
	@echo "  install      - Install the binary"
	@echo "  clean        - Clean build artifacts"
	@echo "  snapshot     - Create snapshot release"
	@echo "  release      - Create release"
	@echo "  run          - Run the tool (use ARGS=... for arguments)"
	@echo "  dev-setup    - Install development tools"
	@echo "  check        - Run all checks before commit"

