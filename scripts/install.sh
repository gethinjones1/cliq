#!/bin/bash
set -e

# Cliq Installation Script
# Downloads and installs the latest version of Cliq

REPO="cliq-cli/cliq"
INSTALL_DIR="/usr/local/bin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Cliq Installer${NC}"
echo "AI-powered CLI assistant for Neovim and tmux"
echo ""

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
    darwin) OS="darwin" ;;
    linux) OS="linux" ;;
    *)
        echo -e "${RED}Unsupported operating system: $OS${NC}"
        exit 1
        ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo "Detected: ${OS}/${ARCH}"

# Check for required tools
if ! command -v curl &> /dev/null; then
    echo -e "${RED}curl is required but not installed.${NC}"
    exit 1
fi

# Get latest release version
echo "Fetching latest release..."
LATEST_VERSION=$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${YELLOW}Could not determine latest version. Using 'latest'.${NC}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/latest/download/cliq-${OS}-${ARCH}.tar.gz"
else
    echo "Latest version: ${LATEST_VERSION}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/cliq-${OS}-${ARCH}.tar.gz"
fi

# Create temp directory
TMP_DIR=$(mktemp -d)
trap "rm -rf $TMP_DIR" EXIT

# Download
echo "Downloading from ${DOWNLOAD_URL}..."
if ! curl -sL "$DOWNLOAD_URL" -o "${TMP_DIR}/cliq.tar.gz"; then
    echo -e "${RED}Download failed.${NC}"
    exit 1
fi

# Extract
echo "Extracting..."
tar -xzf "${TMP_DIR}/cliq.tar.gz" -C "$TMP_DIR"

# Install
echo "Installing to ${INSTALL_DIR}..."
if [ -w "$INSTALL_DIR" ]; then
    mv "${TMP_DIR}/cliq-${OS}-${ARCH}" "${INSTALL_DIR}/cliq"
else
    echo "Requires sudo to install to ${INSTALL_DIR}"
    sudo mv "${TMP_DIR}/cliq-${OS}-${ARCH}" "${INSTALL_DIR}/cliq"
fi

chmod +x "${INSTALL_DIR}/cliq"

# Verify installation
if command -v cliq &> /dev/null; then
    echo ""
    echo -e "${GREEN}✓ Cliq installed successfully!${NC}"
    echo ""
    cliq version
    echo ""
    echo "Get started:"
    echo "  cliq init                    # Download model and setup"
    echo "  cliq \"how do I delete a line\" # Ask a question"
    echo "  cliq -i                      # Interactive mode"
else
    echo ""
    echo -e "${YELLOW}Installation complete, but 'cliq' not found in PATH.${NC}"
    echo "You may need to add ${INSTALL_DIR} to your PATH."
fi
