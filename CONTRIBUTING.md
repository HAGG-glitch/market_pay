# Contributing to MarketPay USSD Service

Thank you for your interest in contributing to the MarketPay USSD Service! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites
- Go 1.21 or higher
- Git
- Make (optional, but recommended)

### Setup Development Environment

```bash
# Clone the repository
git clone <repository-url>
cd marketpay

# Download dependencies
go mod download

# Build the project
go build -o bin/marketpay-ussd main.go examples.go
```

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write clean, readable code
- Follow Go conventions and style guide
- Add comments for complex logic
- Use meaningful variable and function names

### 3. Testing

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...
```

### 4. Code Formatting

```bash
# Format all code
go fmt ./...

# Run linter (if installed)
golangci-lint run ./...

# Tidy dependencies
go mod tidy
```

### 5. Commit Changes

```bash
git add .
git commit -m "Brief description of changes"
```

### 6. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Then create a pull request on GitHub.

## Code Style Guidelines

### General Rules
- Use meaningful variable names
- Keep functions small and focused
- Add documentation comments for exported functions
- Write error messages that are user-friendly

### Example Function Documentation
```go
// PaymentValidator validates payment inputs according to MarketPay business rules.
// It returns true if the payment amount is within acceptable limits.
func PaymentValidator(amount int64) bool {
    return amount >= 1 && amount <= 10000000
}
```

### Error Handling
```go
// Good: Clear error messaging
if err != nil {
    log.Error().Err(err).Str("session_id", sessionID).Msg("failed to load state")
    return nil, err
}

// Bad: Silent failures
if err != nil {
    return nil, nil
}
```

### Logging
```go
// Good: Structured logging with context
log.Debug().
    Str("session_id", sessionID).
    Str("vendor_code", maskVendorCode(code)).
    Msg("validating vendor code")

// Bad: Unstructured logging
fmt.Println("Validating vendor code:", code)
```

## Adding New Features

### Adding a New USSD Flow Page

1. **Define the page constant** in `internal/ussd/types.go`:
```go
const (
    PageNewFeature FlowPage = "mp_new_feature"
)
```

2. **Update page sequence** in `internal/ussd/types.go`:
```go
var pageSequence = map[FlowPage]FlowPage{
    // ...
    PageNewFeature: PageNextFeature,
}
```

3. **Implement the handler** in `internal/ussd/service.go`:
```go
func (s *MarketPayFlowService) handleNewFeature(ctx context.Context, sessionID string, data map[string]string) (*FlowResult, error) {
    log.Debug().Str("session_id", sessionID).Msg("processing PageNewFeature")
    
    // Implementation logic
    
    return s.navigateToNext(ctx, sessionID, data, PageNextFeature, "Next step message")
}
```

4. **Add the case** in the `Advance` method:
```go
case PageNewFeature:
    return s.handleNewFeature(ctx, input.SessionID, data)
```

### Adding Input Validation

1. **Add validator function** in `internal/ussd/validators.go`:
```go
func ValidateNewInput(input string) bool {
    // Validation logic
    return true
}
```

2. **Add validation message** in `GetValidationMessage`:
```go
messages := map[string]string{
    "new_input": "Error message for new input",
}
```

### Adding State Store Implementation

1. **Implement the StateStore interface** in new file `internal/store/postgres.go`:
```go
type PostgresStateStore struct {
    db *sql.DB
}

func (s *PostgresStateStore) Load(ctx context.Context, sessionID string) (map[string]string, error) {
    // Implementation
}

func (s *PostgresStateStore) Save(ctx context.Context, sessionID string, data map[string]string) error {
    // Implementation
}
```

## Testing Guidelines

### Unit Tests
- Test each handler function
- Test validation logic
- Test error cases
- Aim for > 80% code coverage

### Integration Tests
- Test complete flows end-to-end
- Test state persistence
- Test error scenarios

### Example Test
```go
func TestHandleCollectVendorName(t *testing.T) {
    stateStore := store.NewInMemoryStateStore()
    flowService := ussd.NewMarketPayFlowService(stateStore)
    ctx := context.Background()
    
    result, err := flowService.Advance(ctx, ussd.AdvanceFlowInput{
        SessionID:   "test-001",
        CurrentPage: ussd.PageCollectVendorName,
        Values:      map[string]string{"registration_vendor_name": "John's Store"},
    })
    
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    if result.NextPage != ussd.PageCollectMarketName {
        t.Errorf("expected next page %q, got %q", ussd.PageCollectMarketName, result.NextPage)
    }
}
```

## Documentation

- Update README.md for significant changes
- Update API.md for endpoint changes
- Add comments to complex logic
- Document configuration options

## Submitting a Pull Request

### PR Title Format
```
[TYPE] Brief description

Types: feat, fix, docs, refactor, perf, test, chore
```

### PR Description Template
```markdown
## Description
Brief description of what this PR does.

## Changes
- Change 1
- Change 2
- Change 3

## Testing
Describe how you tested these changes.

## Related Issues
Closes #123

## Checklist
- [ ] Code follows style guidelines
- [ ] Tests pass
- [ ] Documentation updated
- [ ] No breaking changes
```

## Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

### Examples
```
feat: add new loan feature

- Add loan eligibility check
- Add loan application flow
- Validate loan amounts

Closes #456
```

```
fix: correct vendor code validation

The vendor code pattern was too restrictive and rejected valid codes
with leading zeros. Updated pattern to properly handle MP00001.

Fixes #789
```

## Code Review Process

1. Ensure all tests pass
2. Code review by maintainers
3. Address feedback and comments
4. Re-review if significant changes made
5. Merge when approved

## Issue Reporting

When reporting issues, please include:
- Description of the problem
- Steps to reproduce
- Expected behavior
- Actual behavior
- Go version
- System information (OS, etc.)

## Questions?

- Check existing documentation
- Look at existing code examples
- Open a discussion in GitHub Issues
- Contact the development team

## Code of Conduct

Please be respectful and professional in all interactions.

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.
