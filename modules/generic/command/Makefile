.PHONY: build clean

build:
	@echo "Building command.wasm..."
	@tinygo build -o wasm/command.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/command.wasm"

clean:
	@rm -f wasm/command.wasm
