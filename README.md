# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)

## Overview

Godi is a dependency injection container using Go generics.

## Ecosystem Context

The Go dependency injection landscape includes several approaches:

| Project | Approach | Key Characteristic |
|---------|----------|-------------------|
| **dig/fx** (Uber) | Reflection | Runtime dependency resolution |
| **wire** (Google) | Code Generation | Compile-time dependency resolution |
| **samber/do** | Generics + Reflection | Functional container API |
| **godi** | Generics | Direct type-based injection |

Each approach has different tradeoffs:

- **Reflection-based** (dig): Flexible, no setup required, runtime errors possible
- **Code generation** (wire): Type-safe, requires build step, more complex setup
- **Generics-based** (godi, samber/do): Type-safe, no code generation, runtime resolution

Godi focuses on minimal API surface using Go generics, with automatic circular dependency detection and multi-container support.

## Installation

```bash
go get github.com/Just-maple/godi
```

## Core Concepts

### Container

The `Container` holds registered providers and manages dependency injection.

```go
c := &godi.Container{}
```

### Provider Registration

**Provide** - Register concrete values (variadic):

```go
// Register multiple providers at once
err := c.Add(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
)
if err != nil {
    // Handle duplicate registration error
}

// Method chaining with MustAdd
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
)
```

**Lazy** - Register a factory that executes on first request:

```go
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    // Container is available for injecting dependencies
    cfg, _ := godi.Inject[Config](c)
    return sql.Open("mysql", cfg.DSN)
}))
```

### Dependency Retrieval

```go
// Returns (value, error) - includes lazy factory errors with type information
db, err := godi.Inject[Database](c)
if err != nil {
    // Error includes: "lazy factory error: <original error>"
    // or "provider *Database not found"
    // or "circular dependency for *Database"
}

// Panics if not found
db := godi.MustInject[Database](c)
```

---

## Usage Scenarios

### 1. Basic Injection

Register and retrieve simple dependencies:

```go
package main

import "github.com/Just-maple/godi"

type Config struct {
    DSN string
}

func main() {
    c := &godi.Container{}
    c.Add(
        godi.Provide(Config{DSN: "mysql://localhost"}),
        godi.Provide(Database{DSN: "mysql://localhost"}),
    )
    
    cfg, err := godi.Inject[Config](c)
    if err != nil {
        panic(err)
    }
}
```

### 2. Lazy Loading

Defer expensive initialization until first use. Errors from the factory are returned to the caller:

```go
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    // This code runs only when Database is first requested
    return sql.Open("mysql", dsn)
}))

// Factory executes here, any error is returned
db, err := godi.Inject[*Database](c)
if err != nil {
    // Handle connection error
}
```

### 3. Lazy with Dependencies

Lazy factories can inject their own dependencies:

```go
c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))

c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    // Inject dependency inside factory
    cfg, err := godi.Inject[Config](c)
    if err != nil {
        return nil, err
    }
    return sql.Open("mysql", cfg.DSN)
}))
```

### 4. Dependency Chains

Build chains of dependencies:

```go
// Level 1: Config
c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))

// Level 2: Database depends on Config
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    cfg, _ := godi.Inject[Config](c)
    return NewDatabase(cfg.DSN)
}))

// Level 3: Repository depends on Database
c.Add(godi.Lazy(func(c *godi.Container) (*UserRepository, error) {
    db, _ := godi.Inject[*Database](c)
    return NewUserRepository(db)
}))

// Level 4: Service depends on Repository
c.Add(godi.Lazy(func(c *godi.Container) (*UserService, error) {
    repo, _ := godi.Inject[*UserRepository](c)
    return NewUserService(repo)
}))

// Inject top-level service (triggers entire chain)
svc, err := godi.Inject[*UserService](c)
```

### 5. Circular Dependency Detection

Circular dependencies are detected at runtime:

```go
type A struct{ B *B }
type B struct{ A *A }

c.Add(godi.Lazy(func(c *godi.Container) (A, error) {
    b, err := godi.Inject[B](c)
    return A{B: b}, err
}))

c.Add(godi.Lazy(func(c *godi.Container) (B, error) {
    a, err := godi.Inject[A](c)
    return B{A: a}, err
}))

// Returns error: "circular dependency for main.A"
_, err := godi.Inject[A](c)
```

### 6. Multi-Container Injection

Inject from multiple containers:

```go
dbContainer := &godi.Container{}
cacheContainer := &godi.Container{}

dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost"}))

// Searches both containers
db, _ := godi.Inject[Database](dbContainer, cacheContainer)
cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
```

### 7. InjectTo - Inject into Existing Variable

```go
var db Database
err := godi.InjectTo(&db, c)
if err != nil {
    // Handle error
}
```

### 8. Interface-Based Injection

Register and inject interfaces:

```go
// Define interface
type Database interface {
    Query(string) ([]Row, error)
}

// Register implementation
c.Add(godi.Lazy(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(dsn)
}))

// Inject interface
db, err := godi.Inject[Database](c)
```

### 9. Testing with Mocks

Swap implementations for testing:

```go
// Production
prod := &godi.Container{}
prod.Add(godi.Lazy(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(prodDSN)
}))

// Test
test := &godi.Container{}
test.Add(godi.Provide(&MockDatabase{Data: testdata}))

// Same service works with both
svc := NewUserService(db)
```

### 10. Grouping Related Dependencies

Use structs to group configurations:

```go
type AppConfig struct {
    Database DatabaseConfig
    HTTP     HTTPConfig
    Cache    CacheConfig
}

c.Add(godi.Provide(AppConfig{
    Database: DatabaseConfig{DSN: "mysql://localhost"},
    HTTP:     HTTPConfig{Port: 8080},
    Cache:    CacheConfig{Host: "redis://localhost"},
}))

cfg, _ := godi.Inject[AppConfig](c)
```

