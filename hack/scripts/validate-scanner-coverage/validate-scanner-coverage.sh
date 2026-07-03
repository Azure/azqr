#!/bin/bash
# Copyright (c) Microsoft Corporation.
# Licensed under the MIT License.

# This script validates that all service types in APRL with both YAML and KQL files
# are registered in the scanners registry (internal/scanners/registry/scanners.go)

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../../.." && pwd)"

cd "$PROJECT_ROOT"

echo "Validating scanner coverage for APRL, AZQR, and orphan-resources services..."

# Create temporary files
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

ALL_SERVICES="$TEMP_DIR/all_services.txt"
SCANNER_TYPES="$TEMP_DIR/scanner_types.txt"
MISSING_SERVICES="$TEMP_DIR/missing_services.txt"

# Collect services with YAML and KQL files from a path-based directory.
# Path structure: <root>/<Namespace>/<ResourceType>/recommendations.yaml
#                 <root>/<Namespace>/<ResourceType>/kql/*.kql
collect_path_services() {
    local root_dir="$1"
    find "$root_dir" -name "recommendations.yaml" -type f | while read yaml_file; do
        service_dir=$(dirname "$yaml_file")
        if [ -d "$service_dir/kql" ] && [ "$(ls -A "$service_dir/kql"/*.kql 2>/dev/null | wc -l)" -gt 0 ]; then
            service_path=$(echo "$service_dir" | sed "s|^${root_dir}/||")
            namespace=$(echo "$service_path" | cut -d'/' -f1)
            resource_type=$(echo "$service_path" | cut -d'/' -f2)
            case "$namespace" in
                "Oracledatabase") azure_namespace="Oracle.Database" ;;
                *) azure_namespace="Microsoft.$namespace" ;;
            esac
            case "$resource_type" in
                "cloudexadatainfrastructures") resource_type="cloudExadataInfrastructures" ;;
                "cloudexadatavmclusters") resource_type="cloudVmClusters" ;;
            esac
            echo "$service_path|$azure_namespace/$resource_type"
        fi
    done
}

# Collect orphan-resources services. Resource types are declared inside the YAML
# rather than derived from the directory path.
# Path structure: <root>/<Namespace>/queries.yaml  +  kql/*.kql
collect_orphan_services() {
    find internal/graph/azure-orphan-resources -name "queries.yaml" -type f | while read yaml_file; do
        service_dir=$(dirname "$yaml_file")
        namespace=$(basename "$service_dir")
        if [ -d "$service_dir/kql" ] && [ "$(ls -A "$service_dir/kql"/*.kql 2>/dev/null | wc -l)" -gt 0 ]; then
            grep "recommendationResourceType:" "$yaml_file" | \
                sed 's/.*recommendationResourceType: //' | tr -d '\r' | sort -u | \
                while read resource_type; do
                    [ -n "$resource_type" ] && echo "orphan/$namespace|$resource_type"
                done
        fi
    done
}

{
    collect_path_services "internal/graph/aprl/azure-resources"
    collect_path_services "internal/graph/azqr/azure-resources"
    collect_orphan_services
} | sort -u > "$ALL_SERVICES"

# Extract resource types from scanners.go (case-insensitive)
grep -oE '"Microsoft\.[^"]+"|"Oracle\.[^"]+"|"Specialized\.Workload/[^"]+"' internal/scanners/registry/scanners.go | \
    tr -d '"' | sort -u > "$SCANNER_TYPES"

# Resource types that are intentionally excluded from scanner coverage checks.
# Add entries here (one per line, exact case) when a type is known to not need
# a dedicated scanner entry.
KNOWN_EXCEPTIONS="
Microsoft.Resources/Resources
"

# Find missing services
> "$MISSING_SERVICES"
while IFS='|' read -r service_path resource_type; do
    if echo "$KNOWN_EXCEPTIONS" | grep -qi "^${resource_type}$"; then
        continue
    fi
    if ! grep -qi "^${resource_type}$" "$SCANNER_TYPES"; then
        echo "$service_path|$resource_type" >> "$MISSING_SERVICES"
    fi
done < "$ALL_SERVICES"

# Display results
total_services=$(wc -l < "$ALL_SERVICES")
missing_count=$(wc -l < "$MISSING_SERVICES")
present_count=$((total_services - missing_count))

if [ "$missing_count" -eq 0 ]; then
    echo -e "✅ All APRL services are registered in scanners.go${NC}"
    echo "Total services with YAML and KQL: $total_services"
    echo "All services are present in scanners.go"
    exit 0
else
    echo -e "${RED}✗ Found $missing_count service(s) with YAML and KQL files but NOT registered in scanners.go:${NC}"
    echo ""
    
    while IFS='|' read -r service_path resource_type; do
        echo -e "  ${YELLOW}✗${NC} $service_path"
        echo "     Resource Type: $resource_type"
    done < "$MISSING_SERVICES"
    
    echo ""
    echo "Summary:"
    echo "  Total services with YAML and KQL: $total_services"
    echo -e "  ${RED}Missing from scanners.go: $missing_count${NC}"
    echo -e "  ${GREEN}Present in scanners.go: $present_count${NC}"
    echo ""
    echo "Please add the missing services to internal/scanners/registry/scanners.go"
    exit 1
fi
