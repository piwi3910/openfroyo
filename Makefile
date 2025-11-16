.PHONY: build build-runner build-wasm clean test help

help:
	@echo "OpenFroyo Build Targets:"
	@echo "  make build        - Build both froyo CLI and froyo-runner"
	@echo "  make build-wasm   - Build WASM modules (requires TinyGo)"
	@echo "  make clean        - Remove built binaries"
	@echo "  make test         - Run tests"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make build"
	@echo "  2. make build-wasm"
	@echo "  3. Edit examples/inventory/hosts.ofy.yml with your SSH host"
	@echo "  4. ./bin/froyo apply examples/stacks/test.ofy.yml"

build: build-runner build-cli

build-cli:
	@echo "Building froyo CLI..."
	@mkdir -p bin
	@go build -o bin/froyo ./cmd/froyo
	@echo "✓ Built: bin/froyo"

build-runner:
	@echo "Building froyo-runner (static Linux binary)..."
	@mkdir -p bin
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-runner ./cmd/froyo-runner
	@echo "✓ Built: bin/froyo-runner (Linux amd64)"

build-wasm:
	@echo "Building WASM modules..."
	@cd modules/exec && $(MAKE) build
	@echo "✓ All WASM modules built"

clean:
	@echo "Cleaning..."
	@rm -rf bin
	@cd modules/exec && $(MAKE) clean
	@echo "✓ Clean complete"

test:
	@echo "Running tests..."
	@go test ./...
