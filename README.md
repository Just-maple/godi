# Godi - Simple Dependency Injection Container for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

A lightweight, type-safe dependency injection container for Go applications.

## Features

- **Type-safe**: Leverages Go generics for compile-time type safety
- **Zero Reflection**: No reflection overhead, all type resolution at compile time
- **Simple API**: Minimal, intuitive interface for dependency management
- **Thread-safe**: Built-in mutex protection for concurrent access
- **Zero dependencies**: Pure Go implementation with no external dependencies
- **Flexible**: Supports any type including structs, primitives, slices, and maps
- **Multi-container**: Support injection across multiple containers
- **Circular dependency detection**: Automatic detection of circular dependencies
- **Lazy loading**: Deferred initialization with `Lazy` provider

## Installation

```bash
go get github.com/Just-maple/godi
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Config struct {
    AppName string
}

func main() {
    // Create container
    c := &godi.Container{}
    
    // Register dependencies
    c.Add(godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}))
    c.Add(godi.Provide(Config{AppName: "my-app"}))
    
    // Inject dependencies
    db, ok := godi.Inject[Database](c)
    if !ok {
        panic("failed to inject Database")
    }
    
    cfg, ok := godi.Inject[Config](c)
    if !ok {
        panic("failed to inject Config")
    }
    
    fmt.Printf("Connected to %s for %s\n", db.DSN, cfg.AppName)
}
```

## API Reference

### Container

#### `Add(provider Provider) bool`

Registers a provider in the container. Returns `false` if a provider of the same type already exists.

```go
c := &godi.Container{}
success := c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
```

#### `ShouldAdd(provider Provider) error`

Registers a provider and returns an error if a provider of the same type already exists.

```go
err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
if err != nil {
    // Handle duplicate provider error
}
```

#### `MustAdd(provider Provider)`

Registers a provider and panics if a provider of the same type already exists.

```go
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
```

### Provider Registration

#### `Provide[T any](t T) Provider`

Creates a provider for the given type.

```go
db := Database{DSN: "mysql://localhost"}
provider := godi.Provide(db)
```

#### `Lazy[T any](factory func() (T, error)) Provider`

Creates a lazy-loaded provider. The factory function is called only when the dependency is first requested.

```go
c.Add(godi.Lazy(func() (*Database, error) {
    return ConnectDB("mysql://localhost")
}))
```

### Dependency Injection

#### `Inject[T any](c ...*Container) (v T, ok bool)`

Retrieves a dependency from the container(s). Returns the zero value and `false` if not found. Supports multiple containers.

```go
db, ok := godi.Inject[Database](c)
if !ok {
    // Handle missing dependency
}
```

#### `ShouldInject[T any](c ...*Container) (v T, err error)`

Retrieves a dependency and returns an error if not found. Includes circular dependency detection.

```go
db, err := godi.ShouldInject[Database](c)
if err != nil {
    // Handle error (includes circular dependency detection)
}
```

#### `MustInject[T any](c ...*Container) (v T)`

Retrieves a dependency and panics if not found.

```go
db := godi.MustInject[Database](c)
```

#### `InjectTo[T any](v *T, c ...*Container) error`

Injects a dependency directly into a provided pointer. Returns error if not found or circular dependency detected. Supports multiple containers.

```go
var db Database
err := godi.InjectTo(&db, c)
if err != nil {
    // Handle error
}
```

#### `MustInjectTo[T any](v *T, c ...*Container)`

Injects a dependency directly into a provided pointer and panics if not found.

```go
var db Database
godi.MustInjectTo(&db, c)
```

## Examples

### Lazy Loading

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

func main() {
    c := &godi.Container{}
    
    // Lazy loading - factory called on first use
    c.Add(godi.Lazy(func() (Database, error) {
        fmt.Println("Initializing database connection...")
        return Database{DSN: "mysql://localhost"}, nil
    }))
    
    // Factory is called here
    db, _ := godi.Inject[Database](c)
    fmt.Printf("Connected: %s\n", db.DSN)
    
    // Subsequent calls use cached value
    db2, _ := godi.Inject[Database](c)
    fmt.Printf("Cached: %s\n", db2.DSN)
}
```

### Circular Dependency Detection

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type A struct {
    B *B
}

type B struct {
    A *A
}

func main() {
    c := &godi.Container{}
    
    // Circular dependency: A needs B, B needs A
    c.Add(godi.Lazy(func() (A, error) {
        b, err := godi.ShouldInject[B](c)
        return A{B: b}, err
    }))
    
    c.Add(godi.Lazy(func() (B, error) {
        a, err := godi.ShouldInject[A](c)
        return B{A: a}, err
    }))
    
    // Detects circular dependency
    _, err := godi.ShouldInject[A](c)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        // Output: Error: circular dependency for *main.A
    }
}
```

