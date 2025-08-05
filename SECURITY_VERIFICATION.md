# Security and Binary Verification

This document provides guidance on verifying the authenticity and security of Azure Quick Review (azqr) binaries.

## Antivirus False Positives

Some antivirus software may flag azqr binaries as potentially malicious. This is a known issue with unsigned Go binaries and is typically a false positive.

### Verification Steps

Before adding antivirus exceptions, always verify the authenticity of the binary:

1. **Download Source**: Only download from the official [GitHub releases page](https://github.com/Azure/azqr/releases)

2. **Checksum Verification**: Each release includes SHA256 checksums. Verify your download:
   ```bash
   # Download the checksum file
   curl -sL https://github.com/Azure/azqr/releases/download/v<version>/azqr-win-amd64.zip.sha256 -o azqr-win-amd64.zip.sha256
   
   # Verify the checksum (Windows)
   CertUtil -hashfile azqr-win-amd64.zip SHA256
   
   # Verify the checksum (Linux/macOS)
   sha256sum -c azqr-win-amd64.zip.sha256
   ```

3. **Release Integrity**: Check that the release is signed by Azure/azqr maintainers on GitHub

### Reporting False Positives

If you encounter a false positive:

1. **Windows Defender**: Submit to [Microsoft's malware analysis service](https://www.microsoft.com/en-us/wdsi/filesubmission)
   - Select "I think this file is clean"
   - Provide the file details and explain it's a legitimate Azure CLI tool
   - Reference this GitHub repository as the official source

2. **Other Antivirus**: Check your vendor's false positive reporting process

### Current Known Issues

- **Windows Defender**: Some versions may flag azqr binaries as "Trojan:Script/Sabsik.FLA!ml"
  - This is a false positive commonly seen with unsigned Go binaries
  - We are working with Microsoft to resolve this detection issue
  - The binary is safe when downloaded from official GitHub releases

### Code Signing Status

- **Current**: Windows binaries are not yet code-signed
- **Future**: We are working on implementing code signing to reduce false positives
- **macOS**: Binaries are built on GitHub's macOS runners and signed with Apple's certificates
- **Linux**: Distributed as unsigned binaries (standard for Linux CLI tools)

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

## Security Reporting

If you discover a security vulnerability, please follow our [Security Policy](SECURITY.md).

For questions about binary authenticity or security, please [open an issue](https://github.com/Azure/azqr/issues).