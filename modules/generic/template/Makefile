.PHONY: build clean

build:
	@echo "Building template.wasm..."
	@cd wasm && tinygo build -o template.wasm -target=wasi main.go
	@echo "Build complete: wasm/template.wasm"

clean:
	@rm -f wasm/template.wasm
