# Azure Quick Review Tutorial

Welcome to the Azure Quick Review Tutorial: a comprehensive guide on how to install and use the Azure Quick Review (azqr) tool effectively. Azure Quick Review is a command-line interface (CLI) tool designed to analyze Azure resources and ensure compliance with best practices.

## Table of Contents

- [Installation](#installing-azure-quick-review) Step-by-step instructions for installing Azure Quick Review on various operating systems.
- [Usage](#azure-quick-review-usage) Detailed guide on how to use Azure Quick Review after installation, including running scans and filtering recommendations.
- [Examples](#scenario-scanning-all-resources-in-a-subscription) Practical examples demonstrating how to use Azure Quick Review effectively with sample commands and expected outputs.

## Purpose

The purpose of this tutorial is to help users understand how to set up and use Azure Quick Review to analyze their Azure resources. By following the instructions provided, users will be able to identify non-compliant configurations and areas for improvement in their Azure environment.

## Getting Started

To get started, navigate to the installation section to set up Azure Quick Review on your preferred operating system. Once installed, proceed to the usage section to learn how to run scans and interpret the results.

## Repository

Azure Quick Review is an open-source project hosted on GitHub. You can find the source code and contribute to the project [here](https://github.com/azure/azqr).

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

## Azure Quick Review Usage

After installing Azure Quick Review, you can start using it to scan your Azure resources. Here are the basic commands to run scans:

### Authentication

Before running the scan, you need to authenticate with Azure. You can use one of the following methods:

1. **Service Principal**: Set the following environment variables:
   - `AZURE_CLIENT_ID`
   - `AZURE_CLIENT_SECRET`
   - `AZURE_TENANT_ID`

2. **Azure Managed Identity**: If running in an Azure environment that supports managed identities.

3. **Azure CLI**: Ensure you are logged in using the Azure CLI by running:

    ```bash
    az login
    ```

### Scan All Resources

To scan all resources in all subscriptions, use the following command:

```bash
./azqr scan
```

### Scan Specific Management Group

To scan all resources in a specific management group, run:

```bash
./azqr scan --management-group-id <management_group_id>
```

### Scan Specific Subscription

To scan all resources in a specific subscription, execute:

```bash
./azqr scan -s <subscription_id>
```

### Scan Specific Resource Group

To scan a specific resource group within a subscription, use:

```bash
./azqr scan -s <subscription_id> -g <resource_group_name>
```

## Filtering Recommendations

You can customize your scans by filtering specific subscriptions, resource groups, services, or recommendations. To do this, create a YAML file with the following structure:

```yaml
azqr:
  include:
    subscriptions:
      - <subscription_id>
    resourceGroups:
      - <resource_group_resource_id>
    resourceTypes:
      - <resource type abbreviation>
  exclude:
    subscriptions:
      - <subscription_id>
    resourceGroups:
      - <resource_group_resource_id>
    services:
      - <service_resource_id>
    recommendations:
      - <recommendation_id>
```

### Running Scan with Filters

Once you have your YAML file ready, run the scan with the `--filters` flag:

```bash
./azqr scan --filters <path_to_yaml_file>
```

## Interpreting the Output

The output generated by Azure Quick Review is saved by default in an Excel file. The output includes several sheets:

- **Recommendations**: Lists all recommendations and the number of impacted resources.
- **ImpactedResources**: Details all resources that are impacted.
- **ResourceTypes**: Lists the types of impacted resources.
- **Inventory**: Provides details of all scanned resources, including SKU, Tier, and calculated SLA.
- **Advisor**: Contains recommendations from Azure Advisor.
- **DefenderRecommendations**: Lists recommendations from Microsoft Defender for Cloud.
- **OutOfScope**: Details resources that were not scanned.
- **Defender**: Lists Microsoft Defender for Cloud plans and their tiers.
- **Costs**: Shows costs associated with the scanned subscription for the last three months.

## Additional Help

For more information on available commands and help, run:

```bash
./azqr -h
```

This command will provide you with a list of all available options and their descriptions.

## Example Scenario for Using Azure Quick Review

This document provides an example scenario demonstrating how to effectively use Azure Quick Review (azqr) to scan Azure resources for compliance with best practices.

### Scenario: Scanning All Resources in a Subscription

In this example, we will scan all resources within a specific Azure subscription to identify any non-compliant configurations.

#### Prerequisites

- Ensure that Azure Quick Review is installed on your system. Refer to the installation guide in `install.md` for detailed instructions.
- You must have the necessary permissions to access the Azure subscription you wish to scan.

#### Step 1: Authentication

Before running the scan, you need to authenticate with Azure. You can use one of the following methods:

1. **Service Principal**: Set the following environment variables:
   - `AZURE_CLIENT_ID`
   - `AZURE_CLIENT_SECRET`
   - `AZURE_TENANT_ID`

2. **Azure Managed Identity**: If running in an Azure environment that supports managed identities.

3. **Azure CLI**: Ensure you are logged in using the Azure CLI by running:

   ```
   az login
   ```

> **Note**: **azqr** requires Reader permissions over the subscription or management group scope.

#### Step 2: Run the Scan

To scan all resources in your subscription, execute the following command in your terminal:

```
./azqr scan -s <subscription_id>
```

Replace `<subscription_id>` with your actual Azure subscription ID.

#### Step 3: Review the Output

After the scan completes, Azure Quick Review generates an output file (by default in Excel format) containing several sheets:

- **Recommendations**: Lists all recommendations and the number of impacted resources.
- **ImpactedResources**: Details all resources that have issues.
- **Inventory**: Provides a comprehensive list of all scanned resources.

#### Example Command and Expected Output

**Command:**

```
./azqr scan -s 12345678-1234-1234-1234-123456789abc
```

**Expected Output:**
- An Excel file named `azqr_action_plan_YYYY_MM_DD_Thhmmss.xlsx` will be generated in the current directory.

#### Conclusion

This example demonstrates how to use Azure Quick Review to scan an Azure subscription for compliance. For further details on filtering recommendations or troubleshooting, refer to the `usage.md` and `troubleshooting` sections in the documentation.
