.PHONY: build clean

build:
	@echo "Building tempfile.wasm..."
	@cd wasm && tinygo build -o tempfile.wasm -target=wasi main.go
	@echo "Build complete: wasm/tempfile.wasm"

clean:
	@rm -f wasm/tempfile.wasm
