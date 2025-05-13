.PHONY: build clean build-all build-linux build-darwin build-windows

# Get the version from git describe or fallback to a default version
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "development")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -X 'codacy/cli-v2/version.Version=$(VERSION)' \
           -X 'codacy/cli-v2/version.GitCommit=$(COMMIT)' \
           -X 'codacy/cli-v2/version.BuildTime=$(BUILD_TIME)'

# Build the CLI for current platform
build:
	go build -ldflags "$(LDFLAGS)" -o cli-v2

# Build for Linux
build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2-linux

# Build for macOS
build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2-macos

# Build for Windows
build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2.exe

# Build for all platforms
build-all: build-linux build-darwin build-windows

# Clean build artifacts
clean:
	rm -f cli-v2 cli-v2-linux cli-v2-macos cli-v2.exe 