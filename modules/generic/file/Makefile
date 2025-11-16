.PHONY: build clean

build:
	@echo "Building file.wasm..."
	@cd wasm && tinygo build -o file.wasm -target=wasi main.go
	@echo "Build complete: wasm/file.wasm"

clean:
	@rm -f wasm/file.wasm
