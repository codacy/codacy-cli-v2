#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}===== Building Trivy for ARM64 =====${NC}"

# Get the Trivy installation directory
TRIVY_INSTALL_DIR="$HOME/.cache/codacy/tools"
TRIVY_VERSION=$(find $TRIVY_INSTALL_DIR -type d -name "trivy@*" | head -n 1 | xargs basename | cut -d '@' -f 2)

if [ -z "$TRIVY_VERSION" ]; then
    echo -e "${RED}Error: Could not find Trivy installation directory. Please install Trivy using Codacy CLI first.${NC}"
    exit 1
fi

TRIVY_FULL_DIR="$TRIVY_INSTALL_DIR/trivy@$TRIVY_VERSION"
TRIVY_BINARY="$TRIVY_FULL_DIR/trivy"

echo -e "${YELLOW}Found Trivy installation: $TRIVY_FULL_DIR${NC}"
echo -e "${YELLOW}Current Trivy version: $TRIVY_VERSION${NC}"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}Go is not installed. Installing...${NC}"
    sudo apt update
    sudo apt install -y golang-go
fi

echo -e "${GREEN}Go is installed: $(go version)${NC}"

# Create a temporary directory for building
BUILD_DIR=$(mktemp -d)
echo -e "${YELLOW}Building in temporary directory: $BUILD_DIR${NC}"
cd "$BUILD_DIR"

# Clone the Trivy repository
echo -e "${BLUE}Cloning Trivy repository...${NC}"
git clone --depth 1 --branch "v$TRIVY_VERSION" https://github.com/aquasecurity/trivy.git
cd trivy

# Build Trivy
echo -e "${BLUE}Building Trivy specifically for this ARM64 architecture...${NC}"
GOARCH=arm64 go build -o trivy cmd/trivy/main.go

# Test the binary
echo -e "${BLUE}Testing the compiled binary...${NC}"
./trivy --version

# Backup the original binary
if [ -f "$TRIVY_BINARY" ]; then
    echo -e "${YELLOW}Backing up original Trivy binary...${NC}"
    mv "$TRIVY_BINARY" "$TRIVY_BINARY.bak"
fi

# Copy the new binary
echo -e "${GREEN}Installing the new Trivy binary...${NC}"
cp trivy "$TRIVY_BINARY"
chmod +x "$TRIVY_BINARY"

# Clean up
echo -e "${BLUE}Cleaning up...${NC}"
cd
rm -rf "$BUILD_DIR"

echo -e "${GREEN}Successfully built and installed Trivy for ARM64!${NC}"
echo -e "${YELLOW}You can now run Trivy on your ARM64 system.${NC}"
echo -e "${BLUE}If you encounter any issues, the original binary is backed up at: $TRIVY_BINARY.bak${NC}" 