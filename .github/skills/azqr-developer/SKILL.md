---
name: azqr-developer
description: Expert guidance for developing and contributing to Azure Quick Review (azqr) - A Go-based CLI tool for Azure resource compliance analysis
---

# Azure Quick Review (azqr) Development Skill

Expert guidance for autonomous agents and developers contributing to the Azure Quick Review (azqr) project.

## Project Overview

Azure Quick Review (azqr) is a CLI tool written in Go that analyzes Azure resources for compliance with Azure's best practices and recommendations. The tool scans Azure resources using:
- **Azure Resource Graph (ARG) queries** from the Azure Proactive Resiliency Library v2 (APRL)
- **Azure Resource Manager (ARM) rules** built with the Azure Golang SDK

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

## Code Style and Standards

### Naming Conventions
- **Variables**: Use camelCase for all variable names
- **Functions**: Use MixedCaps for exported functions, mixedCaps for unexported
- **Packages**: Use lowercase, single-word package names (avoid underscores)
- **Interfaces**: Name with -er suffix when possible (e.g., `Scanner`, `Renderer`)

### Documentation
- Always add code comments using godoc style for exported functions
- Document why, not what, unless the what is complex
- Start comments with the name of the thing being described
- Write comments in complete sentences

### Authentication
All code must support multiple authentication methods:
- **Service Principal** (environment variables: AZURE_CLIENT_ID, AZURE_CLIENT_SECRET, AZURE_TENANT_ID)
- **Azure Managed Identity**
- **Azure CLI authentication**

### Error Handling
- Follow Go idiomatic error handling patterns
- Check errors immediately after the function call
- Wrap errors with context using `fmt.Errorf` with `%w` verb
- Keep error messages lowercase and don't end with punctuation
- Name error variables `err`

## Development Workflow

### 1. Adding Support for New Azure Services

When adding a new Azure service, follow this systematic approach:

1. **Create scanner** in `internal/scanners/<service>/`
   - Implement the scanner interface
   - Support both ARM-based and ARG-based recommendations
   - Include appropriate error handling and logging
   
2. **Add command** in `cmd/azqr/commands/<service>.go`
   - Follow the existing command pattern
   - Use appropriate service abbreviation (see README.md)
   
3. **Update models** in `internal/models/`
   - Add service-specific models if needed
   - Ensure models support JSON serialization
   
4. **Add comprehensive tests**
   - Include unit tests for scanner logic
   - Test both success and error cases
   - Use table-driven tests for multiple scenarios
   
5. **Update documentation**
   - Add service to supported services list in README.md
   - Document any service-specific requirements

### 2. Testing Requirements

**CRITICAL**: Always run `make test` before submitting pull requests. This is non-negotiable.

The test command includes:
- **Linting** (`golangci-lint`) - Code quality checks
- **Go vet checks** - Static analysis
- **Module tidiness verification** - Dependency management
- **Unit tests** with race condition detection
- **Coverage reporting**

```bash
# Run the full test suite (ALWAYS run before PR)
make test

# Individual test components
make lint    # Run linter
make vet     # Run go vet
make tidy    # Check module tidiness
```

### 3. Building and Distribution

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

# Update recommendations.json after adding rules
make json
```

## Common Development Tasks

### Adding a New Recommendation Rule

1. Identify the target Azure service
2. Locate the appropriate scanner in `internal/scanners/<service>/`
3. Add the rule logic following existing patterns:
   - Use consistent naming conventions
   - Include clear comments explaining the rule
   - Reference official Azure documentation
4. Update tests to cover the new rule
5. Run `make json` to update recommendations.json
6. Verify with `make test`

### Fixing Bugs

1. Reproduce the issue with a minimal test case
2. Add regression test if missing
3. Implement fix following project patterns
4. Verify fix with `make test`
5. Update documentation if the bug revealed unclear behavior

### Performance Optimization

1. Use the throttling utilities in `internal/throttling/` for rate limiting
2. Implement concurrent scanning where appropriate (use goroutines wisely)
3. Cache expensive operations when possible
4. Profile using Go's built-in tools before optimizing
5. Focus on algorithmic improvements first

## Scanner Implementation Patterns

### Scanner Interface
```go
// internal/scanners/<service>/<service>.go
package <service>

import (
    "context"
    "github.com/Azure/azqr/internal/models"
)

// Scanner implements the service scanner interface
type Scanner struct {
    // Scanner fields (config, client, etc.)
}

// Scan performs the compliance scan for the service
func (s *Scanner) Scan(ctx context.Context) ([]models.Recommendation, error) {
    // Implementation
    // 1. Fetch resources
    // 2. Apply recommendation rules
    // 3. Return findings
}
```

### Command Implementation
```go
// cmd/azqr/commands/<service>.go
package commands

import (
    "github.com/spf13/cobra"
)

func init() {
    // Register command with root command
}

var <service>Cmd = &cobra.Command{
    Use:   "<service>",
    Short: "Scan <Service Name>",
    Long:  "Detailed description of what this scanner does",
    Run:   <service>Run,
}

