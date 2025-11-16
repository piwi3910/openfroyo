.PHONY: build clean

build:
	@echo "Building user.wasm..."
	@tinygo build -o wasm/user.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/user.wasm"

clean:
	@rm -f wasm/user.wasm
