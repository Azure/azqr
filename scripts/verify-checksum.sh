#!/bin/bash

# Script to verify azqr binary checksum
# Usage: ./scripts/verify-checksum.sh <version> <platform>
# Example: ./scripts/verify-checksum.sh 2.7.3 windows-amd64

set -e

VERSION=$1
PLATFORM=$2

if [ -z "$VERSION" ] || [ -z "$PLATFORM" ]; then
    echo "Usage: $0 <version> <platform>"
    echo "Examples:"
    echo "  $0 2.7.3 windows-amd64"
    echo "  $0 2.7.3 windows-arm64"
    echo "  $0 2.7.3 linux-amd64"
    echo "  $0 2.7.3 linux-arm64"
    echo "  $0 2.7.3 darwin-amd64"
    echo "  $0 2.7.3 darwin-arm64"
    exit 1
fi

# Define file extension based on platform
if [[ "$PLATFORM" == windows-* ]]; then
    FILE_EXT=".exe"
else
    FILE_EXT=""
fi

# Define URLs
BASE_URL="https://github.com/Azure/azqr/releases/download/v.${VERSION}"
FILE_NAME="azqr-${PLATFORM}${FILE_EXT}"
CHECKSUM_NAME="${FILE_NAME}.sha256"

echo "Verifying azqr ${VERSION} for ${PLATFORM}..."
echo "File: ${FILE_NAME}"
echo "Checksum: ${CHECKSUM_NAME}"

# Check if files exist locally
if [ ! -f "$FILE_NAME" ]; then
    echo "Error: ${FILE_NAME} not found in current directory"
    echo "Please download it first:"
    echo "  curl -L ${BASE_URL}/${FILE_NAME} -o ${FILE_NAME}"
    exit 1
fi

# Download checksum if not exists
if [ ! -f "$CHECKSUM_NAME" ]; then
    echo "Downloading checksum file..."
    echo " curl -sL ${BASE_URL}/${CHECKSUM_NAME} -o $CHECKSUM_NAME"
    curl -sL "${BASE_URL}/${CHECKSUM_NAME}" -o "$CHECKSUM_NAME"
fi

# Verify checksum
echo "Verifying checksum..."
if command -v sha256sum >/dev/null 2>&1; then
    # Linux/Unix
    sha256sum -c "$CHECKSUM_NAME"
elif command -v shasum >/dev/null 2>&1; then
    # macOS
    shasum -a 256 -c "$CHECKSUM_NAME"
else
    echo "Error: No suitable checksum command found (sha256sum or shasum)"
    exit 1
fi

echo "✅ Checksum verification successful!"
echo "The binary ${FILE_NAME} is authentic and has not been tampered with."