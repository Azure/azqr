# Windows Defender ASR Rules Compatibility

## Overview

This document outlines improvements made to azqr.exe to reduce false positives with Windows Defender's Attack Surface Reduction (ASR) rules.

## Changes Made for Version 2.7.4+

### 1. Windows Resource Information
- Added proper Windows PE resource metadata including:
  - Company Name: Microsoft Corporation
  - Product Name: Azure Quick Review
  - File Description: Azure Resource Compliance Scanner
  - Version Information: Embedded in PE header
  - Legal Copyright: MIT License attribution

### 2. Build Process Improvements
- **Reduced Aggressive Symbol Stripping**: Removed `-w` and `-buildid=` flags that can trigger false positives
- **Added Proper Manifest**: Windows application manifest with execution level and compatibility flags
- **Trimpath Support**: Removes local file system paths for better security without triggering ASR
- **Static Linking**: Maintains static linking while improving binary reputation

### 3. ASR Rule Mitigation Strategies

#### Build-time Mitigations:
- Preserve essential PE metadata for Windows Defender reputation
- Include Microsoft Corporation as CompanyName for better trust signals
- Add proper file version resources that ASR rules validate
- Use less aggressive linker flags specific to Windows builds

#### Runtime Mitigations:
- Application manifest requests minimal privileges (`asInvoker`)
- Long path awareness enabled for better Windows compatibility
- Segment heap usage for improved memory management
- Printer driver isolation for enhanced security

## Technical Details

### Resource Files
- `cmd/azqr/winres/winres.json`: Configuration for Windows resources
- `cmd/azqr/rsrc_windows_*.syso`: Generated Windows resource objects (automatically embedded)

### Build Flags Comparison

#### Before (Triggering ASR):
```bash
LDFLAGS := -w -X github.com/Azure/azqr/cmd/azqr/commands.version=$(VERSION) -buildid= -extldflags="-static"
```

#### After (ASR-Compatible):
```bash
LDFLAGS := -X github.com/Azure/azqr/cmd/azqr/commands.version=$(VERSION) -extldflags="-static"
BUILD_TAGS := -tags="netgo,osusergo" -buildmode=exe
TRIM_PATH := -trimpath
```

## Testing ASR Compatibility

To test if your azqr.exe binary triggers ASR rules:

1. Enable Windows Defender ASR rules in audit mode
2. Run the binary in a controlled environment
3. Check Windows Security event logs for ASR triggers
4. Verify binary metadata using tools like `sigcheck.exe`

## For Developers

When building Windows binaries locally:

```bash
GOOS=windows GOARCH=amd64 make
```

The Makefile automatically applies Windows-specific build settings that include:
- Embedded Windows resources
- ASR-compatible linker flags
- Proper PE metadata

## Troubleshooting

If you still experience ASR rule blocks:

1. **Check ASR Rule Configuration**: Some organizations have very strict ASR policies
2. **Verify Binary Signature**: Consider code signing for enterprise environments
3. **Whitelist the Application**: Add azqr.exe to ASR rule exceptions
4. **Update Windows Defender**: Ensure you have the latest definitions

## References

- [Windows Defender Attack Surface Reduction Rules](https://docs.microsoft.com/en-us/windows/security/threat-protection/microsoft-defender-atp/attack-surface-reduction)
- [PE File Format and Resources](https://docs.microsoft.com/en-us/windows/win32/debug/pe-format)
- [Go Build Modes](https://golang.org/cmd/go/#hdr-Build_modes)