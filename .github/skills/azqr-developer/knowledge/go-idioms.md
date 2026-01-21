# Go Idioms and Best Practices for azqr

This document contains Go-specific idioms and best practices relevant to the azqr project.

## Error Handling Patterns

### Standard Error Handling
```go
// Check errors immediately
result, err := someFunction()
if err != nil {
    return nil, fmt.Errorf("failed to do something: %w", err)
}
```

### Creating Context-Rich Errors
```go
// Wrap errors with context
if err := scanner.Scan(ctx); err != nil {
    return fmt.Errorf("scanning %s failed: %w", serviceName, err)
}
```

### Sentinel Errors
```go
// Define package-level error variables for specific error types
var (
    ErrNotFound = errors.New("resource not found")
    ErrInvalidConfig = errors.New("invalid configuration")
)

// Check using errors.Is
if errors.Is(err, ErrNotFound) {
    // handle not found case
}
```

## Concurrency Patterns

### Safe Goroutine Usage
```go
// Always use WaitGroup or channels to synchronize
var wg sync.WaitGroup
for _, resource := range resources {
    wg.Add(1)
    go func(r Resource) {
        defer wg.Done()
        // process resource
    }(resource)
}
wg.Wait()
```

### Context for Cancellation
```go
// Always respect context cancellation
func (s *Scanner) Scan(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // continue processing
    }
}
```

## Azure SDK Patterns

### Client Initialization
```go
// Use credential chain for authentication
cred, err := azidentity.NewDefaultAzureCredential(nil)
if err != nil {
    return fmt.Errorf("failed to create credential: %w", err)
}

client, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
if err != nil {
    return fmt.Errorf("failed to create client: %w", err)
}
```

### Pagination
```go
// Handle pagination properly
pager := client.NewListPager(nil)
for pager.More() {
    page, err := pager.NextPage(ctx)
    if err != nil {
        return fmt.Errorf("failed to get next page: %w", err)
    }
    for _, item := range page.Value {
        // process item
    }
}
```

## Testing Patterns

### Mock Interfaces
```go
// Define interfaces for testability
type ResourceClient interface {
    Get(ctx context.Context, id string) (*Resource, error)
    List(ctx context.Context) ([]*Resource, error)
}

// Use mocks in tests
type mockClient struct {
    resources []*Resource
    err       error
}

func (m *mockClient) List(ctx context.Context) ([]*Resource, error) {
    return m.resources, m.err
}
```

### Table-Driven Tests with Subtests
```go
func TestScanner_CheckCompliance(t *testing.T) {
    tests := []struct {
        name     string
        resource Resource
        want     []Recommendation
    }{
        {"compliant resource", Resource{/* ... */}, nil},
        {"non-compliant resource", Resource{/* ... */}, []Recommendation{/* ... */}},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := checkCompliance(tt.resource)
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Memory Efficiency

### Preallocate Slices
```go
// When size is known, preallocate
recommendations := make([]models.Recommendation, 0, len(resources))
```

### Avoid Unnecessary Allocations
```go
// Reuse buffers when possible
var buf bytes.Buffer
for _, item := range items {
    buf.Reset()
    buf.WriteString(item.String())
    // use buf
}
```

## String Building

### Use strings.Builder
```go
// For efficient string concatenation
var sb strings.Builder
for _, part := range parts {
    sb.WriteString(part)
}
result := sb.String()
```

## Struct Design

### Zero Values
```go
// Make the zero value useful
type Scanner struct {
    // Don't use pointers for simple types
    MaxRetries int  // defaults to 0, which is fine
    Timeout    time.Duration  // defaults to 0, use as "no timeout"
}
```

### Pointer Receivers
```go
// Use pointer receivers for methods that modify the receiver
func (s *Scanner) Configure(config Config) {
    s.config = config
}

// Use value receivers for small types that don't modify
func (r Recommendation) String() string {
    return r.Message
}
```

## Resource Management

### Always Use defer for Cleanup
```go
file, err := os.Open(filename)
if err != nil {
    return err
}
defer file.Close()

// use file
```

### Context Timeouts
```go
// Set reasonable timeouts
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.Get(ctx, id)
```

## Package Organization

### Internal Packages
```go
// Use internal/ for packages that shouldn't be imported externally
// internal/scanners/aks/aks.go - only importable within azqr
```

### Avoid Cyclic Dependencies
```go
// Bad: models imports scanners, scanners imports models
// Good: both import a shared interface package
```

## Documentation

### Godoc Examples
```go
// Document with examples
// Example:
//   scanner := NewScanner(config)
//   results, err := scanner.Scan(ctx)
//   if err != nil {
//       log.Fatal(err)
//   }
func NewScanner(config Config) *Scanner {
    return &Scanner{config: config}
}
```

## Performance Tips

### Avoid Reflection
```go
// Prefer type assertions over reflection
if scanner, ok := s.(CustomScanner); ok {
    scanner.CustomMethod()
}
```

### Use Benchmarks
```go
func BenchmarkScanner_Scan(b *testing.B) {
    scanner := NewScanner(testConfig)
    ctx := context.Background()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanner.Scan(ctx)
    }
}
```

## Common Anti-Patterns to Avoid

❌ **Don't**
```go
// Don't ignore errors
_ = doSomething()

// Don't use panic for normal error handling
if err != nil {
    panic(err)
}

// Don't use global mutable state
var globalCache = make(map[string]interface{})
```

✅ **Do**
```go
// Handle errors properly
if err := doSomething(); err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// Use explicit error returns
if err != nil {
    return err
}

// Pass dependencies explicitly
type Scanner struct {
    cache Cache
}
```
