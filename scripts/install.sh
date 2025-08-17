#!/bin/bash

set -e

GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

REPO="makestatic/droplink"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux) OS="linux" ;;
  mingw*|msys*|cygwin*) OS="windows" ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

echo -e "${BLUE}Getting latest version...${NC}"
VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

echo -e "${BLUE}Downloading droplink $VERSION for $OS/$ARCH...${NC}"
FILENAME="droplink-$OS-$ARCH"
if [ "$OS" = "windows" ]; then
  FILENAME="$FILENAME.zip"
else
  FILENAME="$FILENAME.tar.gz"
fi

URL="https://github.com/$REPO/releases/download/$VERSION/$FILENAME"
curl -sL "$URL" | tar -xz -C "$INSTALL_DIR" --strip-components=0 droplink 2>/dev/null || {
  TMP=$(mktemp -d)
  curl -sL "$URL" -o "$TMP/$FILENAME"
  if [ "$OS" = "windows" ]; then
    unzip -q "$TMP/$FILENAME" -d "$TMP"
    cp "$TMP/droplink.exe" "$INSTALL_DIR/"
  else
    tar -xzf "$TMP/$FILENAME" -C "$INSTALL_DIR"
  fi
  rm -rf "$TMP"
}

chmod +x "$INSTALL_DIR/droplink"
echo -e "${GREEN}✓ Installed droplink $VERSION to $INSTALL_DIR${NC}"
