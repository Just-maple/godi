# Godi - Simple Dependency Injection Container for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

A lightweight, type-safe dependency injection container for Go applications.

## Features

- **Type-safe**: Leverages Go generics for compile-time type safety
- **Simple API**: Minimal, intuitive interface for dependency management
- **Thread-safe**: Built-in mutex protection for concurrent access
- **Zero dependencies**: Pure Go implementation with no external dependencies
- **Flexible**: Supports any type including structs, primitives, slices, and maps
- **Multi-container**: Support injection across multiple containers

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
    db, _ := godi.Inject[Database](c)
    cfg, _ := godi.Inject[Config](c)
    
    fmt.Printf("Connected to %s for %s\n", db.DSN, cfg.AppName)
}
```

## API Reference

### Container

The `Container` type holds all registered providers and manages dependency injection.

#### `Add(provider Provider) bool`

Registers a provider in the container. Returns `false` if a provider of the same type already exists.

```go
c := &godi.Container{}
success := c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
```

#### `ShouldAdd(provider Provider) error`

Registers a provider and returns an error if a provider of the same type already exists.

```go
c := &godi.Container{}
err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
if err != nil {
    // Handle duplicate provider error
}
```

#### `MustAdd(provider Provider)`

Registers a provider and panics if a provider of the same type already exists.

```go
c := &godi.Container{}
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
```

### Provider Registration

#### `Provide[T any](t T) Provider`

Creates a provider for the given type.

```go
db := Database{DSN: "mysql://localhost"}
provider := godi.Provide(db)
```

### Dependency Injection

#### `Inject[T any](containers ...*Container) (v T, ok bool)`

Retrieves a dependency from the container(s). Returns the zero value and `false` if not found. Supports multiple containers.

```go
db, ok := godi.Inject[Database](c)
if !ok {
    // Handle missing dependency
}
```

#### `ShouldInject[T any](containers ...*Container) (v T, err error)`

Retrieves a dependency and returns an error if not found.

```go
db, err := godi.ShouldInject[Database](c)
if err != nil {
    // Handle error
}
```

#### `MustInject[T any](containers ...*Container) (v T)`

Retrieves a dependency and panics if not found.

```go
db := godi.MustInject[Database](c)
```

#### `InjectTo[T any](v *T, containers ...*Container) (ok bool)`

Injects a dependency directly into a provided pointer. Supports multiple containers.

```go
var db Database
ok := godi.InjectTo(&db, c)
```

#### `ShouldInjectTo[T any](v *T, containers ...*Container) error`

Injects a dependency and returns an error if not found.

```go
var db Database
err := godi.ShouldInjectTo(&db, c)
```

#### `MustInjectTo[T any](v *T, containers ...*Container)`

Injects a dependency and panics if not found.

```go
var db Database
godi.MustInjectTo(&db, c)
```

## Examples

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Logger struct {
    Level string
}

type Service struct {
    DB     Database
    Logger Logger
}

func main() {
    c := &godi.Container{}
    
    // Register dependencies
    c.Add(godi.Provide(Database{DSN: "postgres://localhost/mydb"}))
    c.Add(godi.Provide(Logger{Level: "info"}))
    
    // Inject into struct
    db, _ := godi.Inject[Database](c)
    logger, _ := godi.Inject[Logger](c)
    
    service := Service{
        DB:     db,
        Logger: logger,
    }
    
    fmt.Printf("Service ready: %v\n", service)
}
```

### Using ShouldInject for Error Handling

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
    
    // Use ShouldInject for proper error handling
    config, err := godi.ShouldInject[Config](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Server starting on port %d\n", config.Port)
}
```

### Working with Different Types

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

func main() {
    c := &godi.Container{}
    
    // Register various types
    c.Add(godi.Provide("application-name"))
    c.Add(godi.Provide(42))
    c.Add(godi.Provide(3.14))
    c.Add(godi.Provide(true))
    c.Add(godi.Provide([]string{"a", "b", "c"}))
    c.Add(godi.Provide(map[string]int{"x": 1}))
    
    str, _ := godi.Inject[string](c)
    num, _ := godi.Inject[int](c)
    slice, _ := godi.Inject[[]string](c)
    
    fmt.Printf("String: %s, Number: %d, Slice: %v\n", str, num, slice)
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
    // Inject will search through all provided containers
    db, _ := godi.Inject[Database](dbContainer, cacheContainer)
    cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
    
    fmt.Printf("DB: %s, Cache: %s\n", db.DSN, cache.Host)
}
```

### Using InjectTo

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    AppName string
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(Config{AppName: "my-app"}))
    
    // Use InjectTo to inject directly into a variable
    var cfg Config
    ok := godi.InjectTo(&cfg, c)
    
    fmt.Printf("Injected: %v, App: %s\n", ok, cfg.AppName)
}
```

### Using MustInject for Critical Dependencies

```go
package main

import (
    "github.com/Just-maple/godi"
)

type CriticalConfig struct {
    SecretKey string
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(CriticalConfig{SecretKey: "my-secret"}))
    
    // Use MustInject when dependency is critical
    config := godi.MustInject[CriticalConfig](c)
    
    // Continue with guaranteed config
    _ = config
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
3. **Group related dependencies**: Consider using structs to group related configurations
4. **Avoid circular dependencies**: Design your dependency graph carefully
5. **Use multiple containers for modularity**: Separate concerns by using different containers for different modules

## License

MIT License - see LICENSE file for details
