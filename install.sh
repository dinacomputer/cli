#!/bin/sh
set -e

REPO="dinacomputer/cli"
BINARY="dina"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS
OS="$(uname -s)"
case "$OS" in
  Linux)  OS="linux" ;;
  Darwin) OS="darwin" ;;
  *)      echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)  ARCH="amd64" ;;
  aarch64|arm64)  ARCH="arm64" ;;
  *)              echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# Get latest version
if [ -z "$VERSION" ]; then
  VERSION="$(curl -sL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"v([^"]+)".*/\1/')"
  if [ -z "$VERSION" ]; then
    echo "Error: Could not determine latest version."
    exit 1
  fi
fi

FILENAME="dina_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}"

echo "Downloading dina v${VERSION} for ${OS}/${ARCH}..."

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

curl -sL "$URL" -o "${TMPDIR}/${FILENAME}"
tar -xzf "${TMPDIR}/${FILENAME}" -C "$TMPDIR"

echo "Installing to ${INSTALL_DIR}/${BINARY}..."
install -d "$INSTALL_DIR"
install "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"

echo "dina v${VERSION} installed successfully!"
echo "Run 'dina --help' to get started."
