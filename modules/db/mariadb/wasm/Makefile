.PHONY: build clean test

WASM_TARGET = mariadb.wasm
GO_SOURCE = main.go

build:
	@echo "Building MariaDB WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_TARGET) $(GO_SOURCE)
	@echo "Build complete: $(WASM_TARGET)"
	@ls -lh $(WASM_TARGET)

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_TARGET)
	@echo "Clean complete"

test: build
	@echo "WASM module built successfully"
	@file $(WASM_TARGET) || true
