# Azure Quick Review Installation Guide

## Installing Azure Quick Review

This guide provides step-by-step instructions on how to install Azure Quick Review (azqr) on different operating systems, including Linux, Windows, and Mac.

### Prerequisites

Before you begin the installation, ensure you have the following prerequisites:

- **For Windows**: Ensure you have `winget` or the ability to download executable files.
- **For Linux**: Ensure you have `curl` installed.
- **For Mac**: Ensure you have `homebrew` or `curl` installed.

### Installation on Linux or Azure Cloud Shell (Bash)

1. Open your terminal.
2. Run the following command to download the latest version of Azure Quick Review:

   ```bash
   latest_azqr=$(curl -sL https://api.github.com/repos/Azure/azqr/releases/latest | jq -r ".tag_name" | cut -c1-)
   wget https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-ubuntu-latest-amd64 -O azqr
   ```

3. Make the downloaded file executable:

   ```bash
   chmod +x azqr
   ```

### Installation on Windows

1. Open your command prompt or PowerShell.
2. You can install Azure Quick Review using `winget` by running:

   ```bash
   winget install azqr
   ```

   Alternatively, you can download the executable file directly:

   ```bash
   $latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
   iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-windows-latest-amd64.exe -OutFile azqr.exe
   ```

### Installation on Mac

1. Open your terminal.
2. You can install Azure Quick Review using `homebrew` by running:

   ```bash
   brew install azqr
   ```

   Alternatively, you can download the latest release from [here](https://github.com/Azure/azqr/releases).

### Verification

After installation, you can verify that Azure Quick Review is installed correctly by running:

   ```bash
   ./azqr -h
   ```

This command should display the help information for Azure Quick Review, confirming that the installation was successful.
