---
title: Install
weight: 2
description: Learn how to install Azure Quick Review (azqr)
---

## Install on Linux or Azure Cloud Shell

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-linux-amd64.zip -O azqr.zip
unzip -uj -qq azqr.zip
rm azqr.zip
chmod +x azqr
```

> For ARM64 architecture, use `azqr-linux-arm64.zip` instead of `azqr-linux-amd64.zip`.

## Install on Windows

Use `winget`:

``` console
winget install azqr
```

or download the executable file:

``` console
$latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-win-amd64.zip -OutFile azqr.zip
Expand-Archive -Path azqr.zip -DestinationPath ./azqr_bin
Get-ChildItem -Path ./azqr_bin -Recurse -File | ForEach-Object { Move-Item -Path $_.FullName -Destination . -Force }
Remove-Item -Path ./azqr_bin -Recurse -Force
Remove-Item -Path azqr.zip
```

> For ARM64 architecture, use `azqr-win-arm64.zip` instead of `azqr-win-amd64.zip`.

## Install on Mac

Use `homebrew`:

```console
brew install azqr
```

or download the latest release from [here](https://github.com/Azure/azqr/releases).