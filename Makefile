.PHONY: test test-unit test-integration test-coverage build clean lint fmt vet

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Binary name
BINARY_NAME=teeny-orb
BINARY_PATH=./cmd/teeny-orb

# Test parameters
TEST_PACKAGES=./...
TEST_TIMEOUT=30s
COVERAGE_FILE=coverage.out

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(BINARY_PATH)

# Run all tests
test: test-unit

# Run unit tests
test-unit:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) $(TEST_PACKAGES)

# Run unit tests with short flag (skip integration tests)
test-short:
	$(GOTEST) -v -short -timeout $(TEST_TIMEOUT) $(TEST_PACKAGES)

# Run integration tests
test-integration:
	$(GOTEST) -v -timeout 2m -tags=integration $(TEST_PACKAGES)

# Run tests with coverage
test-coverage:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -coverprofile=$(COVERAGE_FILE) $(TEST_PACKAGES)
	$(GOCMD) tool cover -html=$(COVERAGE_FILE)

# Run tests with race detection
test-race:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -race $(TEST_PACKAGES)

# Run benchmarks
bench:
	$(GOTEST) -v -timeout $(TEST_TIMEOUT) -bench=. -benchmem $(TEST_PACKAGES)

# Format code
fmt:
	$(GOFMT) ./...

# Vet code
vet:
	$(GOVET) $(TEST_PACKAGES)

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(COVERAGE_FILE)

# Install dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Development helpers
dev-setup: deps
	$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run all quality checks
check: fmt vet test-short

# Quick development cycle
dev: fmt vet test-short build

# Full CI pipeline
ci: fmt vet test-coverage lint

help:
	@echo "Available targets:"
	@echo "  build        - Build the application"
	@echo "  test         - Run all tests"
	@echo "  test-unit    - Run unit tests"
	@echo "  test-short   - Run unit tests (skip integration)"
	@echo "  test-coverage- Run tests with coverage report"
	@echo "  test-race    - Run tests with race detection"
	@echo "  bench        - Run benchmarks"
	@echo "  fmt          - Format code"
	@echo "  vet          - Vet code"
	@echo "  lint         - Run linter"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  dev-setup    - Setup development environment"
	@echo "  check        - Run quick quality checks"
	@echo "  dev          - Development cycle (fmt, vet, test, build)"
	@echo "  ci           - Full CI pipeline"
