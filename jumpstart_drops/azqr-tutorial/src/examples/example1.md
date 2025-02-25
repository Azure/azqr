# Example Scenario for Using Azure Quick Review

This document provides an example scenario demonstrating how to effectively use Azure Quick Review (azqr) to scan Azure resources for compliance with best practices.

## Scenario: Scanning All Resources in a Subscription

In this example, we will scan all resources within a specific Azure subscription to identify any non-compliant configurations.

### Prerequisites

- Ensure that Azure Quick Review is installed on your system. Refer to the installation guide in `install.md` for detailed instructions.
- You must have the necessary permissions to access the Azure subscription you wish to scan.

### Step 1: Authentication

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

### Step 2: Run the Scan

To scan all resources in your subscription, execute the following command in your terminal:

```
./azqr scan -s <subscription_id>
```

Replace `<subscription_id>` with your actual Azure subscription ID.

### Step 3: Review the Output

After the scan completes, Azure Quick Review generates an output file (by default in Excel format) containing several sheets:

- **Recommendations**: Lists all recommendations and the number of impacted resources.
- **ImpactedResources**: Details all resources that have issues.
- **Inventory**: Provides a comprehensive list of all scanned resources.

### Example Command and Expected Output

**Command:**

```
./azqr scan -s 12345678-1234-1234-1234-123456789abc
```

**Expected Output:**
- An Excel file named `azqr_action_plan_YYYY_MM_DD_Thhmmss` will be generated in the current directory.

### Conclusion

This example demonstrates how to use Azure Quick Review to scan an Azure subscription for compliance. For further details on filtering recommendations or troubleshooting, refer to the `usage.md` and `troubleshooting` sections in the documentation.
