# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)

Lightweight Go dependency injection framework with generics. Zero reflection, zero code generation.

## Features

- ✅ **Generics** - Type-safe, no reflection
- ✅ **Lazy Loading** - Dependencies initialized on first use
- ✅ **Circular Dependency Detection** - Automatic runtime detection
- ✅ **Multi-Container Injection** - Cross-container lookup support
- ✅ **Thread-Safe** - All operations are concurrent-safe
- ✅ **Interface Injection** - Supports Dependency Inversion Principle
- ✅ **Hook System** - Lifecycle hooks for initialization and cleanup

## Quick Start

```go
package main

import "github.com/Just-maple/godi"

type Config struct {
    DSN string
}

type Database struct {
    Conn string
}

func main() {
    c := &godi.Container{}
    
    // Register dependencies
    c.MustAdd(
        godi.Provide(Config{DSN: "mysql://localhost"}),
        godi.Build(func(c *godi.Container) (*Database, error) {
            cfg, _ := godi.Inject[Config](c)
            return &Database{Conn: cfg.DSN}, nil
        }),
    )
    
    // Inject dependencies
    db := godi.MustInject[*Database](c)
    println(db.Conn)
}
```

## Core API

### Register Dependencies

```go
c := &godi.Container{}

// Provide - Register instance value
c.Add(godi.Provide(Config{Port: 8080}))

// Build - Register factory function (lazy, singleton)
c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    return NewDatabase("dsn")
}))

// Chain - Derive new dependency from existing one
c.Add(godi.Chain(func(cfg Config) (*Connection, error) {
    return NewConnection(cfg.DSN)
}))
```

### Inject Dependencies

```go
// Inject - Returns typed value and error
db, err := godi.Inject[*Database](c)

// MustInject - Panics on failure
db := godi.MustInject[*Database](c)

// InjectTo - Inject into existing variable (generic)
var db Database
err := godi.InjectTo(&db, c)

// MustInjectTo - Inject into existing variable, panics on failure
godi.MustInjectTo(&db, c)

// InjectAs - Inject into existing variable (non-generic)
db = Database{}
err = godi.InjectAs(&db, c)

// MustInjectAs - Non-generic version with panic
godi.MustInjectAs(&db, c)

// Container.Inject - Inject multiple dependencies at once
service := &Service{}
err = c.Inject(&service.DB, &service.Config, &service.Cache)
```

### Multi-Container Injection

```go
db, err := godi.Inject[*Database](container1, container2, container3)
```

Searches containers in order, returns first match.

## Usage Examples

### 1. Struct Field Injection

```go
type Service struct {
    DB     Database
    Config Config
    Cache  Cache
}

c := &godi.Container{}
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// Inject directly into struct fields
service := &Service{}
err := c.Inject(&service.DB, &service.Config, &service.Cache)
```

### 2. Basic Injection

```go
c := &godi.Container{}
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    godi.Provide(Database{DSN: "mysql://localhost"}),
)

cfg, err := godi.Inject[Config](c)
```

### 3. Lazy Loading

Factory executes only on first request, result is cached:

```go
c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    // Executes on first call
    return sql.Open("mysql", dsn)
}))

// Factory executes here
db, err := godi.Inject[*Database](c)
```

### 4. Dependency Chains

```go
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    
    godi.Build(func(c *godi.Container) (*Database, error) {
        cfg, _ := godi.Inject[Config](c)
        return NewDatabase(cfg.DSN)
    }),
    
    godi.Build(func(c *godi.Container) (*UserRepository, error) {
        db, _ := godi.Inject[*Database](c)
        return NewUserRepository(db)
    }),
    
    godi.Build(func(c *godi.Container) (*UserService, error) {
        repo, _ := godi.Inject[*UserRepository](c)
        return NewUserService(repo)
    }),
)

svc := godi.MustInject[*UserService](c)
```

### 5. Circular Dependency Detection

