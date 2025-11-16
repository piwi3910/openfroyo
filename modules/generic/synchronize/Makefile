.PHONY: build clean

build:
	@echo "Building synchronize.wasm..."
	@cd wasm && tinygo build -o synchronize.wasm -target=wasi main.go
	@echo "Build complete: wasm/synchronize.wasm"

clean:
	@rm -f wasm/synchronize.wasm
