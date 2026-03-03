# GoDI Examples

This directory contains examples demonstrating various features of the godi dependency injection framework.

## Examples Overview

| # | Example | Description | Key Features |
|---|---------|-------------|--------------|
| 01 | [basic](01-basic/) | Basic dependency injection | Provide, Build patterns, struct field injection, dependency chains |
| 02 | [error-handling](02-error-handling/) | Error handling patterns | Add/Inject vs MustAdd/MustInject |
| 03 | [all-types](03-all-types/) | All supported types + generics | Structs, pointers, slices, maps, generics, interfaces |
| 04 | [concurrent](04-concurrent/) | Concurrent access | Thread-safe operations |
| 05 | [testing-mock](05-testing-mock/) | Testing with mocks | Interface-based testing, DI |
| 06 | [web-app](06-web-app/) | Production web app | SOLID principles, DIP, layered architecture |
| 07 | [lifecycle-cleanup](07-lifecycle-cleanup/) | Lifecycle & hooks | Hook, HookOnce, startup/shutdown, graceful cleanup |
| 08 | [nested-container-hooks](08-nested-container-hooks/) | Nested container hooks | Multi-level containers, hook propagation |
| 09 | [runtime-container-add](09-runtime-container-add/) | Dynamic container registration | Environment-based selection, conditional features, interface selection |

## Quick Start

```bash
# Run basic example (Provide, Build, field injection, chains)
cd examples/01-basic
go run main.go

# Run error handling example (Add/Inject vs MustAdd/MustInject)
cd examples/02-error-handling
go run main.go

# Run all types example (including generics)
cd examples/03-all-types
go run main.go

# Run web app example (best practices)
cd examples/06-web-app
go run cmd/main.go

# Run lifecycle & hooks example
cd examples/07-lifecycle-cleanup
go run main.go

# Run nested container hooks example
cd examples/08-nested-container-hooks
go run main.go

# Run runtime container add example (dynamic registration)
cd examples/09-runtime-container-add
go run main.go
```

## Key Concepts

### 1. Provide and Inject

```go
c := &godi.Container{}
c.Add(godi.Provide(MyService{}))

service, err := godi.Inject[MyService](c)
```

### 2. Build Patterns

```go
// Pattern 1: Single dependency (auto-injected)
c.Add(godi.Build(func(cfg Config) (*Database, error) {
    return ConnectDB(cfg.DSN)
}))

// Pattern 2: Container access (multiple dependencies)
c.Add(godi.Build(func(c *godi.Container) (*Service, error) {
    db, _ := godi.Inject[*Database](c)
    cache, _ := godi.Inject[*Cache](c)
    return NewService(db, cache), nil
}))

// Pattern 3: No dependency (struct{})
c.Add(godi.Build(func(_ struct{}) (*Logger, error) {
    return NewLogger(), nil
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

See [06-web-app](06-web-app/) for production-ready example using interfaces:

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

1. **Use interfaces for cross-layer dependencies** (see 06-web-app)
2. **Use Build for expensive resources** (database connections, etc.)
3. **Use Add/Inject for recoverable errors**
4. **Use MustAdd/MustInject for required dependencies**
5. **Keep containers scoped** (one per application layer)

## Documentation

- [Main README](../../README.md)
- [API Documentation](https://pkg.go.dev/github.com/Just-maple/godi)
