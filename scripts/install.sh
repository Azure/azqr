#!/bin/bash

PREVIEW=false

usage() {
    echo "Usage: $0 [--preview]"
    echo "  --preview    Install the latest preview (pre-release) version"
    echo "  --help       Show this help message"
    exit 0
}

while [[ $# -gt 0 ]]; do
    case $1 in
        --preview)
            PREVIEW=true
            shift
            ;;
        --help|-h)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

if ! command -v jq &> /dev/null || ! command -v unzip &> /dev/null || ! command -v wget &> /dev/null
then
    echo "jq, unzip or wget could not be found, please install them."
    exit
fi

arch=$(uname -m)
if [ "$arch" == "aarch64" ]; then
    arch="arm64"
else
    arch="amd64"
fi

if [ "$PREVIEW" = true ]; then
    latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases | jq -r '[.[] | select(.prerelease == true)] | first | .tag_name')
    if [ "$latest_azqr" == "null" ] || [ -z "$latest_azqr" ]; then
        echo "No preview version available."
        exit 1
    fi
    echo "Installing preview version: $latest_azqr"
else
    latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
fi

wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-linux-$arch.zip -O azqr.zip
unzip -uj -qq azqr.zip
rm azqr.zip
chmod +x azqr
./azqr --version
