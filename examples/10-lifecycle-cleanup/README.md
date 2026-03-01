# Lifecycle Cleanup Example

Demonstrates resource cleanup and graceful shutdown using lifecycle hooks.

## What This Example Shows

- Creating a lifecycle manager
- Registering shutdown hooks during dependency initialization
- Executing cleanup in reverse order
- Graceful shutdown with context timeout
- Constructor-based dependency assembly

## Key Concepts

```go
// Lifecycle manager
type Lifecycle struct {
    hooks []func(context.Context) error
}

func (l *Lifecycle) AddShutdownHook(hook func(context.Context) error) {
    l.hooks = append(l.hooks, hook)
}

func (l *Lifecycle) Shutdown(ctx context.Context) error {
    // Execute hooks in reverse order
    for i := len(l.hooks) - 1; i >= 0; i-- {
        l.hooks[i](ctx)
    }
}

// Register dependencies with cleanup hooks
c.MustAdd(
    godi.Provide(lifecycle),
    godi.Build(func(c *godi.Container) (*Database, error) {
        db := &Database{name: "main-db"}
        lifecycle.AddShutdownHook(func(ctx context.Context) error {
            return db.Close()
        })
        return db, nil
    }),
    godi.Build(func(c *godi.Container) (*App, error) {
        db, _ := godi.Inject[*Database](c)
        cache, _ := godi.Inject[*Cache](c)
        service, _ := godi.Inject[*Service](c)
        lc, _ := godi.Inject[*Lifecycle](c)
        return NewApp(db, cache, service, lc), nil
    }),
)

// Graceful shutdown
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
lifecycle.Shutdown(shutdownCtx)
```

## Running the Example

```bash
go run main.go
```

## Output

```
=== Lifecycle Cleanup Example ===

[Database] main-db connected
[Cache] redis-cache connected
[Service] user-service initialized
Application is running...
Press Ctrl+C or wait for timeout to shutdown

=== Starting Shutdown ===
[Service] user-service shutting down gracefully
[Service] user-service shutdown complete
[Cache] redis-cache connection closed
[Database] main-db connection closed
=== Shutdown Complete ===
```

## Cleanup Order

Hooks execute in **reverse order** (LIFO):
1. Service shutdown (first registered, last to close)
2. Cache close
3. Database close (last registered, first to close)

## When to Use

- Database connections
- Network connections
- File handles
- Background goroutines
- Any resource requiring cleanup
