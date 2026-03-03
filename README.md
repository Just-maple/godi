# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)
[![Test](https://github.com/Just-maple/godi/actions/workflows/test.yml/badge.svg)](https://github.com/Just-maple/godi/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Just-maple/godi/branch/master/graph/badge.svg)](https://codecov.io/gh/Just-maple/godi)

Lightweight Go dependency injection framework built on generics. Zero reflection, zero code generation.

## 🚀 Features

| Feature | Description |
|---------|-------------|
| **Type-Safe** | Full generics support, compile-time type checking |
| **Lazy Loading** | Dependencies initialized on first use (singleton) |
| **Circular Detection** | Automatic runtime detection of circular dependencies |
| **Thread-Safe** | All operations are concurrent-safe (sync.Map) |
| **Interface Support** | Full dependency inversion principle support |
| **Hook System** | Lifecycle hooks with explicit execution |
| **Container Nesting** | Tree-structured containers with freeze protection |
| **Runtime Add** | Dynamic container registration in Build functions |

## 📦 Installation

```bash
go get github.com/Just-maple/godi
```

## ⚡ Quick Start

```go
package main

import "github.com/Just-maple/godi"

type Config struct{ DSN string }
type Database struct{ Conn string }

func main() {
    c := &godi.Container{}
    
    c.MustAdd(
        godi.Provide(Config{DSN: "mysql://localhost"}),
        godi.Build(func(c *godi.Container) (*Database, error) {
            cfg, _ := godi.Inject[Config](c)
            return &Database{Conn: cfg.DSN}, nil
        }),
    )
    
    db := godi.MustInject[*Database](c)
    println(db.Conn) // Output: mysql://localhost
}
```

## 📖 Core API

### Registration

| Method | Description | Use Case |
|--------|-------------|----------|
| `Provide(T)` | Register instance value | Simple values, configuration |
| `Build(func) (T, error)` | Register factory function (lazy singleton) | Complex initialization |

```go
c := &godi.Container{}

// Provide - Register instance value
c.Add(godi.Provide(Config{Port: 8080}))

// Build - Register factory function (lazy singleton)
// Pattern 1: Single dependency (auto-injected)
c.Add(godi.Build(func(cfg Config) (*Database, error) {
    return NewDatabase(cfg.DSN)
}))

// Pattern 2: Container access (multiple dependencies)
c.Add(godi.Build(func(c *godi.Container) (*Service, error) {
    db, _ := godi.Inject[*Database](c)
    cache, _ := godi.Inject[*Cache](c)
    return NewService(db, cache), nil
}))

// Pattern 3: No dependency (using struct{})
c.Add(godi.Build(func(_ struct{}) (*Logger, error) {
    return NewLogger(), nil
}))
```

### Injection

| Method | Returns | Panics | Use Case |
|--------|---------|--------|----------|
| `Inject[T](c)` | `(T, error)` | No | Standard injection |
| `MustInject[T](c)` | `T` | Yes | Known to exist |
| `InjectTo(c, &v)` | `error` | No | Inject to existing var |
| `InjectAs(c, &v)` | `error` | No | Non-generic injection |
| `c.Inject(&a, &b)` | `error` | No | Multi-injection |

```go
// Generic injection
db, err := godi.Inject[*Database](c)

// Panic on failure
db := godi.MustInject[*Database](c)

// Inject to existing variable
var db Database
err := godi.InjectTo(c, &db)

// Multi-injection
err = c.Inject(&service.DB, &service.Config)
```

### Lifecycle Hooks

Hooks allow registering callbacks that execute when dependencies are injected. **Hooks are explicitly executed** - you must call the returned executor function.

```go
package main

import (
    "context"
    "fmt"
    "github.com/Just-maple/godi"
)

c := &godi.Container{}

// Register hook BEFORE injecting dependencies
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// Add and inject dependencies (hooks are registered)
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
_, _ = godi.Inject[Database](c)

// Execute hooks explicitly
shutdown.Iterate(context.Background(), false)
```

**Hook Mechanisms:**

| Aspect | Behavior |
|--------|----------|
| **Trigger Point** | Hooks register when dependency is injected |
| **Execution** | Explicit - must call executor function |
| **`provided` Counter** | Tracks injection count (0 = first time) |
| **HookOnce** | Automatically skips when `provided > 0` |
| **Hook** | Manual control via `provided` parameter |
| **Nested Containers** | Hooks trigger on each container in path |

**Hook in Nested Containers:**

```go
// Infrastructure layer
infra := &godi.Container{}
infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
infraHook := infra.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) { fmt.Printf("[Infra] %T\n", v) }
})

// Application layer
app := &godi.Container{}
app.MustAdd(infra)
appHook := app.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) { fmt.Printf("[App] %T\n", v) }
})

// Inject from parent - triggers hooks on BOTH containers
_, _ = godi.Inject[Database](app)

// Execute hooks for each container separately
ctx := context.Background()
infraHook.Iterate(ctx, false)
appHook.Iterate(ctx, false)

// Output:
// [Infra] Database
// [App] Database
```

### Container Nesting

Containers can be nested to create modular applications. Child containers become **frozen** after being added to parent.

```go
// Child container
child := &godi.Container{}
child.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

// Parent container with nested child
parent := &godi.Container{}
parent.MustAdd(child)  // child is now frozen

// Inject from parent (finds Database in child)
db, _ := godi.Inject[Database](parent)

// Duplicate prevention
err := parent.Add(godi.Provide(Database{DSN: "other"}))
// err: provider *godi.Database already exists
```

**Key Mechanisms:**

| Mechanism | Behavior |
|-----------|----------|
| **Tree Search** | Inject traverses child containers depth-first |
| **Duplicate Detection** | Add checks all nested containers |
| **Container Freezing** | Child containers frozen after parent addition |
| **Hook Propagation** | Hooks trigger on each container in path |
| **Runtime Add** | Build functions CAN add containers dynamically |

**Container Freezing:**

```go
child := &godi.Container{}
child.MustAdd(godi.Provide(Config{Value: "child"}))

parent := &godi.Container{}
parent.MustAdd(child)  // child is now frozen

// This fails: "container is frozen"
err := child.Add(godi.Provide(struct{ Name string }{Name: "new"}))
```

**Runtime Add in Build (Allowed):**

```go
c := &godi.Container{}
nested := &godi.Container{}
nested.MustAdd(godi.Provide("value"))

c.MustAdd(godi.Build(func(c *godi.Container) (string, error) {
    // This is OK - adding during Build execution
    c.MustAdd(nested)
    return godi.Inject[string](c)
}))
```

## 📚 Usage Patterns

### 1. Constructor-Based Injection

```go
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

### 2. Interface-Based (Dependency Inversion)

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

### 3. Environment-Based Selection (Runtime Add)

```go
prodDB := &godi.Container{}
prodDB.MustAdd(godi.Provide(Database{DSN: "mysql://prod-db"}))

devDB := &godi.Container{}
devDB.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

c.MustAdd(
    godi.Provide(Config{Env: "production"}),
    godi.Build(func(c *godi.Container) (Database, error) {
        cfg, _ := godi.Inject[Config](c)
        if cfg.Env == "production" {
            c.MustAdd(prodDB)
        } else {
            c.MustAdd(devDB)
        }
        return godi.Inject[Database](c)
    }),
)
```

### 4. Graceful Shutdown with Hooks

```go
c := &godi.Container{}

// Register hook BEFORE injection
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

c.MustAdd(
    godi.Build(func(c *godi.Container) (*Database, error) {
        return NewDatabase("dsn")
    }),
)

// Inject (hook is registered)
_, _ = godi.Inject[*Database](c)

// Execute with timeout
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
shutdown.Iterate(ctx, true)  // true = reverse order
```

### 5. Testing with Mocks

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
| **Container Nesting** | ✅ | ❌ | ❌ | ❌ |
| **Runtime Add** | ✅ | ❌ | ❌ | ❌ |
| **Learning Curve** | Low | Medium | High | Low |

### When to Choose godi

- You prefer **compile-time safety** without code generation
- You want **minimal dependencies** and small bundle size
- You need **lifecycle management** for resources
- You value **simple, intuitive API**
- You need **container nesting** for modular architecture
- You need **runtime container registration** for dynamic scenarios

## 📁 Examples

Complete examples available in [`examples/`](examples/):

| # | Example | Description |
|---|---------|-------------|
| 01 | [basic](examples/01-basic/) | Basic injection patterns |
| 02 | [error-handling](examples/02-error-handling/) | Error handling strategies |
| 03 | [all-types](examples/03-all-types/) | All supported types + generics |
| 04 | [concurrent](examples/04-concurrent/) | Concurrent safety |
| 05 | [testing-mock](examples/05-testing-mock/) | Mock testing patterns |
| 06 | [web-app](examples/06-web-app/) | Production web app with SOLID principles |
| 07 | [lifecycle-cleanup](examples/07-lifecycle-cleanup/) | Resource cleanup with hooks |
| 08 | [nested-container-hooks](examples/08-nested-container-hooks/) | Multi-level container hooks |
| 09 | [runtime-container-add](examples/09-runtime-container-add/) | Dynamic container registration |

## 🤝 Contributing

Contributions are welcome!

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
