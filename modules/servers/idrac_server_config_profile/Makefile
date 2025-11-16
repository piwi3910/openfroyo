.PHONY: build clean

build:
	@echo "Building idrac_server_config_profile.wasm..."
	@tinygo build -o wasm/idrac_server_config_profile.wasm -target=wasi -no-debug wasm/main.go
	@echo "Build complete: wasm/idrac_server_config_profile.wasm"

clean:
	@rm -f wasm/idrac_server_config_profile.wasm