func <service>Run(cmd *cobra.Command, args []string) {
    // Command implementation
    // 1. Parse flags
    // 2. Initialize scanner
    // 3. Run scan
    // 4. Output results
}
```

## Testing Patterns

### Table-Driven Tests
```go
func TestScanner_Scan(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() // setup test environment
        want    int    // expected number of recommendations
        wantErr bool
    }{
        {
            name:    "success case",
            setup:   func() { /* setup */ },
            want:    5,
            wantErr: false,
        },
        {
            name:    "error case",
            setup:   func() { /* setup */ },
            want:    0,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            // test implementation
        })
    }
}
```

## Runtime Debugging and Troubleshooting

### Enable Debug Mode
```bash
# Set environment variable for detailed logging
export AZURE_SDK_GO_LOGGING=all

# Run with debug flag
./azqr scan --debug
```

### Common Issues and Solutions

1. **Authentication Failures**
   - Verify Azure credentials are set correctly
   - Check permissions (requires `Reader` on Subscription/Management Group)
   - Test with `az account show` if using Azure CLI auth

2. **Rate Limiting**
   - Use appropriate throttling settings
   - Implement exponential backoff for retries
   - Consider batching requests when possible

3. **Memory Usage**
   - Monitor for large subscriptions with many resources
   - Use streaming or pagination for large datasets
   - Profile memory usage with `pprof`

4. **Network Connectivity**
   - Ensure access to Azure APIs
   - Check firewall and proxy settings
   - Verify DNS resolution

## Environment Variables

### Authentication
```bash
# Service Principal
AZURE_CLIENT_ID="<service-principal-id>"
AZURE_CLIENT_SECRET="<service-principal-secret>"  
AZURE_TENANT_ID="<tenant-id>"

# Credential Chain Configuration
AZURE_TOKEN_CREDENTIALS="dev"   # Use Azure CLI/Azure Developer CLI
AZURE_TOKEN_CREDENTIALS="prod"  # Use env vars/workload identity/managed identity
```

### Debugging
```bash
AZURE_SDK_GO_LOGGING="all"      # Enable detailed SDK logging
```

## Contributing Guidelines

### Pull Request Requirements

1. **Testing**: Run `make test` and ensure all tests pass (100% required)
2. **Code Quality**: Follow the existing code style and patterns
3. **Documentation**: Update relevant documentation (README, code comments)
4. **Commit Messages**: Use clear, descriptive commit messages
5. **Dependencies**: Minimize new dependencies; justify if necessary

### Supported Azure Services

The project currently supports 50+ Azure services including:
- Compute: VMs, AKS, Azure Functions, App Service, etc.
- Storage: Storage Accounts, Disks, NetApp Files, etc.
- Networking: Virtual Networks, Load Balancers, Application Gateway, etc.
- Databases: SQL, Cosmos DB, PostgreSQL, MySQL, etc.
- And many more...

When adding new services:
1. Use appropriate abbreviations (see README.md for existing conventions)
2. Implement both ARM-based and ARG-based recommendations where applicable
3. Follow the scanner interface pattern
4. Include appropriate error handling and logging

### Output Formats

azqr generates reports in multiple formats:
- **Excel** (default): Multi-sheet workbook with recommendations, impacted resources, inventory, etc.
- **CSV**: Same data as Excel but in CSV format (use `--csv` flag)
- **JSON**: Machine-readable format for automation

## Key Resources and References

- [Azure Proactive Resiliency Library v2 (APRL)](https://aka.ms/aprl) - Source of ARG queries
- [Azure Orphaned Resources](https://github.com/dolevshor/azure-orphan-resources) - Orphan detection patterns
- [Azure SDK for Go](https://github.com/Azure/azure-sdk-for-go) - Official Azure Go SDK
- [Project Documentation](https://azure.github.io/azqr/) - Full documentation site
- [Effective Go](https://go.dev/doc/effective_go) - Go best practices
- [Go Code Review Comments](https://go.dev/wiki/CodeReviewComments) - Go code review guide

## Support and Community

- **Issues**: Use [GitHub Issues](https://github.com/Azure/azqr/issues) for bug reports and feature requests
- **Discussions**: Use [GitHub Discussions](https://github.com/Azure/azqr/discussions) for questions and support
- **Security**: Report security issues following the SECURITY.md guidelines
- **Code of Conduct**: Follow the Microsoft Open Source Code of Conduct

## Critical Reminders

1. **Always run `make test` before submitting a pull request** - This is the most important rule
2. Use camelCase for variable names
3. Add godoc-style comments to all exported functions
4. Support all authentication methods (Service Principal, Managed Identity, Azure CLI)
5. Follow Go idiomatic error handling
6. Keep code simple and readable
7. Update recommendations.json with `make json` after adding rules
8. Reference APRL documentation when implementing ARG-based rules
9. Test both success and error paths
10. Keep scanner implementations consistent with existing patterns
