#!/usr/bin/env sh
# fractalx-init installer
# Usage: curl -fsSL https://raw.githubusercontent.com/fractalx/fractalx-init/main/install.sh | sh

set -e

REPO="fractalx/fractalx-init"
INSTALL_DIR="${FRACTALX_INSTALL_DIR:-/usr/local/bin}"
BINARY="fractalx-init"

# Detect OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

if [ "$OS" = "darwin" ]; then
  EXT="tar.gz"
elif [ "$OS" = "linux" ]; then
  EXT="tar.gz"
else
  echo "Unsupported OS: $OS"
  echo "On Windows, download the .zip from https://github.com/$REPO/releases"
  exit 1
fi

# Fetch latest version tag
LATEST=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')
if [ -z "$LATEST" ]; then
  echo "Could not determine latest release. Check https://github.com/$REPO/releases"
  exit 1
fi

VERSION="${LATEST#v}"
FILENAME="${BINARY}_${VERSION}_${OS}_${ARCH}.${EXT}"
URL="https://github.com/$REPO/releases/download/$LATEST/$FILENAME"

echo "Installing fractalx-init $LATEST ($OS/$ARCH)..."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$FILENAME"
tar -xzf "$TMP/$FILENAME" -C "$TMP"

if [ ! -f "$TMP/$BINARY" ]; then
  echo "Binary not found in archive. Please report this at https://github.com/$REPO/issues"
  exit 1
fi

chmod +x "$TMP/$BINARY"

if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR requires sudo..."
  sudo mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
fi

echo ""
echo "  fractalx-init $LATEST installed to $INSTALL_DIR/$BINARY"
echo ""
echo "  Get started:"
echo "    fractalx-init"
echo ""
