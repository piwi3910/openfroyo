.PHONY: build clean

build:
	@echo "Building iso_extract.wasm..."
	@cd wasm && tinygo build -o iso_extract.wasm -target=wasi main.go
	@echo "Build complete: wasm/iso_extract.wasm"

clean:
	@rm -f wasm/iso_extract.wasm
