// Package godi is a lightweight Go dependency injection framework built on generics.
// Zero reflection, zero code generation.
//
// # Core Architecture
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                         Container                                │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │                    providers (sync.Map)                  │   │
//	│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  │   │
//	│  │  │    Type A    │  │    Type B    │  │    Type C    │  │   │
//	│  │  │   Provider   │  │   Provider   │  │   Provider   │  │   │
//	│  │  └──────────────┘  └──────────────┘  └──────────────┘  │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │                     hooks (sync.Map)                     │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────────────┘
//	                              ▲
//	                              │ Add()
//	                              │
//	┌─────────────────────────────────────────────────────────────────┐
//	│                        Provider                                  │
//	│  ┌─────────────────┐           ┌─────────────────┐             │
//	│  │  Provide(value) │           │  Build(func)    │             │
//	│  │  - Register     │           │  - Register     │             │
//	│  │    instance     │           │    factory      │             │
//	│  │  - Immediate    │           │  - Lazy loading │             │
//	│  │    storage      │           │    singleton    │             │
//	│  └─────────────────┘           └─────────────────┘             │
//	└─────────────────────────────────────────────────────────────────┘
//
// # Registration Flow
//
//	┌──────────────┐     ┌──────────────┐     ┌──────────────┐
//	│  Create       │ ──▶ │  Add          │ ──▶ │   Freeze     │
//	│  Container    │     │  Provider     │     │ (add to      │
//	│  &Container{} │     │  c.Add(...)   │     │  parent)     │
//	└──────────────┘     └──────────────┘     └──────────────┘
//
//	Example:
//	  c := &godi.Container{}
//	  c.Add(
//	      godi.Provide(Config{DSN: "mysql://localhost"}),
//	      godi.Build(func(c *godi.Container) (*Database, error) {
//	          cfg, _ := godi.Inject[Config](c)
//	          return &Database{Conn: cfg.DSN}, nil
//	      }),
//	  )
//
// # Injection Flow
//
//	┌──────────────┐     ┌──────────────┐     ┌──────────────┐
//	│  Call Inject  │ ──▶ │  Find         │ ──▶ │  Execute     │
//	│  Inject[T](c) │     │  Provider     │     │  Factory     │
//	│               │     │  providers.   │     │  Build(func) │
//	│               │     │  Load         │     │              │
//	└──────────────┘     └──────────────┘     └──────────────┘
//	                             │                    │
//	                             ▼                    ▼
//	┌──────────────┐     ┌──────────────┐     ┌──────────────┐
//	│  Trigger Hook │ ◀── │  Return       │ ◀── │  Cache       │
//	│  hooks.Range  │     │  Instance     │     │  sync.Once   │
//	│               │     │  (singleton)  │     │              │
//	└──────────────┘     └──────────────┘     └──────────────┘
//
//	Example:
//	  // Method 1: Generic injection
//	  db, err := godi.Inject[*Database](c)
//
//	  // Method 2: Inject to existing variable
//	  var db Database
//	  err := godi.InjectTo(c, &db)
//
//	  // Method 3: Must mode (panics on failure)
//	  db := godi.MustInject[*Database](c)
//
// # Nested Container Architecture
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                      Parent Container                           │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │  providers: [Config] [Child Container]                   │   │
//	│  │  hooks: ["shutdown" -> cleanupFn]                        │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	│                              │                                  │
//	│                              ▼ Add(child)                       │
//	│  ┌─────────────────────────────────────────────────────────┐   │
//	│  │                    Child Container                       │   │
//	│  │  ┌─────────────────────────────────────────────────┐    │   │
//	│  │  │  providers: [Database] [Cache] [Service]         │    │   │
//	│  │  │  hooks: ["startup" -> initFn]                    │    │   │
//	│  │  └─────────────────────────────────────────────────┘    │   │
//	│  │                              │                           │   │
//	│  │                              ▼ Add(grandchild)           │   │
//	│  │  ┌─────────────────────────────────────────────────┐    │   │
//	│  │  │                 Grandchild Container              │    │   │
//	│  │  │  providers: [Logger] [Metrics]                   │    │   │
//	│  │  └─────────────────────────────────────────────────┘    │   │
//	│  └─────────────────────────────────────────────────────────┘   │
//	└─────────────────────────────────────────────────────────────────┘
//
//	Injection search path (depth-first):
//	  Parent.Inject[*Database]()
//	    │
//	    ├─▶ Parent.providers.Load(*Database) ──▶ Not found
//	    │
//	    ├─▶ Child Container
//	    │     │
//	    │     ├─▶ Child.providers.Load(*Database) ──▶ Found! Return
//	    │
//	    └─▶ Grandchild Container (no need to search)
//
//	Example:
//	  // Infrastructure layer
//	  infra := &godi.Container{}
//	  infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
//
//	  // Application layer
//	  app := &godi.Container{}
//	  app.MustAdd(infra, godi.Provide(Config{AppName: "my-app"}))
//
//	  // Inject from parent (searches in child)
//	  db, _ := godi.Inject[Database](app)
//
// # Hook Lifecycle
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    Hook Execution Flow                          │
//	└─────────────────────────────────────────────────────────────────┘
//
//	Registration Phase:              Execution Phase:
//	┌──────────────┐                  ┌──────────────┐
//	│  c.HookOnce  │                  │  Inject      │
//	│  ("startup", │                  │  Inject[T](c)│
//	│   hookFn)    │                  └──────────────┘
//	└──────────────┘                         │
//	     │                                   ▼
//	     ▼                          ┌──────────────┐
//	┌──────────────┐                │  Trigger Hook│
//	│  Returns     │                │  hooks.Range │
//	│  executor    │                └──────────────┘
//	│  startup     │                       │
//	└──────────────┘                       ▼
//	                               ┌──────────────┐
//	┌──────────────┐               │  Call        │
//	│  Execute Hook│◀──────────────│  executor.   │
//	│  startup(fn) │               │  Iterate()   │
//	└──────────────┘               └──────────────┘
//
//	Example:
//	  c := &godi.Container{}
//
//	  // Register shutdown hook (before injecting dependencies)
//	  shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
//	      return func(ctx context.Context) {
//	          if closer, ok := v.(interface{ Close() error }); ok {
//	              closer.Close()
//	          }
//	      }
//	  })
//
//	  // Add dependencies
//	  c.MustAdd(
//	      godi.Build(func(c *godi.Container) (*Database, error) {
//	          return NewDatabase("dsn")
//	      }),
//	  )
//
//	  // Inject dependencies (hook is registered)
//	  _, _ = godi.Inject[*Database](c)
//
//	  // Execute hook
//	  shutdown.Iterate(ctx, false) // false = forward order
//
// # Concurrency Safety
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│                 Concurrency Safety Design                       │
//	└─────────────────────────────────────────────────────────────────┘
//
//	providers: sync.Map          // Lock-free read/write
//	hooks:     sync.Map          // Lock-free read/write
//	Build:     sync.Once         // Lazy loading singleton
//
//	Concurrent injection scenario:
//	  Goroutine 1 ──┐
//	  Goroutine 2 ──┼──▶ Inject[*Database](c) ──▶ sync.Once.Do() ──▶ Execute once
//	  Goroutine 3 ──┘                              │
//	                                               ▼
//	                                        ┌──────────────┐
//	                                        │ Return cached│
//	                                        │ instance     │
//	                                        │ (thread-safe)│
//	                                        └──────────────┘
//
//	Example:
//	  var wg sync.WaitGroup
//	  for i := 0; i < 100; i++ {
//	      wg.Add(1)
//	      go func() {
//	          defer wg.Done()
//	          db, _ := godi.Inject[*Database](c)
//	          _ = db
//	      }()
//	  }
//	  wg.Wait()
//
// # Circular Dependency Detection
//
//	┌─────────────────────────────────────────────────────────────────┐
//	│              Circular Dependency Detection                      │
//	└─────────────────────────────────────────────────────────────────┘
//
//	Normal dependency chain:     Circular dependency chain:
//	  A ──▶ B ──▶ C               A ──▶ B ──▶ C
//	  │                           ▲         │
//	  └───────────────────────────┴─────────┘
//
//	Detection mechanism:
//	  1. Create temporary container nc during injection, mark injecting type
//	  2. If marked type is encountered in injection path, return circular dependency error
//	  3. Clean up markers after injection completes
//
//	Example:
//	  type ServiceA interface{}
//	  type ServiceB interface{}
//
//	  parent := &godi.Container{}
//	  child := &godi.Container{}
//
//	  child.MustAdd(godi.Build(func(c *godi.Container) (ServiceA, error) {
//	      // A depends on B
//	      return godi.Inject[ServiceB](c)
//	  }))
//
//	  parent.MustAdd(child, godi.Build(func(c *godi.Container) (ServiceB, error) {
//	      // B depends on A (circular!)
//	      return godi.Inject[ServiceA](c)
//	  }))
//
//	  // Trigger circular dependency detection
//	  _, err := godi.Inject[ServiceA](parent)
//	  // err: circular dependency for ServiceA
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
//	type Config struct {
//	    DSN string
//	}
//
//	type Database struct {
//	    Conn string
//	}
//
//	type Service struct {
//	    DB *Database
//	}
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
//	        godi.Build(func(c *godi.Container) (*Database, error) {
//	            cfg, _ := godi.Inject[Config](c)
//	            return &Database{Conn: cfg.DSN}, nil
//	        }),
//	        godi.Build(func(c *godi.Container) (*Service, error) {
//	            db, _ := godi.Inject[*Database](c)
//	            return &Service{DB: db}, nil
//	        }),
//	    )
//
//	    // Inject dependencies
//	    svc := godi.MustInject[*Service](c)
//	    fmt.Printf("Service DB: %s\n", svc.DB.Conn)
//
//	    // Execute cleanup hook
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
//	┌─────────────────────────────────────────────────────────────────┐
//	│                    Framework Comparison                         │
//	├──────────────┬──────────┬──────────┬──────────┬───────────────┤
//	│    Feature   │   godi   │  dig/fx  │   wire   │  samber/do    │
//	├──────────────┼──────────┼──────────┼──────────┼───────────────┤
//	│   Type System│ Generics │Reflection│ Code Gen │   Generics    │
//	│ Runtime Error│    No    │ Possible │    No    │    Possible   │
//	│ Build Step   │    No    │    No    │ Required │      No       │
//	│ API Style    │Functional│Functional│ Code Gen │   Functional  │
//	│ Learn Curve  │   Low    │  Medium  │   High   │      Low      │
//	│ Nesting      │    ✅    │    ❌    │    ❌    │      ❌       │
//	└──────────────┴──────────┴──────────┴──────────┴───────────────┘
package godi
