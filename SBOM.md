# Software Bill of Materials (SBOM)

This document describes how Azure Quick Review (azqr) publishes a Software Bill of Materials for its dependencies.

## Release Artifacts

Each [release](https://github.com/Azure/azqr/releases) publishes a Software Bill of Materials for the Go module dependencies, in both CycloneDX (`azqr.cdx.json`) and SPDX (`azqr.spdx.json`) formats, along with `.sha256` checksums, generated automatically by [anchore/sbom-action](https://github.com/anchore/sbom-action) (which wraps [Syft](https://github.com/anchore/syft)).

## Docker Image

The published Docker image (`ghcr.io/azure/azqr`) includes an SBOM and build provenance attestation (via Docker Buildx `sbom: true` / `provenance: true`), which can be inspected with:

```bash
docker buildx imagetools inspect ghcr.io/azure/azqr:latest --format '{{ json .SBOM }}'
```
