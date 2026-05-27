#!/bin/sh
set -e

REPO="declaw-ai/declaw-cli"
BINARY="declaw"
INSTALL_DIR="/usr/local/bin"

command -v curl >/dev/null || { echo "Error: curl is required but not installed."; exit 1; }

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  mingw*|msys*|cygwin*) echo "Windows detected. Please download the binary from https://github.com/$REPO/releases"; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed 's/.*"v//' | sed 's/".*//')
if [ -z "$VERSION" ]; then
  echo "Failed to fetch latest version"
  exit 1
fi

URL="https://github.com/$REPO/releases/download/v${VERSION}/${BINARY}-${OS}-${ARCH}"
echo "Downloading declaw v${VERSION} for ${OS}/${ARCH}..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$BINARY"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
  chmod +x "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
  sudo chmod +x "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "declaw v${VERSION} installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Next: declaw auth login"
echo "Docs: https://docs.declaw.ai"
