.PHONY: build clean

build:
	@echo "Building ini_file.wasm..."
	@cd wasm && tinygo build -o ini_file.wasm -target=wasi main.go
	@echo "Build complete: wasm/ini_file.wasm"

clean:
	@rm -f wasm/ini_file.wasm
