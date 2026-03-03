# Lifecycle & Hook Example

Demonstrates resource lifecycle management, graceful shutdown, and the Hook system in GoDI.

## What This Example Shows

- Using `HookOnce` for automatic single execution
- Using `Hook` with provided counter for manual control
- Registering cleanup hooks during dependency initialization
- Executing cleanup in reverse order (LIFO)
- Graceful shutdown with context timeout
- Startup and shutdown sequences

## Key Concepts

### HookOnce - Automatic Single Execution

```go
c := &godi.Container{}

// HookOnce: automatically runs only once per type
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            _ = closer.Close()
        }
    }
})

// Inject dependencies (hooks are registered automatically)
_, _ = godi.Inject[*Database](c)
_, _ = godi.Inject[*Cache](c)

// Execute shutdown hooks
shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
shutdown.Iterate(shutdownCtx, true) // true = reverse order (LIFO)
```

### Hook with Provided Counter - Manual Control

```go
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // Skip if already injected before
    }
    return func(ctx context.Context) {
        fmt.Printf("Initializing: %T\n", v)
    }
})

// Execute startup hooks in forward order
startup.Iterate(ctx, false) // false = forward order (FIFO)
```

### Complete Lifecycle Example

```go
c := &godi.Container{}

// Register hooks BEFORE injecting dependencies
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // Handle closable resources
        if closer, ok := v.(interface{ Close() error }); ok {
            _ = closer.Close()
        }
        // Handle shutdownable resources
        if shutdowner, ok := v.(interface{ Shutdown(context.Context) error }); ok {
            _ = shutdowner.Shutdown(ctx)
        }
    }
})

// Register dependencies
c.MustAdd(
    godi.Provide(&Database{name: "main-db"}),
    godi.Provide(&Cache{name: "redis-cache"}),
    godi.Provide(&Service{name: "user-service"}),
)

// Inject (hooks are registered)
_, _ = godi.Inject[*Database](c)
_, _ = godi.Inject[*Cache](c)
_, _ = godi.Inject[*Service](c)

// Run startup hooks
startup.Iterate(ctx, false)

// Run shutdown hooks (reverse order)
shutdown.Iterate(shutdownCtx, true)
```

## Running the Example

```bash
go run main.go
```

## Output

```
=== Lifecycle & Hook Example ===

--- Injecting Dependencies ---

--- Running Startup Hooks ---
  [Startup] Initializing: *Database
  [Startup] Initializing: *Cache
  [Startup] Initializing: *Service

--- Application Running ---
App: DB=main-db, Cache=redis-cache, Service=user-service

--- Running Shutdown Hooks (Reverse Order) ---
  [Service] user-service shutting down gracefully
  [Service] user-service shutdown complete
  [Cache] redis-cache connection closed
  [Database] main-db connection closed

=== Demo Complete ===
```

## Hook Execution Order

### Startup (Forward Order - FIFO)
1. Database init
2. Cache init
3. Service init

### Shutdown (Reverse Order - LIFO)
1. Service shutdown
2. Cache close
3. Database close

## Hook Patterns Summary

| Pattern | Method | Execution | Use Case |
|---------|--------|-----------|----------|
| **Single execution** | `HookOnce` | Automatic once | Cleanup, shutdown |
| **Conditional** | `Hook` + provided | Manual control | Startup with skip logic |
| **Forward order** | `Iterate(ctx, false)` | FIFO | Startup sequence |
| **Reverse order** | `Iterate(ctx, true)` | LIFO | Cleanup sequence |

## When to Use

- Database connections (Close)
- Network connections (Close)
- File handles (Close)
- Background goroutines (Shutdown)
- HTTP servers (Shutdown)
- Any resource requiring cleanup

## Benefits vs Manual Lifecycle

| Feature | Manual Lifecycle | Hook System |
|---------|-----------------|-------------|
| Registration | Explicit `AddShutdownHook` | Automatic on inject |
| Type Safety | Manual type assertions | Type-safe switch |
| Duplication | Manual prevention | Automatic via `HookOnce` |
| Execution Order | Manual management | Built-in FIFO/LIFO |
| Boilerplate | Lifecycle struct needed | Built-in to container |
