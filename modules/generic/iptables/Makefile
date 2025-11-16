.PHONY: build clean

build:
	@echo "Building iptables.wasm..."
	@tinygo build -o wasm/iptables.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/iptables.wasm"

clean:
	@rm -f wasm/iptables.wasm
