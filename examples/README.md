# GoDI Examples

This directory contains examples demonstrating various features of the godi dependency injection framework.

## Examples Overview

| Example | Description | Key Features |
|---------|-------------|--------------|
| [01-basic](01-basic/) | Basic dependency injection | Provide, Inject |
| [02-error-handling](02-error-handling/) | Error handling patterns | ShouldInject, error returns |
| [03-must-inject](03-must-inject/) | Panic-on-error injection | MustInject, panic handling |
| [04-all-types](04-all-types/) | All supported types | Structs, pointers, slices, maps |
| [05-multi-container](05-multi-container/) | Multiple containers | Container isolation |
| [06-concurrent](06-concurrent/) | Concurrent access | Thread-safe operations |
| [07-generics](07-generics/) | Generic type injection | Type parameters |
| [08-testing-mock](08-testing-mock/) | Testing with mocks | Interface-based testing |
| [09-web-app](09-web-app/) | Production web app | SOLID principles, DIP |

## Quick Start

```bash
# Run basic example
cd examples/01-basic
go run main.go

# Run web app example (best practices)
cd examples/09-web-app
go run cmd/main.go
```

## Key Concepts

### 1. Provide and Inject

```go
c := &godi.Container{}
c.Add(godi.Provide(MyService{}))

service, ok := godi.Inject[MyService](c)
```

### 2. Lazy Loading

```go
c.Add(godi.Lazy(func() (*Database, error) {
    return ConnectDB("dsn")
}))
```

### 3. Error Handling

```go
// Returns error instead of bool
service, err := godi.ShouldInject[MyService](c)

// Panics on error
service := godi.MustInject[MyService](c)
```

### 4. Dependency Inversion

See [09-web-app](09-web-app/) for production-ready example using interfaces:

```go
// Register interface, not concrete type
c.Add(godi.Lazy(func() (interfaces.Database, error) {
    return infrastructure.NewDBConnection(dsn), nil
}))

// Depend on abstraction
type Service struct {
    db interfaces.Database  // Interface, not struct
}
```

## Best Practices

1. **Use interfaces for cross-layer dependencies** (see 09-web-app)
2. **Lazy load expensive resources** (database connections, etc.)
3. **Use ShouldInject for recoverable errors**
4. **Use MustInject for required dependencies**
5. **Keep containers scoped** (one per application layer)

## Documentation

- [Main README](../../README.md)
- [API Documentation](https://pkg.go.dev/github.com/Just-maple/godi)
