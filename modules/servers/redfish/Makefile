.PHONY: build clean test

build:
	@echo "Building redfish.wasm..."
	@tinygo build -o wasm/redfish.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/redfish.wasm"

clean:
	@rm -f wasm/redfish.wasm

test: build
	@echo "Running Redfish module tests..."
	@echo "Note: Tests require a Redfish-enabled server"
	@echo "Update test.ofy with your server details before running"
