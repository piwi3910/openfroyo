.PHONY: build clean

WASM_OUT = postgresql.wasm

build:
	@echo "Building PostgreSQL WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(WASM_OUT) main.go
	@echo "Build complete: $(WASM_OUT)"

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(WASM_OUT)
	@echo "Clean complete"

test:
	@echo "Running Go tests..."
	go test -v ./...

vet:
	@echo "Running Go vet..."
	go vet ./...

fmt:
	@echo "Formatting Go code..."
	go fmt ./...

all: fmt vet build
