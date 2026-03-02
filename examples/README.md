# GoDI Examples

This directory contains examples demonstrating various features of the godi dependency injection framework.

## Examples Overview

| # | Example | Description | Key Features |
|---|---------|-------------|--------------|
| 01 | [basic](01-basic/) | Basic dependency injection | Provide, Inject |
| 02 | [error-handling](02-error-handling/) | Error handling patterns | Add, Inject with error returns |
| 03 | [must-inject](03-must-inject/) | Panic-on-error injection | MustAdd, MustInject, MustInjectTo |
| 04 | [all-types](04-all-types/) | All supported types | Structs, pointers, slices, maps, functions |
| 05 | [concurrent](05-concurrent/) | Concurrent access | Thread-safe operations |
| 06 | [generics](06-generics/) | Generic type injection | Type parameters, constructor injection |
| 07 | [testing-mock](07-testing-mock/) | Testing with mocks | Interface-based testing, DI |
| 08 | [web-app](08-web-app/) | Production web app | SOLID principles, DIP, layered architecture |
| 09 | [lifecycle-cleanup](09-lifecycle-cleanup/) | Resource cleanup | Lifecycle hooks, graceful shutdown |
| 10 | [chain](10-chain/) | Dependency transformation | Chain, Build, constructor injection |
| 11 | [struct-field-inject](11-struct-field-inject/) | Struct field injection | Inject multiple fields |
| 12 | [hook](12-hook/) | Hook lifecycle management | Hook, HookOnce, startup/shutdown |
| 13 | [nested-container-hooks](13-nested-container-hooks/) | Nested container hooks | Multi-level containers, hook propagation |

## Quick Start

```bash
# Run basic example
cd examples/01-basic
go run main.go

# Run web app example (best practices)
cd examples/08-web-app
go run cmd/main.go

# Run nested container hooks example
cd examples/13-nested-container-hooks
go run main.go
```

## Key Concepts

### 1. Provide and Inject

```go
c := &godi.Container{}
c.Add(godi.Provide(MyService{}))

service, err := godi.Inject[MyService](c)
```

### 2. Build

```go
c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    return ConnectDB("dsn")
}))
```

### 3. Error Handling

```go
// Returns error
service, err := godi.Inject[MyService](c)

// Panics on error
service := godi.MustInject[MyService](c)
```

### 4. Dependency Inversion

See [08-web-app](08-web-app/) for production-ready example using interfaces:

```go
// Register interface, not concrete type
c.Add(godi.Build(func(c *godi.Container) (interfaces.Database, error) {
    return infrastructure.NewDBConnection(dsn), nil
}))

// Depend on abstraction
type Service struct {
    db interfaces.Database  // Interface, not struct
}
```

## Best Practices

1. **Use interfaces for cross-layer dependencies** (see 08-web-app)
2. **Use Build for expensive resources** (database connections, etc.)
3. **Use Add/Inject for recoverable errors**
4. **Use MustAdd/MustInject for required dependencies**
5. **Keep containers scoped** (one per application layer)

## Documentation

- [Main README](../../README.md)
- [API Documentation](https://pkg.go.dev/github.com/Just-maple/godi)
