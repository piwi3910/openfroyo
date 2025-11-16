.PHONY: build clean

build:
	@echo "Building xattr.wasm..."
	@cd wasm && tinygo build -o xattr.wasm -target=wasi main.go
	@echo "Build complete: wasm/xattr.wasm"

clean:
	@rm -f wasm/xattr.wasm
