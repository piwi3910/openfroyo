.PHONY: build clean

build:
	@echo "Building authorized_keys.wasm..."
	@tinygo build -o wasm/authorized_keys.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/authorized_keys.wasm"

clean:
	@rm -f wasm/authorized_keys.wasm
