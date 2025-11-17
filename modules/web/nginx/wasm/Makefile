.PHONY: build clean test

# WASM build settings
GOOS := wasip1
GOARCH := wasm
MODULE_NAME := nginx.wasm

# Build the WASM module
build:
	@echo "Building Nginx WASM module..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o $(MODULE_NAME) main.go
	@echo "Build complete: $(MODULE_NAME)"
	@ls -lh $(MODULE_NAME)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(MODULE_NAME)
	@echo "Clean complete"

# Test compilation (build without creating final artifact)
test:
	@echo "Testing compilation..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o /dev/null main.go
	@echo "Compilation test passed"

# Format Go code
fmt:
	@echo "Formatting Go code..."
	go fmt main.go
	@echo "Format complete"

# Vet Go code
vet:
	@echo "Vetting Go code..."
	go vet main.go
	@echo "Vet complete"

# Run all checks
check: fmt vet test
	@echo "All checks passed"

# Build with verbose output
build-verbose:
	@echo "Building Nginx WASM module (verbose)..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -v -o $(MODULE_NAME) main.go
	@echo "Build complete: $(MODULE_NAME)"
	@ls -lh $(MODULE_NAME)
