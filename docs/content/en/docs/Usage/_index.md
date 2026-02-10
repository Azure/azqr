---
title: Usage
description: Use Azure Quick Review &mdash; to analyze Azure resources and identify whether they comply with Azure's best practices and recommendations.
weight: 3
---

## Authorization

**Azure Quick Review (azqr)** requires the following permissions:

* **Reader** over Subscription or Management Group scope (required for all scans)

## Authentication

**Azure Quick Review (azqr)** requires the following permissions:

* Reader over Subscription or Management Group scope

### Credential Chain Configuration

**Azure Quick Review (azqr)** uses the Azure SDK's `DefaultAzureCredential` which automatically selects the most appropriate credential based on your environment. By default, it tries credentials in order: environment variables, workload identity, managed identity, Azure CLI, and Azure Developer CLI.

You can customize this behavior by setting the `AZURE_TOKEN_CREDENTIALS` environment variable:

* `dev` - Prioritize Azure CLI (`az`) or Azure Developer CLI (`azd`) credentials (recommended for local development)
* `prod` - Prioritize environment variables, workload identity, or managed identity (recommended for CI/CD and production)

### Service Principal Authentication

Set the following environment variables:

Powershell:

``` powershell
$env:AZURE_CLIENT_ID = '<service-principal-client-id>'
$env:AZURE_CLIENT_SECRET = '<service-principal-client-secret>'
$env:AZURE_TENANT_ID = '<tenant-id>'
```

Bash:

``` bash
export AZURE_CLIENT_ID='<service-principal-client-id>'
export AZURE_CLIENT_SECRET='<service-principal-client-secret>'
export AZURE_TENANT_ID='<tenant-id>'
```

### Authenticate with a Managed Identity

Set the following environment variables:

Powershell:

``` powershell
$env:AZURE_CLIENT_ID = '<managed-identity-client-id>'
$env:AZURE_TENANT_ID = '<tenant-id>'
```

Bash:

``` bash
export AZURE_CLIENT_ID='<managed-identity-client-id>'
export AZURE_TENANT_ID='<tenant-id>'
```

### Authenticate with Azure CLI

Authenticate to Azure:

```console
az login
```

## Cloud Configuration

**Azure Quick Review (azqr)** supports scanning resources in different Azure cloud environments. You can configure the target cloud using environment variables.

### Predefined Cloud Environments

Set the `AZURE_CLOUD` environment variable to specify the Azure cloud environment:

**Azure Public Cloud (default):**

Powershell:
``` powershell
$env:AZURE_CLOUD = 'AzurePublic'
```

Bash:
``` bash
export AZURE_CLOUD='AzurePublic'
```

**Azure US Government Cloud:**

Powershell:
``` powershell
$env:AZURE_CLOUD = 'AzureGovernment'
```

Bash:
``` bash
export AZURE_CLOUD='AzureGovernment'
```

**Azure China Cloud:**

Powershell:
``` powershell
$env:AZURE_CLOUD = 'AzureChina'
```

Bash:
``` bash
export AZURE_CLOUD='AzureChina'
```

Supported values for `AZURE_CLOUD`:
- `AzurePublic`, `public`, or empty (default)
- `AzureGovernment`, `AzureUSGovernment`, or `usgovernment`
- `AzureChina` or `china`

### Custom Cloud Configuration

For custom or sovereign cloud environments, you can specify custom endpoints that will override the predefined cloud settings:

Powershell:
``` powershell
$env:AZURE_AUTHORITY_HOST = 'https://login.microsoftonline.custom/'
$env:AZURE_RESOURCE_MANAGER_ENDPOINT = 'https://management.custom.azure.com'
$env:AZURE_RESOURCE_MANAGER_AUDIENCE = 'https://management.core.custom.azure.com/'
```

Bash:
``` bash
export AZURE_AUTHORITY_HOST='https://login.microsoftonline.custom/'
export AZURE_RESOURCE_MANAGER_ENDPOINT='https://management.custom.azure.com'
export AZURE_RESOURCE_MANAGER_AUDIENCE='https://management.core.custom.azure.com/'
```

**Environment Variables:**
- `AZURE_AUTHORITY_HOST`: Custom Active Directory authority host (e.g., `https://login.microsoftonline.us/`)
- `AZURE_RESOURCE_MANAGER_ENDPOINT`: Custom ARM endpoint (e.g., `https://management.usgovcloudapi.net`)
- `AZURE_RESOURCE_MANAGER_AUDIENCE`: Custom ARM token audience (optional, e.g., `https://management.core.usgovcloudapi.net/`)