### Multi-Container Injection

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Cache struct {
    Host string
}

func main() {
    // Create multiple containers
    dbContainer := &godi.Container{}
    cacheContainer := &godi.Container{}
    
    // Register in different containers
    dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
    cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost"}))
    
    // Inject from multiple containers
    db, _ := godi.Inject[Database](dbContainer, cacheContainer)
    cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
    
    fmt.Printf("DB: %s, Cache: %s\n", db.DSN, cache.Host)
}
```

### Error Handling with ShouldInject

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    Port int
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(Config{Port: 8080}))
    
    // Graceful error handling
    config, err := godi.ShouldInject[Config](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Server starting on port %d\n", config.Port)
}
```

### Complex Dependency Graph

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    DSN string
}

type Database struct {
    DSN string
}

type Repository struct {
    DB Database
}

type Service struct {
    Repo Repository
}

func main() {
    c := &godi.Container{}
    
    // Register config
    c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))
    
    // Lazy loading with dependencies
    c.Add(godi.Lazy(func() (Database, error) {
        cfg, err := godi.ShouldInject[Config](c)
        if err != nil {
            return Database{}, err
        }
        return Database{DSN: cfg.DSN}, nil
    }))
    
    c.Add(godi.Lazy(func() (Repository, error) {
        db, err := godi.ShouldInject[Database](c)
        if err != nil {
            return Repository{}, err
        }
        return Repository{DB: db}, nil
    }))
    
    c.Add(godi.Lazy(func() (Service, error) {
        repo, err := godi.ShouldInject[Repository](c)
        if err != nil {
            return Service{}, err
        }
        return Service{Repo: repo}, nil
    }))
    
    // Inject top-level service
    svc, err := godi.ShouldInject[Service](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Service ready: DB=%s\n", svc.Repo.DB.DSN)
}
```

## Supported Types

Godi supports injecting any Go type:

- **Structs**: `Database`, `Config`, custom types
- **Primitives**: `string`, `int`, `bool`, `float64`
- **Pointers**: `*Database`, `*Config`
- **Slices**: `[]string`, `[]int`
- **Maps**: `map[string]int`, `map[string]any`
- **Interfaces**: `any`, custom interfaces
- **Arrays**: `[3]int`, `[5]string`
- **Channels**: `chan int`, `chan string`
- **Functions**: `func()`, `func(int) string`

## Thread Safety

All container operations are protected by a mutex, making it safe for concurrent use:

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// Safe to use from multiple goroutines
go func() {
    db, _ := godi.Inject[Database](c)
    // Use db...
}()

go func() {
    db, _ := godi.Inject[Database](c)
    // Use db...
}()
```

## Best Practices

1. **Register dependencies early**: Set up your container at application startup
2. **Use ShouldInject for error handling**: Prefer `ShouldInject` over `MustInject` when you want to handle errors gracefully
3. **Use Lazy for expensive resources**: Database connections, HTTP clients, etc.
4. **Avoid circular dependencies**: Design your dependency graph carefully
5. **Use multiple containers for modularity**: Separate concerns by using different containers for different modules
6. **Depend on interfaces**: Use interfaces for cross-layer dependencies (see examples/09-web-app)

## Examples Directory

| Example | Description |
|---------|-------------|
| [01-basic](examples/01-basic/) | Basic dependency injection |
| [02-error-handling](examples/02-error-handling/) | Error handling patterns |
| [03-must-inject](examples/03-must-inject/) | Panic-on-error injection |
| [04-all-types](examples/04-all-types/) | All supported types |
| [05-multi-container](examples/05-multi-container/) | Multiple containers |
| [06-concurrent](examples/06-concurrent/) | Concurrent access |
| [07-generics](examples/07-generics/) | Generic type injection |
| [08-testing-mock](examples/08-testing-mock/) | Testing with mocks |
| [09-web-app](examples/09-web-app/) | Production web app (SOLID principles) |

## Comparison

| Feature | Godi | dig/fx | wire |
|---------|------|--------|------|
| **Type Resolution** | Generics | Reflection | Code Gen |
| **Error Detection** | Compile-time | Runtime | Compile-time |
| **Performance** | Zero overhead | Reflection cost | Zero overhead |
| **Setup** | No setup | No setup | Code generation |
| **Multi-container** | ✅ Built-in | ❌ Single | ❌ Single |
| **Circular Detection** | ✅ Runtime | ✅ Runtime | ✅ Compile-time |

## License

MIT License - see LICENSE file for details
