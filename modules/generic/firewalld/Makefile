.PHONY: build clean

build:
	@echo "Building firewalld.wasm..."
	@tinygo build -o wasm/firewalld.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/firewalld.wasm"

clean:
	@rm -f wasm/firewalld.wasm
