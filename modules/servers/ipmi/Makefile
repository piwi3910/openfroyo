.PHONY: build clean test

build:
	@echo "Building IPMI WASM module..."
	tinygo build -o wasm/ipmi.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/ipmi.wasm"

clean:
	@echo "Cleaning build artifacts..."
	rm -f wasm/ipmi.wasm
	@echo "Clean complete"

test:
	@echo "Testing IPMI module..."
	@if [ ! -f wasm/ipmi.wasm ]; then \
		echo "Error: wasm/ipmi.wasm not found. Run 'make build' first."; \
		exit 1; \
	fi
	@echo "WASM module exists, ready for testing"
