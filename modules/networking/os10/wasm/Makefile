.PHONY: build clean test

WASM_FILE = os10.wasm
GO_SOURCE = main.go

build:
	@echo "Building Dell OS10 WASM module..."
	tinygo build -o $(WASM_FILE) -target=wasi -no-debug $(GO_SOURCE)
	@echo "Build complete: $(WASM_FILE)"

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_FILE)
	@echo "Clean complete"

test:
	@echo "Running Go tests..."
	go test -v .
	@echo "Tests complete"

verify:
	@echo "Verifying WASM module..."
	@if [ -f $(WASM_FILE) ]; then \
		echo "WASM module exists: $(WASM_FILE)"; \
		ls -lh $(WASM_FILE); \
	else \
		echo "ERROR: WASM module not found"; \
		exit 1; \
	fi

all: clean build verify
