# Hook Example

Demonstrates lifecycle hook management with GoDI.

## What This Example Shows

- Registering lifecycle hooks with `Hook`
- Preventing duplicate hook execution using `provided` parameter
- Using `HookOnce` for automatic single execution

## Key Concepts

```go
c := &godi.Container{}

c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
)

// Hook with provided counter - only execute on first provision
startup := c.Hook("startup", func(v any, provided int) godi.HookFunc {
    if provided > 0 {
        return nil // Skip if already called
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
```

## Running the Example

```bash
go run main.go
```

## Output

```
Starting: main.Config
Starting: main.Database
Starting: main.Service
Running: my-app, mysql://localhost, user-service
Stopping: main.Config
Stopping: main.Database
Stopping: main.Service
```

## When to Use

Use hooks for:
- Application initialization and cleanup
- Resource management (database connections, file handles)
- Logging and monitoring at lifecycle stages
- Graceful shutdown handling

## API

### Hook

```go
func (c *Container) Hook(name string, build func(v any, provided int) HookFunc) func(func([]HookFunc))
```

- `v`: The injected value
- `provided`: Number of times this value has been provided (0 = first time)
- Return `nil` to skip hook execution

### HookOnce

```go
func (c *Container) HookOnce(name string, build func(v any, provided int) HookFunc) func(func([]HookFunc))
```

Automatically skips execution after first provision.
