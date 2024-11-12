---
title: Install
weight: 3
description: Learn how to install Azure Quick Review (azqr)
---

## Install on Linux or Azure Cloud Shell

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
chmod +x azqr
```

## Install on Windows

Use `winget`:

```console
winget install azqr
```

or download the executable file:

```
$latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-windows-latest-amd64.exe -OutFile azqr.exe
```

## Install on Mac

Use `homebrew`:

```console
brew install azqr
```

or download the latest release from [here](https://github.com/Azure/azqr/releases).