```go
type A struct{ B *B }
type B struct{ A *A }

c.MustAdd(
    godi.Build(func(c *godi.Container) (A, error) {
        b, _ := godi.Inject[B](c)
        return A{B: b}, nil
    }),
    godi.Build(func(c *godi.Container) (B, error) {
        a, _ := godi.Inject[A](c)
        return B{A: a}, nil
    }),
)

// Returns error: "circular dependency for main.A"
_, err := godi.Inject[A](c)
```

### 6. Interface Injection

```go
type Database interface {
    Query(string) ([]Row, error)
}

c.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(dsn)
}))

db, err := godi.Inject[Database](c)
```

### 7. Testing with Mocks

```go
// Production
prod := &godi.Container{}
prod.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(prodDSN)
}))

// Test
test := &godi.Container{}
test.Add(godi.Provide(&MockDatabase{Data: testData}))

// Same service code, different implementations
svc := NewUserService(db)
```

### 8. Chain - Transform Dependencies

```go
type Name string
type Length int

c.MustAdd(
    godi.Provide(Name("hello")),
    godi.Chain(func(s Name) (Length, error) {
        return Length(len(s)), nil
    }),
)

len := godi.MustInject[Length](c) // 5
```

### 9. Hook - Lifecycle Management

```go
c := godi.Container{}
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
)

// Hook with provided counter - only execute on first provision
startup := c.Hook("startup", func(v any, provided int) godi.HookFunc {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) {
        fmt.Printf("Starting: %T\n", v)
    }
})

// Or use HookOnce for automatic single execution
shutdown := c.HookOnce("shutdown", func(v any, provided int) godi.HookFunc {
    return func(ctx context.Context) {
        fmt.Printf("Stopping: %T\n", v)
    }
})

db, _ := godi.Inject[Database](c)
cfg, _ := godi.Inject[Config](c)

ctx := context.Background()

startup(func(hooks []godi.HookFunc) {
    for _, fn := range hooks {
        fn(ctx)
    }
})

shutdown(func(hooks []godi.HookFunc) {
    for _, fn := range hooks {
        fn(ctx)
    }
})
```

## Supported Types

- Structs: `Database`, `Config`
- Primitives: `string`, `int`, `bool`, `float64`
- Pointers: `*Database`
- Slices: `[]string`
- Maps: `map[string]int`
- Interfaces: `any`, custom interfaces
- Arrays: `[3]int`
- Channels: `chan int`
- Functions: `func() error`

## Thread Safety

All container operations are thread-safe:

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// Safe for concurrent injection
go func() {
    db, _ := godi.Inject[Database](c)
}()
```

## Examples

See [examples/](examples/) for complete examples:

| Example | Description |
|---------|-------------|
| [01-basic](examples/01-basic/) | Basic injection |
| [02-error-handling](examples/02-error-handling/) | Error handling |
| [03-must-inject](examples/03-must-inject/) | Panic mode |
| [04-all-types](examples/04-all-types/) | All supported types |
| [05-multi-container](examples/05-multi-container/) | Multi-container injection |
| [06-concurrent](examples/06-concurrent/) | Concurrent safety |
| [07-generics](examples/07-generics/) | Generic injection |
| [08-testing-mock](examples/08-testing-mock/) | Mock testing |
| [09-web-app](examples/09-web-app/) | Web app best practices |
| [10-lifecycle-cleanup](examples/10-lifecycle-cleanup/) | Lifecycle management |
| [11-chain](examples/11-chain/) | Dependency transformation |
| [12-struct-field-inject](examples/12-struct-field-inject/) | Struct field injection |
| [13-hook](examples/13-hook/) | Hook lifecycle management |

## Comparison with Other Frameworks

| Framework | Approach | Characteristics |
|-----------|----------|-----------------|
| **dig/fx** (Uber) | Reflection | Runtime resolution, possible runtime errors |
| **wire** (Google) | Code Generation | Compile-time resolution, requires build step |
| **samber/do** | Generics + Reflection | Functional API |
| **godi** | Pure Generics | Type-safe, no code generation, minimal API |

## License

MIT License
