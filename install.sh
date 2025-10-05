#!/usr/bin/env bash
set -euo pipefail

# Version logic: positional argument > environment variable > default 'v1.0.0'
VERSION="${1:-v1.0.0}"
REPO="Mayank12agrawal/kcap"
BINARY="kcap"

# Detect OS and normalize architecture names
OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

if [[ "$ARCH" == "x86_64" ]]; then 
  ARCH="amd64"
elif [[ "$ARCH" == "arm64" || "$ARCH" == "aarch64" ]]; then 
  ARCH="arm64"
elif [[ "$ARCH" == "i386" ]]; then
  ARCH="386"
else
  echo "❌ Unsupported architecture: $ARCH"
  exit 1
fi

# Remove leading 'v' from version for the tarball name
CLEAN_VERSION="${VERSION#v}"

# Construct the tarball name matching your release assets
# Change this line in install.sh:
TARBALL="${BINARY}_vv${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "📥 Checking if asset $TARBALL exists at $URL..."

if ! curl -fsI "$URL" > /dev/null; then
  echo "❌ Release asset not found:"
  echo "   $URL"
  echo "❓ Please verify the release exists and the asset is uploaded exactly with this name."
  exit 1
fi

echo "⬇️ Downloading $TARBALL ..."
curl -fLo "$TARBALL" "$URL"

echo "📦 Extracting $BINARY from $TARBALL ..."
tar -xzf "$TARBALL"

chmod +x "$BINARY"

echo "🛠 Installing $BINARY to /usr/local/bin (may require sudo)..."
if mv "$BINARY" /usr/local/bin/ 2>/dev/null; then
  echo "✅ Installed without sudo."
else
  echo "🔐 Installing with sudo..."
  sudo mv "$BINARY" /usr/local/bin/
  echo "✅ Installed with sudo."
fi

rm -f "$TARBALL"

echo "🎉 $BINARY $VERSION installed successfully!"
$BINARY --help