> **Note:** When custom endpoints are provided (both `AZURE_AUTHORITY_HOST` and `AZURE_RESOURCE_MANAGER_ENDPOINT`), they take priority over the `AZURE_CLOUD` setting.

## Scan Azure Resources with default settings

* Scan All Resources

  ```console
  azqr scan
  ```

* Scan a Management Group

  ```console
  azqr scan --management-group-id <management_group_id>
  ```

* Scan a Subscription
  
  ```console
  azqr scan --subscription-id <subscription_id>
  ```

* Scan a Resource Group

  ```console
  azqr scan --subscription-id <subscription_id> --resource-group <resource_group_name>
  ```

* Scan Multiple Subscriptions

  ```console
  azqr scan --subscription-id <sub_id_1> --subscription-id <sub_id_2>
  ```

* Scan Multiple Resource Groups

  ```console
  azqr scan --subscription-id <sub_id> --resource-group <rg_1> --resource-group <rg_2>
  ```

## Advanced Filtering

You can configure Azure Quick Review to include or exclude specific subscriptions or resource groups and also exclude services or recommendations. To do so, create a `yaml` file with the following format:

```yaml
azqr:
  include:
    subscriptions:
      - <subscription_id> # format: <subscription_id>
    resourceGroups:
      - <resource_group_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>
    resourceTypes:
      - <resource type abbreviation> # format: Abbreviation of the resource type. For example: "vm" for "Microsoft.Compute/virtualMachines"
  exclude:
    subscriptions:
      - <subscription_id> # format: <subscription_id>
    resourceGroups:
      - <resource_group_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>
    services:
      - <service_resource_id> # format: /subscriptions/<subscription_id>/resourceGroups/<resource_group_name>/providers/<service_provider>/<service_name>
    recommendations:
      - <recommendation_id> # format: <recommendation_id>
```

Then run the scan with the `--filters` flag:

```bash
./azqr scan --filters <path_to_yaml_file>
```

