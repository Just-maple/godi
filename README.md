# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

Lightweight Go dependency injection framework built on generics. Zero reflection, zero code generation.

## 🚀 Features

| Feature | Description |
|---------|-------------|
| **Type-Safe** | Full generics support, compile-time type checking |
| **Lazy Loading** | Dependencies initialized on first use |
| **Circular Detection** | Automatic runtime detection of circular dependencies |

| **Thread-Safe** | All operations are concurrent-safe |
| **Interface Support** | Full dependency inversion principle support |
| **Hook System** | Lifecycle hooks for initialization and cleanup |

## 📦 Installation

```bash
go get github.com/Just-maple/godi
```

## ⚡ Quick Start

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
    println(db.Conn) // Output: mysql://localhost
}
```

## 📖 Core API

### Registration

| Method | Description | Use Case |
|--------|-------------|----------|
| `Provide(T)` | Register instance value | Simple values, configuration |
| `Build(func) (T, error)` | Register factory function (lazy, singleton) | Complex initialization, lazy loading |
| `Chain(func) (T, error)` | Derive from existing dependency | Transform/derive new types |

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

### Injection

| Method | Returns | Panics | Use Case |
|--------|---------|--------|----------|
| `Inject[T](c)` | `(T, error)` | No | Standard injection |
| `MustInject[T](c)` | `T` | Yes | Known to exist |
| `InjectTo(&v, c)` | `error` | No | Inject to existing var |
| `InjectAs(&v, c)` | `error` | No | Non-generic injection |
| `c.Inject(&a, &b)` | `error` | No | Multi-injection |

```go
// Generic injection
db, err := godi.Inject[*Database](c)

// Panic on failure
db := godi.MustInject[*Database](c)

// Inject to existing variable
var db Database
err := godi.InjectTo(&db, c)

// Multi-injection
service := &Service{}
err = c.Inject(&service.DB, &service.Config)
```

### Lifecycle Hooks

```go
c := &godi.Container{}

// Hook with execution counter
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // Skip if already called
    }
    return func(ctx context.Context) {
        fmt.Printf("Starting: %T\n", v)
    }
})

// HookOnce - automatic single execution
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// Execute hooks
ctx := context.Background()
shutdown(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(ctx)
    }
})
```

## 📚 Usage Patterns

### 1. Constructor-Based Injection

```go
type Service struct {
    DB   Database
    Config Config
}

c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
    godi.Build(func(c *godi.Container) (*Service, error) {
        db, _ := godi.Inject[Database](c)
        cfg, _ := godi.Inject[Config](c)
        return &Service{DB: db, Config: cfg}, nil
    }),
)
```

### 2. Field Injection

```go
service := &Service{}
err := c.Inject(&service.DB, &service.Config, &service.Cache)
```

### 3. Interface-Based (Dependency Inversion)

```go
type Database interface {
    Query(string) ([]Row, error)
    Close() error
}

c.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(dsn), nil
}))

db, err := godi.Inject[Database](c)
```

### 4. Dependency Chains

```go
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    godi.Build(func(c *godi.Container) (*Database, error) {
        cfg, _ := godi.Inject[Config](c)
        return NewDatabase(cfg.DSN)
    }),
    godi.Build(func(c *godi.Container) (*Repository, error) {
        db, _ := godi.Inject[*Database](c)
        return NewRepository(db)
    }),
    godi.Build(func(c *godi.Container) (*Service, error) {
        repo, _ := godi.Inject[*Repository](c)
        return NewService(repo)
    }),
)
```

### 5. Graceful Shutdown with Hooks

```go
c := &godi.Container{}

// Register shutdown hook
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// Register resources
c.MustAdd(
    godi.Build(func(c *godi.Container) (*Database, error) {
        return NewDatabase("dsn")
    }),
    godi.Build(func(c *godi.Container) (*Cache, error) {
        return NewCache("redis://localhost")
    }),
)

// Execute shutdown
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

shutdown(func(hooks []func(context.Context)) {
    for i := len(hooks) - 1; i >= 0; i-- {
        hooks[i](shutdownCtx)
    }
})
```

### 6. Testing with Mocks

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



### 8. Transform with Chain

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

## 🔧 Supported Types

- ✅ Structs: `Database`, `Config`
- ✅ Primitives: `string`, `int`, `bool`, `float64`
- ✅ Pointers: `*Database`
- ✅ Slices: `[]string`
- ✅ Maps: `map[string]int`
- ✅ Interfaces: `any`, custom interfaces
- ✅ Arrays: `[3]int`
- ✅ Channels: `chan int`
- ✅ Functions: `func() error`

## 📊 Framework Comparison

| Feature | godi | dig/fx | wire | samber/do |
|---------|------|--------|------|-----------|
| **Type System** | Generics | Reflection | Code Gen | Generics |
| **Runtime Errors** | No | Possible | No | Possible |
| **Build Step** | No | No | Required | No |
| **API Style** | Functional | Functional | Code Gen | Functional |
| **Learning Curve** | Low | Medium | High | Low |
| **Bundle Size** | Minimal | Medium | Large | Small |
| **Lifecycle Hooks** | ✅ | ✅ | ❌ | ✅ |

| **Circular Detection** | ✅ | ✅ | ✅ | ✅ |
| **Lazy Loading** | ✅ | ✅ | ❌ | ✅ |
| **Project Status** | Active | Active | ⚠️ Archived | Active |

### When to Choose godi

- You prefer **compile-time safety** without code generation
- You want **minimal dependencies** and small bundle size
- You need **lifecycle management** for resources
- You value **simple, intuitive API**


## 📁 Examples

Complete examples available in [`examples/`](examples/):

| # | Example | Description |
|---|---------|-------------|
| 01 | [basic](examples/01-basic/) | Basic injection patterns |
| 02 | [error-handling](examples/02-error-handling/) | Error handling strategies |
| 03 | [must-inject](examples/03-must-inject/) | Panic-mode injection |
| 04 | [all-types](examples/04-all-types/) | All supported types |
| 05 | [concurrent](examples/05-concurrent/) | Concurrent safety |
| 06 | [generics](examples/06-generics/) | Advanced generics |
| 07 | [testing-mock](examples/07-testing-mock/) | Mock testing patterns |
| 08 | [web-app](examples/08-web-app/) | Production web app structure |
| 09 | [lifecycle-cleanup](examples/09-lifecycle-cleanup/) | Resource cleanup with hooks |
| 10 | [chain](examples/10-chain/) | Dependency transformation |
| 11 | [struct-field-inject](examples/11-struct-field-inject/) | Struct field injection |
| 12 | [hook](examples/12-hook/) | Hook lifecycle management |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
