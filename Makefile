.PHONY: test test-unit test-integration test-race test-coverage bench lint fmt vet clean build examples help

# Default target
help:
	@echo "Available targets:"
	@echo "  test              - Run all tests"
	@echo "  test-unit         - Run unit tests only"
	@echo "  test-integration  - Run integration tests (requires internet)"
	@echo "  test-race         - Run tests with race detector"
	@echo "  test-coverage     - Run tests with coverage report"
	@echo "  bench             - Run benchmarks"
	@echo "  lint              - Run linting tools"
	@echo "  fmt               - Format code"
	@echo "  vet               - Run go vet"
	@echo "  clean             - Clean build artifacts"
	@echo "  build             - Build examples"
	@echo "  examples          - Run example programs"

# Test targets
test: test-unit test-integration

test-unit:
	go test -v ./...

test-integration:
	TELEGRAPH_INTEGRATION_TEST=1 go test -v ./examples/...

test-race:
	go test -race ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Benchmark
bench:
	go test -bench=. -benchmem ./...

# Code quality
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

fmt:
	go fmt ./...

vet:
	go vet ./...

# Build
build:
	go build -o bin/basic examples/basic/main.go
	go build -o bin/advanced examples/advanced/main.go

# Examples
examples: build
	@echo "Running basic example..."
	./bin/basic
	@echo
	@echo "Running advanced example..."
	./bin/advanced

# Clean
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Development helpers
mod-tidy:
	go mod tidy

mod-download:
	go mod download

deps: mod-download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# CI targets
ci-test: test-unit test-race test-coverage lint vet

ci-integration: test-integration

# Release preparation
check-release: ci-test
	@echo "Release checks passed!"

.DEFAULT_GOAL := help