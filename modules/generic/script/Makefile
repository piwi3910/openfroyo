.PHONY: build clean

build:
	@echo "Building script.wasm..."
	@tinygo build -o wasm/script.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/script.wasm"

clean:
	@rm -f wasm/script.wasm
