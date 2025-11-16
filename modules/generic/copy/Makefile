.PHONY: build clean

build:
	@echo "Building copy.wasm..."
	@cd wasm && tinygo build -o copy.wasm -target=wasi main.go
	@echo "Build complete: wasm/copy.wasm"

clean:
	@rm -f wasm/copy.wasm
