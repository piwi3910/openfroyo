.PHONY: build clean

build:
	@echo "Building xml.wasm..."
	@cd wasm && tinygo build -o xml.wasm -target=wasi main.go
	@echo "Build complete: wasm/xml.wasm"

clean:
	@rm -f wasm/xml.wasm
