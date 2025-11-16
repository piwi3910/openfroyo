.PHONY: build clean

build:
	@echo "Building iLO WASM module..."
	cd wasm && tinygo build -o ilo.wasm -target=wasi -no-debug main.go
	@echo "Build complete: wasm/ilo.wasm"

clean:
	rm -f wasm/ilo.wasm
