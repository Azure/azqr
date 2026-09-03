#!/usr/bin/env bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.
#
# Generate NOTICE.md from third-party Go dependency licenses, grouping
# dependencies that share identical license text under a single heading
# (see hack/code/notice_gen).
#
# Requires `go-licenses` to be on PATH:
#   go install github.com/google/go-licenses/v2@latest
#
# Writes NOTICE.md in the repo root by default, or to the path supplied
# as the first argument.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../../.."

OUTPUT="${1:-NOTICE.md}"

if ! command -v go-licenses >/dev/null 2>&1; then
    echo "Error: go-licenses not found on PATH" >&2
    echo "Install with: go install github.com/google/go-licenses/v2@latest" >&2
    exit 1
fi

STD_PACKAGES="$(go list std | tr '\n' ',' | sed 's/,$//')"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

# Raw template emitting one delimiter-separated record per dependency:
# \x1e (record separator) name \x1f version \x1f license-name \x1f license-text
RAW_TEMPLATE="${WORKDIR}/raw.tmpl"
printf '{{ range . }}%s{{.Name}}%s{{.Version}}%s{{.LicenseName}}%s{{.LicenseText}}{{ end }}' \
    $'\x1e' $'\x1f' $'\x1f' $'\x1f' > "${RAW_TEMPLATE}"

RAW_RECORDS="${WORKDIR}/raw.records"
go-licenses report ./cmd/... \
    --ignore github.com/Azure/azqr \
    --ignore "${STD_PACKAGES}" \
    --template "${RAW_TEMPLATE}" \
    > "${RAW_RECORDS}"

go run ./hack/code/notice_gen "${RAW_RECORDS}" "${OUTPUT}"

echo "Wrote ${OUTPUT}"
