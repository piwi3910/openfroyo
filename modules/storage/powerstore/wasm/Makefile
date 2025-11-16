.PHONY: build clean test

WASM_NAME = powerstore.wasm
GO_FILES = main.go

build:
	@echo "Building PowerStore WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_NAME) $(GO_FILES)
	@echo "Build complete: $(WASM_NAME)"
	@ls -lh $(WASM_NAME)

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_NAME)
	@echo "Clean complete"

test: build
	@echo "WASM module built successfully"
	@file $(WASM_NAME) || echo "WASM binary created"
