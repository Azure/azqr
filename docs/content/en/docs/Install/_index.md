---
title: Install
weight: 2
description: Learn how to install Azure Quick Review (azqr)
---

## Install on Linux or Azure Cloud Shell

```bash
bash -c "$(curl -fsSL https://raw.githubusercontent.com/azure/azqr/main/scripts/install.sh)"
```

## Install on Windows

Use `winget`:

``` console
winget install azqr
```

or download the executable file:

```
Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://raw.githubusercontent.com/azure/azqr/main/scripts/install.ps1'))
```

## Install on Mac

Use `homebrew`:

```console
brew install azqr
```

or download the latest release from [here](https://github.com/Azure/azqr/releases).