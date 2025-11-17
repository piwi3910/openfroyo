.PHONY: build clean test

build:
	@echo "Building service.wasm..."
	@tinygo build -o service.wasm -target=wasi -no-debug main.go
	@echo "Build complete: service.wasm"

clean:
	@rm -f service.wasm
	@echo "Cleaned"

test:
	@echo "Running tests..."
	@go test -v ./...
