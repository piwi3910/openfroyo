.PHONY: build clean

build:
	@echo "Building lineinfile.wasm..."
	@cd wasm && tinygo build -o lineinfile.wasm -target=wasi main.go
	@echo "Build complete: wasm/lineinfile.wasm"

clean:
	@rm -f wasm/lineinfile.wasm
