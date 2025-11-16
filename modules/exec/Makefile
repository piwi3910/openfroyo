.PHONY: build clean

build:
	@echo "Building exec.wasm..."
	@tinygo build -o wasm/exec.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/exec.wasm"

clean:
	@rm -f wasm/exec.wasm
