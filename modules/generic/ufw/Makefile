.PHONY: build clean

build:
	@echo "Building ufw.wasm..."
	@tinygo build -o wasm/ufw.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/ufw.wasm"

clean:
	@rm -f wasm/ufw.wasm
