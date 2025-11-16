.PHONY: build clean

build:
	@echo "Building unarchive.wasm..."
	@cd wasm && tinygo build -o unarchive.wasm -target=wasi main.go
	@echo "Build complete: wasm/unarchive.wasm"

clean:
	@rm -f wasm/unarchive.wasm
