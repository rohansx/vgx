#!/bin/bash
# VGX Installer
# Usage: curl -sSL https://vgx.sh/install | bash

set -e

REPO="rohansx/vgx"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="vgx"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}"
echo "╔═══════════════════════════════════════╗"
echo "║       VGX Installer                   ║"
echo "║   AI Code Security Scanner            ║"
echo "╚═══════════════════════════════════════╝"
echo -e "${NC}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

case $OS in
    linux|darwin)
        ;;
    mingw*|msys*|cygwin*)
        OS="windows"
        ;;
    *)
        echo -e "${RED}Unsupported OS: $OS${NC}"
        exit 1
        ;;
esac

SUFFIX="${OS}_${ARCH}"
if [ "$OS" = "windows" ]; then
    SUFFIX="${SUFFIX}.exe"
    BINARY_NAME="vgx.exe"
fi

echo "Detected: $OS/$ARCH"

# Get latest release
echo "Fetching latest release..."
LATEST=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
    echo -e "${YELLOW}No release found, using main branch...${NC}"
    DOWNLOAD_URL="https://github.com/${REPO}/raw/main/vgx"
else
    echo "Latest version: $LATEST"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST}/vgx_${SUFFIX}"
fi

# Download
echo "Downloading VGX..."
TMP_FILE=$(mktemp)
if ! curl -sSL "$DOWNLOAD_URL" -o "$TMP_FILE"; then
    echo -e "${RED}Failed to download VGX${NC}"
    rm -f "$TMP_FILE"
    exit 1
fi

# Install
echo "Installing to $INSTALL_DIR..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
    sudo mv "$TMP_FILE" "$INSTALL_DIR/$BINARY_NAME"
    sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

# Verify
if command -v vgx &> /dev/null; then
    echo -e "${GREEN}"
    echo "╔═══════════════════════════════════════╗"
    echo "║  ✓ VGX installed successfully!        ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"
    echo ""
    echo "Quick start:"
    echo "  vgx detect --path ./src    # Detect AI-generated code"
    echo "  vgx scan                   # Security scan"
    echo "  vgx help                   # Show all commands"
    echo ""
else
    echo -e "${RED}Installation failed. Please check permissions.${NC}"
    exit 1
fi
