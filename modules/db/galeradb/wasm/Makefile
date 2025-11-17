.PHONY: build clean test

WASM_OUTPUT = galeradb.wasm
GO_SOURCE = main.go

build:
	@echo "Building Galera Cluster management WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUTPUT) $(GO_SOURCE)
	@echo "Build complete: $(WASM_OUTPUT)"
	@ls -lh $(WASM_OUTPUT)

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_OUTPUT)
	@echo "Clean complete"

test: build
	@echo "WASM module built successfully"
	@echo "Size: $$(ls -lh $(WASM_OUTPUT) | awk '{print $$5}')"

# Verify the WASM module structure
verify: build
	@echo "Verifying WASM module..."
	@file $(WASM_OUTPUT) || echo "file command not available"
	@wasm-objdump -h $(WASM_OUTPUT) 2>/dev/null || echo "wasm-objdump not available (optional)"
