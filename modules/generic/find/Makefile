.PHONY: build clean

build:
	@echo "Building find.wasm..."
	@cd wasm && tinygo build -o find.wasm -target=wasi main.go
	@echo "Build complete: wasm/find.wasm"

clean:
	@rm -f wasm/find.wasm
