.PHONY: build clean

build:
	@echo "Building npm.wasm..."
	@tinygo build -o wasm/npm.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/npm.wasm"

clean:
	@rm -f wasm/npm.wasm