> Check the [rules](https://azure.github.io/azqr/docs/recommendations/) to get the recommendation ids.

> Check the [overview](https://azure.github.io/azqr/docs/overview/) to get the resource type abbreviations.

## Controlling Scan Stages

Azure Quick Review allows you to control which scan stages are executed. By default, `diagnostics`, `advisor`, and `defender` stages are enabled.

### Available Stages

- **advisor**: Azure Advisor recommendations
- **defender**: Microsoft Defender for Cloud status
- **defender-recommendations**: Microsoft Defender for Cloud recommendations
- **arc**: Azure Arc-enabled SQL Server instances
- **policy**: Azure Policy compliance states
- **cost**: Cost analysis for the last 3 months
- **diagnostics**: Diagnostic settings scan

### Stage Control Examples

```bash
# Enable specific stages (replaces defaults)
azqr scan --stages cost,policy

# Disable specific stages (keeps other defaults)
azqr scan --stages -diagnostics

# Enable all stages
azqr scan --stages advisor,defender,defender-recommendations,arc,policy,cost,diagnostics

```

> **Note**: Use stage names with the `-` prefix to disable specific stages (e.g., `-diagnostics`).

## Internal Plugins

Azure Quick Review includes specialized internal plugins for advanced analytics. Plugins can be run as standalone commands or integrated with full scans.

### Running Plugins as Standalone Commands

For fast, focused analysis, run plugins as top-level commands:

```bash
# Run OpenAI throttling analysis
azqr openai-throttling

# Run carbon emissions analysis
azqr carbon-emissions

# Run zone mapping analysis
azqr zone-mapping

# With specific subscription
azqr zone-mapping --subscription-id <sub-id>
```

### Integrating Plugins with Full Scans

Run plugins alongside standard scanning:

```bash
# Single plugin with scan
azqr scan --plugin openai-throttling

# Multiple plugins with scan
azqr scan --plugin openai-throttling --plugin carbon-emissions --plugin zone-mapping

# With other options
azqr scan --subscription-id <sub-id> --plugin zone-mapping
```

### Listing Available Plugins

View all registered plugins:

```bash
azqr plugins list
```

## View All Recommendations

You can list all available recommendations in markdown or JSON format:

```bash
# List recommendations as markdown table
azqr rules

# List recommendations as JSON
azqr rules --json
```

## File Outputs

Currently Azure Quick Review supports 3 types of file outputs: `xlsx` (default), `csv`, `json`

### xlsx

`xlsx` is the default output format.

> Check the [overview](https://azure.github.io/azqr/docs/overview/) to get the more information.

### csv

By default `azqr` will create an xlsx document, However if you need to export to `csv` you can use the following flag: `--csv`

Example:

```bash
azqr scan --csv
```

### - json

By default `azqr` will create an xlsx document, However if you need to export to `json` you can use the following flag: `--json`

Example:

```bash
azqr scan --json
```

The scan will generate a single consolidated `json` file:

``` 
<file-name>.json
```

### Changing the Output File Name

You can change the output file name by using the `--output-name` or `-o` flag:

Powershell:

```powershell
$timestamp = Get-Date -Format 'yyyyMMddHHmmss'
azqr scan --output-name "azqr_action_plan_$timestamp"
```

Bash:

```bash
timestamp=$(date '+%Y%m%d%H%M%S')
azqr scan --output-name "azqr_action_plan_$timestamp"
```

> By default, the output file name is `azqr_action_plan_YYYY_MM_DD_THHMMSS`.

### Output to STDOUT

You can output JSON results directly to stdout:

```bash
# Output JSON to stdout
azqr scan --json --stdout
```

### Masking Subscription IDs

By default, Azure Quick Review masks subscription IDs in reports for security. You can control this behavior:

```bash
# Disable masking (show full subscription IDs)
azqr scan --mask=false

# Enable masking explicitly (default)
azqr scan --mask=true
```
## Interactive Dashboard (show command)

You can explore your scan results with a lightweight embedded web UI using the `show` command. The dashboard supports both Excel and JSON report formats.

### Usage

1. Generate a report (Excel or JSON):

```bash
# Excel format (default)
azqr scan --subscription-id <subscription_id> --output-name report

# JSON format
azqr scan --subscription-id <subscription_id> --output-name report --json
```

2. Launch the dashboard:

```bash
# With Excel file
azqr show --file report.xlsx --open

# With JSON file
azqr show --file report.json --open

# On custom port
azqr show --file report.xlsx --port 3000
```

## Copilot (AI Assistant)

Azure Quick Review includes an interactive AI assistant powered by GitHub Copilot. This command starts a conversational TUI session that connects to GitHub Copilot and exposes azqr tools for natural language interaction.

### Prerequisites

1. [GitHub CLI](https://cli.github.com/) installed
2. Authenticated: `gh auth login`
3. Active GitHub Copilot subscription

### Starting the Assistant

```bash
# Start interactive AI assistant
azqr copilot

# Use a specific model (default: claude-sonnet-4.5)
azqr copilot --model claude-sonnet-4.5

# Resume a previous session
azqr copilot --resume <session-id>
```
### Available Tools

The assistant can invoke the following azqr tools:

- **scan** – Run Azure resource compliance scans
- **get-recommendations-catalog** – View the azqr recommendations catalog
- **get-supported-services** – List supported Azure services

It also has access to the [Microsoft Learn MCP server](https://learn.microsoft.com/api/mcp) for fetching official Azure documentation.

## MCP Server (Model Context Protocol)

Azure Quick Review includes a Model Context Protocol (MCP) server that enables AI assistants and tools to interact with azqr functionality. The MCP server can run in two modes:

### stdio Mode (Default)

The stdio mode is designed for integration with tools like VS Code and AI assistants that communicate via standard input/output:

```bash
# Start MCP server in stdio mode
azqr mcp
```

This mode is typically used when azqr is configured as an MCP server in your IDE or AI assistant configuration.

### HTTP/SSE Mode

The HTTP/SSE (Server-Sent Events) mode allows the MCP server to be accessed over HTTP, enabling remote access and web-based integrations:

```bash
# Start MCP server in HTTP mode on default port (:8080)
azqr mcp --mode http

# Start MCP server on a custom port
azqr mcp --mode http --addr :3000

# Start with specific host and port
azqr mcp --mode http --addr localhost:9090
```

## Debugging and Troubleshooting

### Debug Mode

Azure Quick Review supports a global `--debug` flag for troubleshooting. This flag is available for all commands:

```bash
# Enable debug logging for scan
azqr scan --debug

# Enable debug logging for plugins
azqr zone-mapping --debug
azqr openai-throttling --debug

# Combine with other flags
azqr scan --subscription-id <sub-id> --debug --stages cost
```

### Full Diagnostic Output

For comprehensive troubleshooting, combine environment variables with the debug flag:

```bash
# Enable full debugging output
export AZURE_SDK_GO_LOGGING=all
azqr scan --debug
```

### Common Issues

If you encounter any issue while using **Azure Quick Review (azqr)**:

1. Enable debug mode with `--debug` flag
2. Set `AZURE_SDK_GO_LOGGING=all` environment variable
3. Run the command and capture the output
4. Share the console output by filing a new [issue](https://github.com/Azure/azqr/issues)

## Help

You can get help for `azqr` commands by running:

```console
azqr --help
```
