.PHONY: build clean

build:
	@echo "Building acl.wasm..."
	@cd wasm && tinygo build -o acl.wasm -target=wasi main.go
	@echo "Build complete: wasm/acl.wasm"

clean:
	@rm -f wasm/acl.wasm
