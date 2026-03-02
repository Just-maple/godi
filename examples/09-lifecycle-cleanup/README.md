# Lifecycle Cleanup Example

Demonstrates resource cleanup and graceful shutdown using GoDI Hook system.

## What This Example Shows

- Using `HookOnce` for automatic single execution
- Registering cleanup hooks during dependency initialization
- Executing cleanup in reverse order (LIFO)
- Graceful shutdown with context timeout

## Key Concepts

```go
c := &godi.Container{}

// Use HookOnce for automatic single execution cleanup
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            _ = closer.Close()
        }
    }
})

// Register dependencies (hooks are automatically registered when injected)
c.MustAdd(
    godi.Build(func(c *godi.Container) (*Database, error) {
        db := &Database{name: "main-db"}
        return db, nil
    }),
    godi.Build(func(c *godi.Container) (*Cache, error) {
        cache := &Cache{name: "redis-cache"}
        return cache, nil
    }),
)

// Execute shutdown hooks in reverse order
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

// Option 1: Manual iteration (reverse order)
shutdown(func(hooks []func(context.Context)) {
    for i := len(hooks) - 1; i >= 0; i-- {
        hooks[i](shutdownCtx)
    }
})

// Option 2: Using Iterate helper (recommended)
shutdown.Iterate(shutdownCtx, true) // true = reverse order (LIFO)
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

=== Demo Complete ===
```

## Cleanup Order

Hooks execute in **reverse order** (LIFO):
1. Service shutdown (first injected, last to close)
2. Cache close
3. Database close (last injected, first to close)

## When to Use

- Database connections
- Network connections
- File handles
- Background goroutines
- Any resource requiring cleanup

## Benefits vs Manual Lifecycle

| Feature | Manual Lifecycle | Hook System |
|---------|-----------------|-------------|
| Registration | Explicit `AddShutdownHook` | Automatic on inject |
| Type Safety | Manual type assertions | Type-safe switch |
| Duplication | Manual prevention | Automatic via `HookOnce` |
| Boilerplate | Lifecycle struct needed | Built-in to container |
