.PHONY: build-cli install-cli clean help

# Build the CLI binary
build-cli:
	@echo "Building Strata CLI..."
	cd cli && go build -ldflags="-s -w" -o strata .

# Install the CLI globally
install-cli: build-cli
	@echo "Installing Strata CLI to /usr/local/bin..."
	cp cli/strata /usr/local/bin/strata
	@echo "Done! Run 'strata --help' to get started."

# Build for multiple platforms
build-all:
	@echo "Building Strata CLI for all platforms..."
	cd cli && GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o strata-linux-amd64 .
	cd cli && GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o strata-darwin-amd64 .
	cd cli && GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o strata-darwin-arm64 .
	cd cli && GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o strata-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -f cli/strata cli/strata-*

help:
	@echo "Targets:"
	@echo "  build-cli    - Build the CLI for current platform"
	@echo "  install-cli  - Build and install to /usr/local/bin"
	@echo "  build-all    - Cross-compile for Linux, macOS, Windows"
	@echo "  clean        - Remove build artifacts"
