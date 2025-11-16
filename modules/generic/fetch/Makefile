.PHONY: build clean

build:
	@echo "Building fetch.wasm..."
	@cd wasm && tinygo build -o fetch.wasm -target=wasi main.go
	@echo "Build complete: wasm/fetch.wasm"

clean:
	@rm -f wasm/fetch.wasm
