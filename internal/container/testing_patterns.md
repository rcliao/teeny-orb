# Testing Patterns for teeny-orb

This document outlines the testing patterns and conventions used in the teeny-orb project.

## Test Structure

### File Organization
- Unit tests: `*_test.go` files alongside the code they test
- Test utilities: `testutils.go` for shared test helpers
- Integration tests: Use build tags `//go:build integration`

### Naming Conventions
- Test functions: `Test<FunctionName>` or `Test<Type>_<Method>`
- Benchmark functions: `Benchmark<FunctionName>`
- Test helper functions: Start with lowercase, use descriptive names

## Test Categories

### Unit Tests
- Test individual functions and methods in isolation
- Use mocks for external dependencies
- Fast execution (< 1 second per test)
- No external services (Docker, databases, etc.)

### Integration Tests
- Test interactions between components
- May use real external services
- Marked with `//go:build integration`
- Run with `go test -tags=integration`

### Benchmarks
- Performance testing for critical paths
- Use `testing.B` parameter
- Focus on ID generation, session management, command execution

## Mocking Strategy

### Interface-Based Mocking
- All external dependencies have interfaces
- Mock implementations in `testutils.go`
- Use dependency injection for testability

### Example Mock Structure
```go
// MockSession implements Session interface
type MockSession struct {
    id       string
    status   SessionStatus
    commands [][]string  // Track executed commands
    closed   bool
}

// Helper methods for test assertions
func (s *MockSession) GetExecutedCommands() [][]string
func (s *MockSession) SetStatus(status SessionStatus)
```

## Test Utilities

### ID Generation
- `StaticIDGenerator` for predictable IDs in tests
- `DefaultIDGenerator` for production use

### Session Management
- `MockManager` for testing session lifecycle
- `MockSession` for testing session operations

### Configuration
- Standard test configurations for different scenarios
- Helper functions to create valid/invalid configs

## Testing Docker Integration

### Unit Tests
- Test logic without Docker client
- Mock Docker client interface
- Focus on error handling and state management

### Integration Tests
- Require Docker daemon running
- Test actual container creation and management
- Use temporary containers that auto-cleanup

## Error Testing

### Pattern
```go
func TestFunction_Error(t *testing.T) {
    // Setup error condition
    // Call function
    // Verify error is returned
    // Verify error message is appropriate
}
```

### Common Error Scenarios
- Invalid configurations
- Missing dependencies
- Resource exhaustion
- Network failures
- Permission issues

## Test Data Management

### Temporary Resources
- Use `os.MkdirTemp()` for temporary directories
- Always cleanup with `defer os.RemoveAll(tempDir)`
- Use unique names to avoid conflicts

### Test Isolation
- Each test should be independent
- Clean up all resources
- Reset global state when necessary

## Performance Testing

### Benchmarks
- Test critical performance paths
- Focus on ID generation, session creation, command execution
- Use `b.ResetTimer()` to exclude setup time

### Example
```go
func BenchmarkIDGeneration(b *testing.B) {
    gen := &DefaultIDGenerator{}
    b.ResetTimer()
    
    for i := 0; i < b.N; i++ {
        _ = gen.GenerateID()
    }
}
```

## Test Commands

### Running Tests
```bash
# All tests
make test

# Unit tests only
make test-short

# With coverage
make test-coverage

# With race detection
make test-race

# Integration tests
make test-integration

# Benchmarks
make bench
```

### Test Flags
- `-v`: Verbose output
- `-short`: Skip slow tests
- `-race`: Enable race detector
- `-timeout`: Set test timeout
- `-count`: Run tests multiple times

## Quality Assurance

### Pre-commit Checks
```bash
make check  # fmt + vet + test-short
```

### CI Pipeline
```bash
make ci     # fmt + vet + test-coverage + lint
```

## Best Practices

1. **Test Names**: Use descriptive names that explain what is being tested
2. **Test Structure**: Arrange-Act-Assert pattern
3. **Error Messages**: Include context in error messages
4. **Test Data**: Use table-driven tests for multiple scenarios
5. **Mocking**: Mock at the boundary, not internal functions
6. **Cleanup**: Always clean up resources
7. **Isolation**: Tests should not depend on each other
8. **Coverage**: Aim for >80% code coverage
9. **Documentation**: Comment complex test scenarios
10. **Refactoring**: Keep tests maintainable and readable
