.PHONY: build clean

build:
	@echo "Building gem.wasm..."
	@tinygo build -o wasm/gem.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/gem.wasm"

clean:
	@rm -f wasm/gem.wasm
