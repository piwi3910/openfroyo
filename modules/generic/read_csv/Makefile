.PHONY: build clean

build:
	@echo "Building read_csv.wasm..."
	@cd wasm && tinygo build -o read_csv.wasm -target=wasi main.go
	@echo "Build complete: wasm/read_csv.wasm"

clean:
	@rm -f wasm/read_csv.wasm
