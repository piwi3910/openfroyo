.PHONY: build clean

build:
	@echo "Building archive.wasm..."
	@cd wasm && tinygo build -o archive.wasm -target=wasi main.go
	@echo "Build complete: wasm/archive.wasm"

clean:
	@rm -f wasm/archive.wasm
