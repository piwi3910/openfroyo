.PHONY: build clean

build:
	@echo "Building pip.wasm..."
	@cd wasm && tinygo build -o pip.wasm -target=wasi -no-debug main.go
	@echo "Build complete: wasm/pip.wasm"

clean:
	@rm -f wasm/pip.wasm
