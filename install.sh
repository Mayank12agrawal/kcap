#!/usr/bin/env bash
set -euo pipefail

# Repository and binary
REPO="Mayank12agrawal/kcap"
BINARY="kcap"

# Get version: argument > latest GitHub release
if [ $# -ge 1 ]; then
  VERSION="$1"
else
  VERSION=$(curl -s https://api.github.com/repos/${REPO}/releases/latest | sed -n 's/.*"tag_name": "\([^"]*\)".*/\1/p')
fi

echo "ℹ️ Installing $BINARY version: $VERSION"

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

# Remove leading 'v' from version for tarball naming
CLEAN_VERSION="${VERSION#v}"

# Construct tarball URL
# Correct (matches your release assets)
TARBALL="${BINARY}_${CLEAN_VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${TARBALL}"

echo "📥 Checking if asset $TARBALL exists at $URL..."

if ! curl -fsI "$URL" > /dev/null; then
  echo "❌ Release asset not found:"
  echo "   $URL"
  echo "❓ Please verify the release exists and the asset is uploaded with this name."
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
