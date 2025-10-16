.PHONY: build clean test package install help

# Provider information
PROVIDER_NAME := linux.pkg
PROVIDER_VERSION := 1.0.0
WASM_OUTPUT := plugin.wasm

# Build configuration
TINYGO := tinygo
TINYGO_FLAGS := -target=wasi -opt=2 -no-debug -scheduler=none

# Package output
PACKAGE_FILE := $(PROVIDER_NAME)-$(PROVIDER_VERSION).tar.gz

help: ## Show this help message
	@echo "OpenFroyo linux.pkg Provider - Build System"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build the WASM provider module
	@echo "Building $(PROVIDER_NAME) WASM module..."
	@./build.sh

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -f $(WASM_OUTPUT)
	@rm -f $(PACKAGE_FILE)
	@rm -f ./*.tar.gz
	@echo "Clean complete"

test: ## Run provider tests (requires Go runtime)
	@echo "Running provider tests..."
	@go test -v ./...

package: build ## Build and package the provider for distribution
	@echo "Packaging $(PROVIDER_NAME) v$(PROVIDER_VERSION)..."
	@tar czf $(PACKAGE_FILE) \
		$(WASM_OUTPUT) \
		manifest.yaml \
		schemas/ \
		README.md \
		LICENSE
	@echo "Package created: $(PACKAGE_FILE)"
	@ls -lh $(PACKAGE_FILE)

install: package ## Build, package, and install the provider locally
	@echo "Installing $(PROVIDER_NAME) provider..."
	@froyo provider install $(PACKAGE_FILE)

validate: ## Validate WASM module
	@echo "Validating WASM module..."
	@if command -v wasm-validate >/dev/null 2>&1; then \
		wasm-validate $(WASM_OUTPUT); \
		echo "WASM module is valid"; \
	else \
		echo "wasm-validate not found. Install wabt for validation."; \
		exit 1; \
	fi

optimize: build ## Optimize WASM module with wasm-opt
	@echo "Optimizing WASM module..."
	@if command -v wasm-opt >/dev/null 2>&1; then \
		wasm-opt -O3 -o $(WASM_OUTPUT).opt $(WASM_OUTPUT); \
		mv $(WASM_OUTPUT).opt $(WASM_OUTPUT); \
		echo "Optimization complete"; \
	else \
		echo "wasm-opt not found. Install binaryen for optimization."; \
		exit 1; \
	fi

inspect: ## Inspect WASM module details
	@echo "Inspecting WASM module..."
	@if [ -f $(WASM_OUTPUT) ]; then \
		ls -lh $(WASM_OUTPUT); \
		echo ""; \
		if command -v wasm-objdump >/dev/null 2>&1; then \
			wasm-objdump -h $(WASM_OUTPUT); \
		else \
			echo "Install wabt for detailed inspection"; \
		fi; \
	else \
		echo "$(WASM_OUTPUT) not found. Run 'make build' first."; \
	fi

dev: ## Build for development (with debug symbols)
	@echo "Building for development..."
	$(TINYGO) build -o $(WASM_OUTPUT) -target=wasi -opt=1 main.go

# TinyGo version check
check-tinygo: ## Check if TinyGo is installed
	@echo "Checking TinyGo installation..."
	@if command -v $(TINYGO) >/dev/null 2>&1; then \
		echo "TinyGo found: $$($(TINYGO) version | head -n1)"; \
	else \
		echo "Error: TinyGo not found"; \
		echo "Install from: https://tinygo.org/getting-started/install/"; \
		exit 1; \
	fi
