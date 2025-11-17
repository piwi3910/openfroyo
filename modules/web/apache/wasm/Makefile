# Makefile for Apache WASM Module

# Output file
WASM_OUTPUT = apache.wasm

# Go source files
GO_SOURCES = main.go

# WASI SDK settings (if available)
GOOS = wasip1
GOARCH = wasm

# Build flags for WASM
BUILD_FLAGS = -ldflags="-s -w"

# Default target
.PHONY: all
all: build

# Build the WASM module
.PHONY: build
build: $(WASM_OUTPUT)

$(WASM_OUTPUT): $(GO_SOURCES)
	@echo "Building Apache WASM module..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) -o $(WASM_OUTPUT) $(GO_SOURCES)
	@echo "Build complete: $(WASM_OUTPUT)"
	@ls -lh $(WASM_OUTPUT)

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_OUTPUT)
	@echo "Clean complete"

# Test compilation without optimization
.PHONY: test-build
test-build:
	@echo "Testing WASM compilation..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o test-$(WASM_OUTPUT) $(GO_SOURCES)
	@echo "Test build complete"
	@ls -lh test-$(WASM_OUTPUT)
	@rm -f test-$(WASM_OUTPUT)

# Format Go code
.PHONY: fmt
fmt:
	@echo "Formatting Go code..."
	go fmt $(GO_SOURCES)

# Run Go vet
.PHONY: vet
vet:
	@echo "Running go vet..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go vet $(GO_SOURCES)

# Check for syntax errors
.PHONY: check
check: fmt vet
	@echo "Running syntax checks..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o /dev/null $(GO_SOURCES)
	@echo "Syntax check complete"

# Rebuild (clean + build)
.PHONY: rebuild
rebuild: clean build

# Help target
.PHONY: help
help:
	@echo "Apache WASM Module Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  all         - Build the WASM module (default)"
	@echo "  build       - Build the WASM module"
	@echo "  clean       - Remove build artifacts"
	@echo "  test-build  - Test compilation without optimization"
	@echo "  fmt         - Format Go source code"
	@echo "  vet         - Run go vet on source code"
	@echo "  check       - Run syntax checks (fmt + vet + compile)"
	@echo "  rebuild     - Clean and rebuild"
	@echo "  help        - Show this help message"
	@echo ""
	@echo "Environment:"
	@echo "  GOOS=$(GOOS)"
	@echo "  GOARCH=$(GOARCH)"
