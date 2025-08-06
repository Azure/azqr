# Binary Verification

This document provides guidance on verifying the authenticity of Azure Quick Review (azqr) binaries.

## Checksum Verification

Each release includes SHA256 checksums. Verify your download:

### Using our verification script (recommended)

```bash
# Download and run the verification script
curl -sL https://raw.githubusercontent.com/Azure/azqr/main/scripts/verify-checksum.sh -o verify-checksum.sh
chmod +x verify-checksum.sh
./verify-checksum.sh 2.7.3 win-amd64
```

### Manual verification

```bash
# Download the checksum file
curl -sL https://github.com/Azure/azqr/releases/download/v<version>/azqr-win-amd64.zip.sha256 -o azqr-win-amd64.zip.sha256

# Verify the checksum (Windows)
CertUtil -hashfile azqr-win-amd64.zip SHA256

# Verify the checksum (Linux/macOS)
sha256sum -c azqr-win-amd64.zip.sha256
```

## Download Source

Only download from the official [GitHub releases page](https://github.com/Azure/azqr/releases).

## Release Integrity

Check that the release is signed by Azure/azqr maintainers on GitHub.

## Build Integrity

All binaries are built using GitHub Actions with:
- Reproducible build environment
- Pinned dependencies
- Public build logs
- Automated testing

You can verify the build process by:
1. Checking the [build workflow](.github/workflows/build.yml)
2. Reviewing the [build logs](https://github.com/Azure/azqr/actions/workflows/build.yml)
3. Comparing source code with the tagged release