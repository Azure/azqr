#!/bin/bash

if ! command -v jq &> /dev/null || ! command -v wget &> /dev/null
then
    echo "jq or wget could not be found, please install them."
    exit
fi

arch=$(uname -m)
if [ "$arch" == "aarch64" ]; then
    arch="arm64"
else
    arch="amd64"
fi

latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-linux-$arch -O azqr
chmod +x azqr
./azqr --version
