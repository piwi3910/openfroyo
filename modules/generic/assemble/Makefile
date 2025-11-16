.PHONY: build clean

build:
	@echo "Building assemble.wasm..."
	@cd wasm && tinygo build -o assemble.wasm -target=wasi main.go
	@echo "Build complete: wasm/assemble.wasm"

clean:
	@rm -f wasm/assemble.wasm
