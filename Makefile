.PHONY: build build-agent build-runner build-wasm clean test help

help:
	@echo "OpenFroyo Build Targets:"
	@echo "  make build        - Build froyo CLI, froyo-agent, and froyo-runner"
	@echo "  make build-agent  - Build froyo-agent binary"
	@echo "  make build-wasm   - Build WASM modules (requires TinyGo)"
	@echo "  make clean        - Remove built binaries"
	@echo "  make test         - Run tests"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make build"
	@echo "  2. make build-wasm"
	@echo "  3. Edit examples/inventory/hosts.yml with your SSH host"
	@echo "  4. ./bin/froyo apply examples/stacks/test.ofy"

build: build-runner build-agent build-cli

build-cli:
	@echo "Building froyo CLI..."
	@mkdir -p bin
	@go build -o bin/froyo ./cmd/froyo
	@echo "✓ Built: bin/froyo"

build-agent:
	@echo "Building froyo-agent for multiple platforms..."
	@mkdir -p bin
	@echo "  → Linux amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-agent-linux-amd64 ./cmd/froyo-agent
	@echo "  → Linux arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/froyo-agent-linux-arm64 ./cmd/froyo-agent
	@echo "  → macOS amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-agent-darwin-amd64 ./cmd/froyo-agent
	@echo "  → macOS arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/froyo-agent-darwin-arm64 ./cmd/froyo-agent
	@echo "  → Windows amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-agent-windows-amd64.exe ./cmd/froyo-agent
	@echo "✓ Built froyo-agent for 5 platforms"

build-runner:
	@echo "Building froyo-runner for multiple platforms..."
	@mkdir -p bin
	@echo "  → Linux amd64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-runner-linux-amd64 ./cmd/froyo-runner
	@echo "  → Linux arm64..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o bin/froyo-runner-linux-arm64 ./cmd/froyo-runner
	@echo "  → macOS amd64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-runner-darwin-amd64 ./cmd/froyo-runner
	@echo "  → macOS arm64..."
	@CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o bin/froyo-runner-darwin-arm64 ./cmd/froyo-runner
	@echo "  → Windows amd64..."
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o bin/froyo-runner-windows-amd64.exe ./cmd/froyo-runner
	@echo "✓ Built froyo-runner for 5 platforms"

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
