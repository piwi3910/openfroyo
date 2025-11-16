.PHONY: build clean

build:
	@echo "Building idrac.wasm..."
	@tinygo build -o wasm/idrac.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/idrac.wasm"

clean:
	@rm -f wasm/idrac.wasm
