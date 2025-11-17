.PHONY: build clean test

build:
	@echo "Building get_url.wasm..."
	@tinygo build -o get_url.wasm -target=wasi -no-debug main.go
	@echo "Build complete: get_url.wasm"

clean:
	@rm -f get_url.wasm
	@echo "Cleaned"

test:
	@echo "Running tests..."
	@go test -v ./...
