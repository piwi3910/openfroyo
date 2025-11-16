.PHONY: build clean

build:
	@echo "Building package.wasm..."
	@tinygo build -o wasm/package.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/package.wasm"

clean:
	@rm -f wasm/package.wasm
