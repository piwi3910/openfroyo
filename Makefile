.PHONY: build test lint clean install-tools

# Build variables
BINARY_NAME=froyo
MICRO_RUNNER=micro-runner
VERSION?=0.1.0-dev
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_DIR=bin
GO=go
GOFLAGS=-trimpath
LDFLAGS=-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildDate=$(BUILD_DATE)

# Build main CLI
build:
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/froyo

# Build micro-runner (static binary)
build-runner:
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(MICRO_RUNNER) ./cmd/micro-runner

# Build all binaries
build-all: build build-runner

# Run tests
test:
	$(GO) test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GO) tool cover -html=coverage.txt -o coverage.html

# Run linter
lint:
	golangci-lint run

# Format code
fmt:
	$(GO) fmt ./...
	gofumpt -l -w .

# Vet code
vet:
	$(GO) vet ./...

# Run all checks
check: fmt vet lint test

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.txt coverage.html
	find . -name "*.test" -delete

# Install development tools
install-tools:
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GO) install mvdan.cc/gofumpt@latest

# Build for all platforms
build-cross:
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/froyo
	GOOS=linux GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/froyo
	GOOS=darwin GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/froyo
	GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/froyo

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the froyo CLI"
	@echo "  build-runner  - Build the micro-runner"
	@echo "  build-all     - Build all binaries"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  lint          - Run linter"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  check         - Run all checks (fmt, vet, lint, test)"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  build-cross   - Build for all platforms"
