#!/bin/bash

set -e

# Determine OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

case "$OS" in
  Darwin)
    OS_NAME="Darwin"
    ;;
  Linux)
    OS_NAME="Linux"
    ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

case "$ARCH" in
  x86_64)
    ARCH_NAME="x86_64"
    ;;
  arm64)
    ARCH_NAME="arm64"
    ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get the latest release URL
echo "Fetching latest Ludwig release..."
RELEASE_URL=$(curl -s https://api.github.com/repos/AlexanderHeffernan/Ludwig-AI/releases/latest | grep "browser_download_url" | grep "${OS_NAME}_${ARCH_NAME}" | head -n1 | cut -d'"' -f4)

if [ -z "$RELEASE_URL" ]; then
  echo "Error: Could not find release for $OS_NAME/$ARCH_NAME"
  exit 1
fi

echo "Downloading from: $RELEASE_URL"
curl -L "$RELEASE_URL" | tar xz

echo "Installing to /usr/local/bin/ludwig..."
sudo mv ludwig /usr/local/bin/

echo ""
echo "✓ Ludwig installed successfully!"
echo ""
ludwig --version