### 11. Lifecycle Management and Cleanup

Register cleanup hooks during initialization for graceful shutdown:

```go
// Create lifecycle manager
lifecycle := lifecycle.New("MyApp")
c.Add(godi.Provide(lifecycle))

// Register database with cleanup hook
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    db := NewDatabase(dsn)
    
    // Register cleanup hook
    lifecycle.AddShutdownHook(func(ctx context.Context) error {
        return db.Close()
    })
    
    return db, nil
}))

// Register service with shutdown hook
c.Add(godi.Lazy(func(c *godi.Container) (*Service, error) {
    svc := NewService()
    
    // Register graceful shutdown hook
    lifecycle.AddShutdownHook(func(ctx context.Context) error {
        return svc.Shutdown(ctx)
    })
    
    return svc, nil
}))

// On application exit, shutdown in reverse order
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
lifecycle.Shutdown(shutdownCtx)
```

Hooks execute in reverse order (LIFO - Last In, First Out), ensuring proper cleanup order.

### 12. Chain - Transform Dependencies

Transform one dependency into another using Chain:

```go
// Simple chain: string -> int -> string
type Name string
type Length int
type Result string

c.Add(
    godi.Provide(Name("hello")),
    godi.Chain(func(s Name) (Length, error) {
        return Length(len(s)), nil
    }),
    godi.Chain(func(n Length) (Result, error) {
        return Result(fmt.Sprintf("len%d", n)), nil
    }),
)

result, _ := godi.Inject[Result](c)
// result == "len5"

// Real-world: Config -> Database -> Repository
c.Add(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    godi.Chain(func(cfg Config) (*Database, error) {
        return &Database{ConnString: cfg.DSN}, nil
    }),
    godi.Chain(func(db *Database) (*Repository, error) {
        return &Repository{DB: db}, nil
    }),
)

repo, _ := godi.Inject[*Repository](c)
```

Chain automatically receives the Container for injecting dependencies - no need to pass it explicitly!

---

## API Reference

### Container Methods

| Method | Description |
|--------|-------------|
| `Add(ps ...Provider) error` | Register providers. Returns error if duplicate. |
| `MustAdd(ps ...Provider) *Container` | Register providers. Panics if duplicate. Returns container for chaining. |

### Injection Functions

| Function | Signature | Behavior |
|----------|-----------|----------|
| `Inject[T](c)` | `(T, error)` | Returns error if not found or circular |
| `MustInject[T](c)` | `T` | Panics if not found |
| `InjectTo[T](&v, c)` | `error` | Injects into provided pointer |
| `MustInjectTo[T](&v, c)` | - | Injects into pointer, panics on failure |

### Provider Functions

| Function | Description |
|----------|-------------|
| `Provide[T](v T)` | Register concrete value |
| `Lazy[T](func() (T, error))` | Register factory for deferred execution |

---

## Lazy Loading Patterns

### Pattern 1: Simple Lazy

```go
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    return sql.Open("mysql", dsn)
}))
```

### Pattern 2: Lazy with Error Handling

```go
c.Add(godi.Lazy(func(c *godi.Container) (*Database, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}))
```

### Pattern 3: Lazy with Dependencies

```go
c.Add(godi.Lazy(func(c *godi.Container) (*UserService, error) {
    db, err := godi.Inject[*Database](c)
    if err != nil {
        return nil, err
    }
    cache, err := godi.Inject[*Cache](c)
    if err != nil {
        return nil, err
    }
    return NewUserService(db, cache)
}))
```

### Pattern 4: Lazy Singleton

Lazy providers are singletons - factory executes once:

```go
c.Add(godi.Lazy(func(c *godi.Container) (*ExpensiveResource, error) {
    fmt.Println("Initializing...") // Prints only once
    return NewExpensiveResource()
}))

// Factory executes here
r1, _ := godi.Inject[*ExpensiveResource](c)

// Returns cached value
r2, _ := godi.Inject[*ExpensiveResource](c)
```

### Pattern 5: Conditional Lazy

```go
c.Add(godi.Lazy(func(c *godi.Container) (Database, error) {
    cfg, _ := godi.Inject[Config](c)
    
    if cfg.Environment == "test" {
        return NewMockDatabase(), nil
    }
    return NewMySQLDatabase(cfg.DSN)
}))
```

---

## Thread Safety

All container operations are thread-safe:

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// Safe for concurrent use
go func() {
    db, _ := godi.Inject[Database](c)
}()

go func() {
    db, _ := godi.Inject[Database](c)
}()
```

---

## Supported Types

Godi supports any Go type:

- Structs: `Database`, `Config`
- Primitives: `string`, `int`, `bool`
- Pointers: `*Database`
- Slices: `[]string`
- Maps: `map[string]int`
- Interfaces: `any`, custom interfaces
- Arrays: `[3]int`
- Channels: `chan int`
- Functions: `func() error`

---

## Examples

See [examples/](examples/) for complete examples:

| Example | Description |
|---------|-------------|
| 01-basic | Basic injection |
| 02-error-handling | Error handling |
| 03-must-inject | Panic on error |
| 04-all-types | All supported types |
| 05-multi-container | Multiple containers |
| 06-concurrent | Concurrent access |
| 07-generics | Generic types |
| 08-testing-mock | Mock testing |
| 09-web-app | Production web app |
| 10-lifecycle-cleanup | Lifecycle management and cleanup |
| 11-chain | Dependency transformation with Chain |

## License

MIT License
