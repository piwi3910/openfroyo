.PHONY: build clean

build:
	@echo "Building stat.wasm..."
	@cd wasm && tinygo build -o stat.wasm -target=wasi main.go
	@echo "Build complete: wasm/stat.wasm"

clean:
	@rm -f wasm/stat.wasm
