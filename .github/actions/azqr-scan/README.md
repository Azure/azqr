# Azure Quick Review (azqr) GitHub Action

This GitHub Action allows you to integrate [Azure Quick Review (azqr)](https://github.com/Azure/azqr) scans into your CI/CD workflows. It automatically downloads and runs azqr on Linux, Windows, and macOS runners.

## Features

- ✅ **Cross-platform support**: Works on Linux, Windows, and macOS runners
- ✅ **Flexible versioning**: Use latest version or specify a particular release
- ✅ **Multiple output formats**: JSON, CSV, XLSX, HTML
- ✅ **Artifact upload**: Automatically uploads scan reports as workflow artifacts
- ✅ **Customizable**: Support for custom arguments and working directories

## Quick Start

### 1. Set up Azure Authentication

First, configure Azure authentication in your repository. Create an Azure service principal and add the credentials as repository secrets:

```bash
az ad sp create-for-rbac --name "github-azqr-scanner" \
  --role "Reader" \
  --scopes "/subscriptions/{subscription-id}" \
  --sdk-auth
```

Add the output as a repository secret named `AZURE_CREDENTIALS`.

### 2. Basic Usage

```yaml
name: Azure Security Scan
on: [push, pull_request]

jobs:
  azqr-scan:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Azure Login
      uses: azure/login@v1
      with:
        creds: ${{ secrets.AZURE_CREDENTIALS }}
    
    - name: Run Azure Quick Review
      uses: Azure/azqr/.github/actions/azqr-scan
      with:
        subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
        output-format: 'json'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `version` | Version of azqr to use (e.g., `v1.12.4`, `latest`) | No | `latest` |
| `subscription-id` | Azure subscription ID to scan | No | |
| `resource-group` | Specific resource group to scan | No | |
| `output-format` | Output format (`json`, `csv`, `xlsx`) | No | `json` |
| `output-path` | Output file path (without extension) | No | `azqr-report` |
| `extra-args` | Additional arguments to pass to azqr | No | |
| `working-directory` | Working directory to run azqr from | No | `.` |

## Outputs

| Output | Description |
|--------|-------------|
| `report-path` | Path to the generated report file |
| `exit-code` | Exit code of the azqr scan |

## Advanced Usage Examples

### Scan Specific Resource Group

```yaml
- name: Scan Production Resource Group
  uses: Azure/azqr/.github/actions/azqr-scan
  with:
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
    resource-group: 'rg-production'
    output-path: 'prod-security-report'
```

### Multiple Output Formats

```yaml
- name: Generate Multiple Reports
  uses: Azure/azqr/.github/actions/azqr-scan
  with:
    subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
    output-format: 'json'
    extra-args: '--csv'
```

### Scheduled Scans

```yaml
name: Weekly Security Scan
on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9 AM UTC

jobs:
  security-audit:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Azure Login
      uses: azure/login@v1
      with:
        creds: ${{ secrets.AZURE_CREDENTIALS }}
    
    - name: Weekly Azure Security Review
      id: scan
      uses: Azure/azqr/.github/actions/azqr-scan
      with:
        subscription-id: ${{ secrets.AZURE_SUBSCRIPTION_ID }}
```

## Troubleshooting

### Common Issues

1. **Authentication failures**: Ensure your Azure credential has the necessary permissions (at least `Reader` role).

2. **Binary download failures**: Check if the specified version exists in the [azqr releases](https://github.com/Azure/azqr/releases).

## Development

To contribute to this action:

1. Fork the repository
2. Make your changes
3. Test with different operating systems using the matrix strategy
4. Submit a pull request

## Security Considerations

- Store Azure credentials securely using GitHub repository secrets
- Use least-privilege access for the Azure credentials
- Regularly rotate Azure credentials
- Review scan reports for sensitive information before sharing

## License

This action is released under the MIT License. See [LICENSE](../../LICENSE) for details.

## Related Resources

- [Azure Quick Review (azqr) Repository](https://github.com/Azure/azqr)
- [Azure Login Action](https://github.com/Azure/login)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)