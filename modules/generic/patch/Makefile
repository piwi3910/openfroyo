.PHONY: build clean

build:
	@echo "Building patch.wasm..."
	@cd wasm && tinygo build -o patch.wasm -target=wasi main.go
	@echo "Build complete: wasm/patch.wasm"

clean:
	@rm -f wasm/patch.wasm
