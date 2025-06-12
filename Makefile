.PHONY: build build-linux build-windows build-darwin clean release

# Version and build information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build for all platforms
build: build-linux build-windows build-darwin

# Build for Linux (amd64)
build-linux:
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/marinenp-linux main.go

# Build for Windows (amd64)
build-windows:
	CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o bin/marinenp-windows.exe main.go

# Build for macOS (arm64 - Apple Silicon)
build-darwin:
	CC=aarch64-linux-gnu-gcc CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o bin/marinenp-macos main.go

# Create release package
release: build
	@echo "Creating release package..."
	@mkdir -p release
	@cp -r public release/
	@cp -r bin/* release/
	@gunzip -c marinenp-sqlite.db.gz > release/marinenp-sqlite.db 2>/dev/null || true
	@cd release && zip -r ../release_app.zip .
	@echo "Release package created: release_app.zip"

# Clean build artifacts
clean:
	rm -rf bin/ release/ release_app.zip 