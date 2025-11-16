.PHONY: build clean

WASM_FILE = sonic.wasm
GO_SOURCE = main.go

build:
	@echo "Building SONiC WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_FILE) $(GO_SOURCE)
	@echo "Build complete: $(WASM_FILE)"

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_FILE)
	@echo "Clean complete"

test:
	@echo "Testing Go code compilation..."
	go build -o /dev/null $(GO_SOURCE)
	@echo "Test complete"
