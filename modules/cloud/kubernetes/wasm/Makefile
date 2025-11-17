.PHONY: build clean test

# WASM build settings
WASM_BINARY = kubernetes.wasm
GO_SOURCES = main.go

# Build the WASM module
build:
	@echo "Building Kubernetes WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_BINARY) $(GO_SOURCES)
	@echo "Build complete: $(WASM_BINARY)"
	@ls -lh $(WASM_BINARY)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_BINARY)
	@echo "Clean complete"

# Test the module (validates Go code)
test:
	@echo "Running Go tests..."
	go test -v ./...
	@echo "Tests complete"

# Validate Go code
lint:
	@echo "Linting Go code..."
	go vet ./...
	@echo "Lint complete"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Format complete"

# Build with optimization
build-optimized:
	@echo "Building optimized Kubernetes WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -ldflags="-s -w" -o $(WASM_BINARY) $(GO_SOURCES)
	@echo "Optimized build complete: $(WASM_BINARY)"
	@ls -lh $(WASM_BINARY)

# All: format, lint, and build
all: fmt lint build
