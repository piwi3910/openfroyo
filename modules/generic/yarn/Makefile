.PHONY: build clean

build:
	@echo "Building yarn.wasm..."
	@tinygo build -o wasm/yarn.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/yarn.wasm"

clean:
	@rm -f wasm/yarn.wasm
