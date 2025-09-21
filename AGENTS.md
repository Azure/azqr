# AGENTS.md - Autonomous Agent Guide for Azure Quick Review (azqr)

This document provides guidance for autonomous agents contributing to the Azure Quick Review (azqr) project.

## Project Overview

Azure Quick Review (azqr) is a CLI tool written in Go that analyzes Azure resources for compliance with Azure's best practices and recommendations. The tool scans Azure resources using:
- Azure Resource Graph (ARG) queries from the Azure Proactive Resiliency Library v2 (APRL)
- Azure Resource Manager (ARM) rules built with the Azure Golang SDK

## Quick Start

### Prerequisites
- Go 1.23.3 or higher
- Valid Azure authentication (Service Principal, Managed Identity, or Azure CLI)

### Essential Commands
```bash
# Build the project
make build

# Run all tests (REQUIRED before submitting pull requests)
make test

# Clean build artifacts
make clean

# View all available targets
make help
```

## Project Structure

```
azqr/
├── cmd/azqr/               # Main CLI application entry point
│   ├── main.go            # Application entry point
│   └── commands/          # CLI command implementations (one file per Azure service)
├── cmd/server/            # Server mode implementation
├── internal/              # Internal packages
│   ├── scanner.go         # Main scanning logic
│   ├── models/            # Data models and filters
│   ├── renderers/         # Output formatters (Excel, CSV, JSON)
│   ├── scanners/          # Service-specific scanners (one per Azure service)
│   ├── graph/             # Azure Resource Graph queries
│   └── throttling/        # Rate limiting utilities
├── data/                  # Static data files
│   └── recommendations.json # Generated recommendations data
├── examples/              # Example configurations and CI/CD pipelines
├── docs/                  # Documentation website (Hugo-based)
└── Makefile              # Build automation
```

## Development Workflow

### 1. Code Style and Standards
- **Variable Naming**: Use camelCase for variable names
- **Comments**: Always add code comments using godoc style for functions
- **Authentication**: Support Service Principal, Managed Identity, and Azure CLI authentication
- **Error Handling**: Follow Go idiomatic error handling patterns

### 2. Adding Support for New Azure Services

When adding a new Azure service:

1. **Create scanner**: Add a new scanner in `internal/scanners/<service>/`
2. **Add command**: Create command file in `cmd/azqr/commands/<service>.go`
3. **Update models**: Add service-specific models to `internal/models/`
4. **Add tests**: Include comprehensive unit tests
5. **Update documentation**: Add service to supported services list in README.md

### 3. Testing Requirements

**CRITICAL**: Always run `make test` before submitting pull requests. The test command includes:
- Linting (`golangci-lint`)
- Go vet checks
- Module tidiness verification
- Unit tests with race condition detection
- Coverage reporting

```bash
# Run the full test suite
make test

# Individual test components
make lint    # Run linter
make vet     # Run go vet
make tidy    # Check module tidiness
```

### 4. Building and Distribution

```bash
# Build for current platform
make build

# Build for specific OS/architecture
GOOS=linux GOARCH=amd64 make build
GOOS=windows GOARCH=amd64 make build

# Build Docker image
make build-image

# Build with version information
PRODUCT_VERSION=1.0.0 make build
```

## Contributing Guidelines

### Pull Request Requirements

1. **Testing**: Run `make test` and ensure all tests pass
2. **Code Quality**: Follow the existing code style and patterns
3. **Documentation**: Update relevant documentation
4. **Commit Messages**: Use clear, descriptive commit messages
5. **Dependencies**: Minimize new dependencies; justify if necessary

### Supported Azure Services

The project currently supports 50+ Azure services. When adding new services, follow the established pattern:

1. Use appropriate abbreviations (see README.md for existing conventions)
2. Implement both ARM-based and ARG-based recommendations where applicable
3. Follow the scanner interface pattern
4. Include appropriate error handling and logging

### Authentication and Permissions

The tool requires `Reader` permissions over Subscription or Management Group scope and supports:
- Service Principal (environment variables: AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
- Azure Managed Identity
- Azure CLI authentication

### Output Formats

azqr generates reports in multiple formats:
- **Excel** (default): Multi-sheet workbook with recommendations, impacted resources, inventory, etc.
- **CSV**: Same data as Excel but in CSV format (use `--csv` flag)
- **JSON**: Machine-readable format

## Common Tasks for Agents

### Adding a New Recommendation Rule

1. Identify the target Azure service
2. Locate the appropriate scanner in `internal/scanners/<service>/`
3. Add the rule logic following existing patterns
4. Update tests
5. Run `make json` to update recommendations.json
6. Verify with `make test`

### Fixing Bugs

1. Try to reproduce the issue with minimal test case
2. Add regression test if missing
3. Implement fix following project patterns
4. Verify fix with `make test`
5. Update documentation if needed

### Performance Optimization

1. Use the throttling utilities in `internal/throttling/` for rate limiting
2. Implement concurrent scanning where appropriate
3. Cache expensive operations
4. Profile using Go's built-in tools

## Runtime Debugging and Troubleshooting

### Enable Debug Mode
```bash
# Set environment variable for detailed logging
export AZURE_SDK_GO_LOGGING=all

# Run with debug flag
./azqr scan --debug
```

### Common Issues
1. **Authentication**: Verify Azure credentials and permissions
2. **Rate Limiting**: Use appropriate throttling settings
3. **Memory Usage**: Monitor for large subscriptions with many resources
4. **Network Connectivity**: Ensure access to Azure APIs

## Environment Variables

```bash
# Authentication
AZURE_CLIENT_ID="<service-principal-id>"
AZURE_CLIENT_SECRET="<service-principal-secret>"  
AZURE_TENANT_ID="<tenant-id>"

# Credential Chain Configuration
AZURE_TOKEN_CREDENTIALS="dev"   # Use Azure CLI/Azure Developer CLI
AZURE_TOKEN_CREDENTIALS="prod"  # Use env vars/workload identity/managed identity

# Debugging
AZURE_SDK_GO_LOGGING="all"      # Enable detailed SDK logging
```

## File Patterns and Conventions

### Scanner Implementation
```go
// internal/scanners/<service>/<service>.go
package <service>

import (
    // Required imports
)

// Scanner interface implementation
type Scanner struct {
    // Scanner fields
}

func (s *Scanner) Scan(ctx context.Context) ([]models.Recommendation, error) {
    // Implementation
}
```

### Command Implementation
```go
// cmd/azqr/commands/<service>.go
package commands

import (
    // Required imports
)

func init() {
    // Register command with cobra
}

var <service>Cmd = &cobra.Command{
    Use:   "<service>",
    Short: "Scan <Service Name>",
    Long:  "Detailed description",
    Run:   <service>Run,
}

func <service>Run(cmd *cobra.Command, args []string) {
    // Command implementation
}
```

## Resources and References

- [Azure Proactive Resiliency Library v2 (APRL)](https://aka.ms/aprl)
- [Azure Orphaned Resources](https://github.com/dolevshor/azure-orphan-resources)
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go)
- [Project Documentation](https://azure.github.io/azqr/)
- [GitHub Issues](https://github.com/Azure/azqr/issues)
- [GitHub Discussions](https://github.com/Azure/azqr/discussions)

## Support and Community

- **Issues**: Use GitHub Issues for bug reports and feature requests
- **Discussions**: Use GitHub Discussions for questions and support
- **Security**: Report security issues following the SECURITY.md guidelines
- **Code of Conduct**: Follow the Microsoft Open Source Code of Conduct

Remember: The key to successful contribution is running `make test` before submitting any pull request!