.PHONY: build clean

build:
	@echo "Building group.wasm..."
	@tinygo build -o wasm/group.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/group.wasm"

clean:
	@rm -f wasm/group.wasm
