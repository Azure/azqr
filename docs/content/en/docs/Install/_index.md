---
title: Install
weight: 2
description: Learn how to install Azure Quick Review (azqr)
---

## Install on Linux or Azure Cloud Shell

```bash
latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
chmod +x azqr
```

## Install on Windows

```console
winget install azqr
```

## Install on Mac

Download the latest release from [here](https://github.com/Azure/azqr/releases).