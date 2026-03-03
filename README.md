# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)
[![Test](https://github.com/Just-maple/godi/actions/workflows/test.yml/badge.svg)](https://github.com/Just-maple/godi/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Just-maple/godi/branch/master/graph/badge.svg)](https://codecov.io/gh/Just-maple/godi)
[![Go Version](https://img.shields.io/github/go-mod/go-version/Just-maple/godi)](https://github.com/Just-maple/godi)
[![License](https://img.shields.io/github/license/Just-maple/godi)](https://github.com/Just-maple/godi/blob/master/LICENSE)

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
| **Container Nesting** | Tree-structured container composition with duplicate detection |

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
| `Build(func) (T, error)` | Register factory function (lazy, singleton) | Complex initialization |

```go
c := &godi.Container{}

// Provide - Register instance value
c.Add(godi.Provide(Config{Port: 8080}))

// Build - Register factory function (lazy, singleton)
// Pattern 1: Single dependency (auto-injected)
c.Add(godi.Build(func(cfg Config) (*Database, error) {
    return NewDatabase(cfg.DSN)
}))

// Pattern 2: Container access (for multiple dependencies)
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
service := &Service{}
err = c.Inject(&service.DB, &service.Config)
```

### Lifecycle Hooks

Hooks allow registering callbacks that execute when dependencies are injected. Hooks are **explicitly executed** - you must call the returned executor function.

```go
package main

import (
    "context"
    "fmt"
    "github.com/Just-maple/godi"
)

c := &godi.Container{}

// Register dependencies
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// Hook with execution counter - runs on every injection
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // Skip if already injected before
    }
    return func(ctx context.Context) {
        fmt.Printf("Starting: %T\n", v)
    }
})

// HookOnce - automatically runs only on first injection
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("Stopping: %T\n", v)
    }
})

// Inject dependencies (this registers the hooks)
_, _ = godi.Inject[Database](c)
_, _ = godi.Inject[Cache](c)

// Execute hooks explicitly
ctx := context.Background()

// Option 1: Manual iteration
startup(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(ctx)
    }
})

// Option 2: Using Iterate helper (recommended)
startup.Iterate(ctx, false) // false = forward order
shutdown.Iterate(ctx, false)

// Output:
// Starting: Database
// Starting: Cache
// Stopping: Database
// Stopping: Cache
```

**Hook Mechanisms:**

| Aspect | Behavior |
|--------|----------|
| **Trigger Point** | Hooks register when dependency is injected |
| **Execution** | Explicit - must call the returned executor function |
| **`provided` Counter** | Tracks how many times a type has been injected (0 = first time) |
| **HookOnce** | Automatically skips when `provided > 0` |
| **Hook** | Manual control via `provided` parameter |
| **Execution Order** | Hooks execute in registration order |
| **Nested Containers** | Hooks trigger on each container in the injection path |

**Hook Behavior in Nested Containers:**

Hooks are triggered on **each container in the injection path**. Each container maintains its own `provided` counter per type:

```go
// Infrastructure layer
infra := &godi.Container{}
infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

infraHook := infra.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) {
        fmt.Printf("[Infra] Starting: %T\n", v)
    }
})

// Application layer
app := &godi.Container{}
app.MustAdd(infra)

appHook := app.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) {
        fmt.Printf("[App] Starting: %T\n", v)
    }
})

// Inject from parent - triggers hooks on BOTH containers
_, _ = godi.Inject[Database](app)

// Execute hooks for each container separately
// Option 1: Manual iteration
infraHook(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(context.Background())
    }
})

appHook(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(context.Background())
    }
})

// Option 2: Using Iterate helper (recommended)
ctx := context.Background()
infraHook.Iterate(ctx, false)
appHook.Iterate(ctx, false)

// Output:
// [Infra] Starting: Database
// [App] Starting: Database
```

**Key Points:**
- Hooks trigger on **each container** in the injection chain
- Each container maintains its **own `provided` counter** per type
- Use `provided > 0` check to run hooks only on first injection
- Execute hooks for each container separately
- Hooks are registered during injection, executed when you call the executor

**Common Patterns:**

```go
// 1. Resource Cleanup with HookOnce
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// 2. Conditional Initialization
// Option A: Using HookOnce (recommended for simple cases)
startup := c.HookOnce("startup", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // Initialize resource - only runs once automatically
    }
})

// Option B: Using Hook with provided check (for conditional logic)
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // Only initialize on first injection
    }
    return func(ctx context.Context) {
        // Initialize resource
    }
})

// 3. Interface-based Cleanup
c.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        switch resource := v.(type) {
        case Database:
            resource.Close()
        case Cache:
            resource.Disconnect()
        }
    }
})

// 4. Graceful Shutdown with Reverse Order
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // Cleanup logic
    }
})

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Execute in reverse order for proper cleanup
shutdown.Iterate(shutdownCtx, true) // true = reverse order

// 5. Multi-Phase Lifecycle
// Use HookOnce for each phase (recommended)
init := c.HookOnce("init", func(v any) func(context.Context) {
    return func(ctx context.Context) { /* init */ }
})

start := c.HookOnce("start", func(v any) func(context.Context) {
    return func(ctx context.Context) { /* start */ }
})

// Or use Hook with provided check for conditional logic per phase
init := c.Hook("init", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) { /* init */ }
})

start := c.Hook("start", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) { /* start */ }
})

// Execute phases in order
ctx := context.Background()
init.Iterate(ctx, false)
start.Iterate(ctx, false)
```

### Container Nesting

Container nesting allows building modular, tree-structured applications with automatic duplicate detection and container freezing.

```go
// Child container
child := &godi.Container{}
child.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

// Parent container with nested child
parent := &godi.Container{}
parent.MustAdd(child)

// Inject from parent (finds Database in child)
db, _ := godi.Inject[Database](parent)

// Adding duplicate types is prevented
err := parent.Add(godi.Provide(Database{DSN: "other"}))
// err: provider *godi.Database already exists
```

**Key Mechanisms:**

| Mechanism | Behavior |
|-----------|----------|
| **Tree Search** | Inject traverses child containers depth-first |
| **Duplicate Detection** | Add checks all nested containers for existing types |
| **Container Freezing** | Child containers are frozen after being added as provider |
| **Hook Propagation** | Hooks trigger on each container in the injection path |
| **Per-Container Counters** | Each container tracks its own `provided` count per type |

**Container Freezing:**

When a container is added to a parent, it becomes **frozen** and cannot accept new providers:

```go
child := &godi.Container{}
child.MustAdd(godi.Provide(Config{Value: "child"}))

parent := &godi.Container{}
parent.MustAdd(child)  // child is now frozen

// This will fail: "container is frozen: already provided to container"
err := child.Add(godi.Provide(struct{ Name string }{Name: "new"}))
```

**Provide Method:**

Check if a type is provided by the container (including nested containers):

```go
c := &godi.Container{}
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

db := Database{}
_, ok := c.Provide(&db)
fmt.Println(ok)  // true

other := struct{ Other string }{}
_, ok = c.Provide(&other)
fmt.Println(ok)  // false
```

**Hook Behavior in Nested Containers:**

See [Lifecycle Hooks](#lifecycle-hooks) for detailed hook behavior in nested containers. In summary:
- Hooks trigger on **each container** in the injection path
- Each container maintains **independent `provided` counters** per type
- Execute hooks for each container separately using their executor functions

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

// Register shutdown hook before injecting dependencies
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
    godi.Build(func(c *godi.Container) (*Cache, error) {
        return NewCache("redis://localhost")
    }),
)

// Inject dependencies (hooks are registered)
_, _ = godi.Inject[*Database](c)
_, _ = godi.Inject[*Cache](c)

// Graceful shutdown with timeout
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Execute hooks in reverse order for proper cleanup
shutdown.Iterate(shutdownCtx, true) // true = reverse order
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

### 7. Container Nesting

```go
// Infrastructure layer
infra := &godi.Container{}
infra.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// Register hooks on infra container
infraShutdown := infra.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("Infra cleanup: %T\n", v)
    }
})

// Application layer
app := &godi.Container{}
app.MustAdd(infra, godi.Provide(Config{AppName: "my-app"}))

// Register hooks on app container
appShutdown := app.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("App cleanup: %T\n", v)
    }
})

// Inject all from parent
db, _ := godi.Inject[Database](app)
cache, _ := godi.Inject[Cache](app)
cfg, _ := godi.Inject[Config](app)

// Execute hooks for each container
ctx := context.Background()
infraShutdown.Iterate(ctx, false)
appShutdown.Iterate(ctx, false)
```

### 8. Build with Dependency Chain

```go
type Name string
type Length int

c.MustAdd(
    godi.Provide(Name("hello")),
    godi.Build(func(s Name) (Length, error) {
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
| **Container Nesting** | ✅ | ❌ | ❌ | ❌ |
| **Project Status** | Active | Active | ⚠️ Archived | Active |

### When to Choose godi

- You prefer **compile-time safety** without code generation
- You want **minimal dependencies** and small bundle size
- You need **lifecycle management** for resources
- You value **simple, intuitive API**
- You need **container nesting** for modular architecture

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
| 13 | [nested-container-hooks](examples/13-nested-container-hooks/) | Multi-level container hooks |

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) for details.
