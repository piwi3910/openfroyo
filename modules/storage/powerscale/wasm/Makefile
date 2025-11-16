.PHONY: build clean test

BINARY_NAME=powerscale.wasm
GO_FILES=$(wildcard *.go)

build: $(BINARY_NAME)

$(BINARY_NAME): $(GO_FILES)
	@echo "Building PowerScale WASM module..."
	GOOS=wasip1 GOARCH=wasm go build -o $(BINARY_NAME) .
	@echo "Build complete: $(BINARY_NAME)"
	@ls -lh $(BINARY_NAME)

clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)

test:
	@echo "Running Go tests..."
	go test -v .

fmt:
	@echo "Formatting Go code..."
	go fmt .

vet:
	@echo "Vetting Go code..."
	go vet .

all: fmt vet build
