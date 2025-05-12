.PHONY: build clean

# Get the version from git describe or fallback to a default version
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "development")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Build flags
LDFLAGS := -X 'codacy/cli-v2/version.Version=$(VERSION)' -X 'codacy/cli-v2/version.GitCommit=$(COMMIT)' -X 'codacy/cli-v2/version.BuildTime=$(BUILD_TIME)'

# Build the CLI
build:
	go build -ldflags "$(LDFLAGS)" -o cli-v2

# Clean build artifacts
clean:
	rm -f cli-v2

# Build for all platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2_linux_amd64
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2_darwin_amd64
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o cli-v2_windows_amd64.exe 