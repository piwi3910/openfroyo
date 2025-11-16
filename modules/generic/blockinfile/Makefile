.PHONY: build clean

build:
	@echo "Building blockinfile.wasm..."
	@cd wasm && tinygo build -o blockinfile.wasm -target=wasi main.go
	@echo "Build complete: wasm/blockinfile.wasm"

clean:
	@rm -f wasm/blockinfile.wasm
