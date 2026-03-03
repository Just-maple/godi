// Package godi is a lightweight Go dependency injection framework built on generics.
// Zero reflection, zero code generation.
//
// # Core Philosophy
//
// Godi follows three design principles:
//   - Type Safety: Full generics support with compile-time type checking
//   - Lazy Loading: Dependencies initialized on first use (singleton pattern)
//   - Explicit Control: No magic - you control when and how injection happens
//
// # Quick Start
//
//	c := &godi.Container{}
//	c.MustAdd(
//	    godi.Provide(Config{DSN: "mysql://localhost"}),
//	    godi.Build(func(cfg Config) (*Database, error) {
//	        return &Database{Conn: cfg.DSN}, nil
//	    }),
//	)
//	db := godi.MustInject[*Database](c)
//
// # Container Architecture
//
// Container is the core structure that manages providers and handles injection.
// It uses sync.Map for thread-safe concurrent access.
//
//	┌─────────────────────────────────────────────────────────┐
//	│                     Container                            │
//	│  ┌─────────────────────────────────────────────────┐   │
//	│  │            providers (sync.Map)                  │   │
//	│  │  Maps type → Provider (Provide/Build)            │   │
//	│  └─────────────────────────────────────────────────┘   │
//	│  ┌─────────────────────────────────────────────────┐   │
//	│  │              hooks (sync.Map)                    │   │
//	│  │  Maps hook name → hook functions                 │   │
//	│  └─────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────┘
//
// # Registration: Provide vs Build
//
// Provide registers an existing value (immediate storage):
//
//	c.Add(godi.Provide(Config{Port: 8080}))
//
// Build registers a factory function (lazy singleton):
//
//	c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
//	    cfg, _ := godi.Inject[Config](c)
//	    return NewDatabase(cfg.DSN), nil
//	}))
//
// Build supports three dependency patterns:
//
// Pattern 1: Single dependency (auto-injected)
//
//	godi.Build(func(cfg Config) (*Database, error) {
//	    return NewDatabase(cfg.DSN), nil
//	})
//
// Pattern 2: Container access (manual multi-inject)
//
//	godi.Build(func(c *godi.Container) (*Service, error) {
//	    db, _ := godi.Inject[*Database](c)
//	    cache, _ := godi.Inject[*Cache](c)
//	    return NewService(db, cache), nil
//	})
//
// Pattern 3: No dependency (using struct{})
//
//	godi.Build(func(_ struct{}) (*Logger, error) {
//	    return NewLogger(), nil
//	})
//
// # Injection Methods
//
// Generic injection (returns value + error):
//
//	db, err := godi.Inject[*Database](c)
//
// Must mode (panics on error):
//
//	db := godi.MustInject[*Database](c)
//
// Inject to existing variable:
//
//	var db Database
//	err := godi.InjectTo(c, &db)
//
// Multi-injection:
//
//	err := c.Inject(&db, &cfg, &cache)
//
// # Container Nesting
//
// Containers can be nested to create modular, tree-structured applications.
// Child containers become frozen (read-only) after being added to parent.
//
//	┌─────────────────────────────────────────────────────────┐
//	│                  Parent Container                        │
//	│  providers: [Config] [Child Container]                  │
//	│  ┌─────────────────────────────────────────────────┐   │
//	│  │              Child Container                     │   │
//	│  │  providers: [Database] [Cache] [Service]         │   │
//	│  │  ┌─────────────────────────────────────────┐    │   │
//	│  │  │          Grandchild Container            │    │   │
//	│  │  │  providers: [Logger] [Metrics]           │    │   │
//	│  │  └─────────────────────────────────────────┘    │   │
//	│  └─────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────┘
//
// Injection search path (depth-first):
//
//	infra := &godi.Container{}
//	infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
//
//	app := &godi.Container{}
//	app.MustAdd(infra, godi.Provide(Config{AppName: "my-app"}))
//
//	// Inject from parent (searches in child)
//	db, _ := godi.Inject[Database](app)
//
// # Container Freezing
//
// When a container is added to a parent, it becomes frozen and cannot accept new providers.
// However, Build functions CAN add containers at runtime (special case).
//
// Frozen (cannot add):
//
//	child := &godi.Container{}
//	parent.MustAdd(child)  // child is frozen
//	child.Add(...)         // ERROR: container frozen
//
// Runtime Add in Build (allowed):
//
//	c.MustAdd(godi.Build(func(c *godi.Container) (T, error) {
//	    c.MustAdd(nestedContainer)  // OK: runtime add
//	    return godi.Inject[T](c)
//	}))
//
// # Hook Lifecycle
//
// Hooks allow registering callbacks that execute when dependencies are injected.
// Hooks are explicitly executed - you must call the returned executor function.
//
// Registration (before injection):
//
//	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
//	    return func(ctx context.Context) {
//	        if closer, ok := v.(interface{ Close() error }); ok {
//	            closer.Close()
//	        }
//	    }
//	})
//
// Injection (hooks are registered):
//
//	_, _ = godi.Inject[Database](c)
//
// Execution (explicit call):
//
//	shutdown.Iterate(ctx, false)  // false = forward order
//
// Hook Behavior in Nested Containers:
// - Hooks trigger on each container in the injection path
// - Each container maintains independent `provided` counters per type
// - Execute hooks for each container separately
//
// # Circular Dependency Detection
//
// Godi automatically detects circular dependencies at runtime.
//
// Normal chain: A → B → C (OK)
// Circular: A → B → C → A (ERROR)
//
// Detection mechanism:
//  1. Create temporary container context during injection
//  2. Mark types being injected with depth tracking
//  3. If marked type encountered, return circular dependency error
//  4. Clean up markers after injection completes
//
// # Concurrency Safety
//
// All operations are thread-safe:
//   - providers: sync.Map (lock-free read/write)
//   - hooks: sync.Map (lock-free read/write)
//   - Build: sync.Once (lazy singleton)
//
// Concurrent injection scenario:
//
//	var wg sync.WaitGroup
//	for i := 0; i < 100; i++ {
//	    wg.Add(1)
//	    go func() {
//	        defer wg.Done()
//	        db, _ := godi.Inject[*Database](c)  // Safe
//	    }()
//	}
//	wg.Wait()
//
// # Complete Example
//
//	package main
//
//	import (
//	    "context"
//	    "fmt"
//	    "github.com/Just-maple/godi"
//	)
//
//	type Config struct{ DSN string }
//	type Database struct{ Conn string }
//	type Cache struct{ Addr string }
//	type Service struct{ DB *Database; Cache *Cache }
//
//	func main() {
//	    c := &godi.Container{}
//
//	    // Register shutdown hook
//	    shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
//	        return func(ctx context.Context) {
//	            fmt.Printf("cleanup: %T\n", v)
//	        }
//	    })
//
//	    // Add dependencies
//	    c.MustAdd(
//	        godi.Provide(Config{DSN: "mysql://localhost"}),
//	        godi.Build(func(cfg Config) (*Database, error) {
//	            return &Database{Conn: cfg.DSN}, nil
//	        }),
//	        godi.Build(func(_ struct{}) (*Cache, error) {
//	            return &Cache{Addr: "redis://localhost"}, nil
//	        }),
//	        godi.Build(func(c *godi.Container) (*Service, error) {
//	            db, _ := godi.Inject[*Database](c)
//	            cache, _ := godi.Inject[*Cache](c)
//	            return &Service{DB: db, Cache: cache}, nil
//	        }),
//	    )
//
//	    // Inject
//	    svc := godi.MustInject[*Service](c)
//	    fmt.Printf("Service DB: %s, Cache: %s\n", svc.DB.Conn, svc.Cache.Addr)
//
//	    // Cleanup
//	    shutdown.Iterate(context.Background(), false)
//	}
//
// # Supported Types
//
//   - Structs: Database, Config
//   - Primitives: string, int, bool, float64
//   - Pointers: *Database
//   - Slices: []string
//   - Maps: map[string]int
//   - Interfaces: any, custom interfaces
//   - Arrays: [3]int
//   - Channels: chan int
//   - Functions: func() error
//
// # Framework Comparison
//
//	┌──────────────┬──────────┬──────────┬──────────┬───────────────┐
//	│    Feature   │   godi   │  dig/fx  │   wire   │  samber/do    │
//	├──────────────┼──────────┼──────────┼──────────┼───────────────┤
//	│   Type System│ Generics │Reflection│ Code Gen │   Generics    │
//	│ Runtime Error│    No    │ Possible │    No    │    Possible   │
//	│ Build Step   │    No    │    No    │ Required │      No       │
//	│ Nesting      │    ✅    │    ❌    │    ❌    │      ❌       │
//	│ Learn Curve  │   Low    │  Medium  │   High   │      Low      │
//	└──────────────┴──────────┴──────────┴──────────┴───────────────┘
package godi
