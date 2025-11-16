.PHONY: build clean

build:
	@echo "Building replace.wasm..."
	@cd wasm && tinygo build -o replace.wasm -target=wasi main.go
	@echo "Build complete: wasm/replace.wasm"

clean:
	@rm -f wasm/replace.wasm
