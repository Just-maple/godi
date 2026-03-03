# Error Handling Example

Demonstrates both graceful error handling and panic-on-error patterns in GoDI.

## What This Example Shows

- Using `Add`/`Inject` for recoverable errors (graceful pattern)
- Using `MustAdd`/`MustInject` for critical dependencies (panic pattern)
- Handling duplicate registration errors
- Handling injection errors for missing dependencies
- Method chaining with `MustAdd`

## Key Concepts

### Pattern 1: Graceful Error Handling (Add/Inject)

```go
c := &godi.Container{}

// Use Add to handle duplicate registration errors
err := c.Add(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{Port: 8080}),
)
if err != nil {
    fmt.Printf("Registration failed: %v\n", err)
    return
}

// Use Inject to handle missing dependencies
db, err := godi.Inject[Database](c)
if err != nil {
    fmt.Printf("Injection failed: %v\n", err)
    return
}
```

### Pattern 2: Panic-on-Error (MustAdd/MustInject)

```go
c := &godi.Container{}

// Use MustAdd - panics if duplicate
c.MustAdd(
    godi.Provide(CriticalConfig{SecretKey: "secret"}),
    godi.Provide(Database{DSN: "mysql://localhost"}),
)

// Use MustInject - panics if not found
config := godi.MustInject[CriticalConfig](c)
db := godi.MustInject[Database](c)

// Use MustInjectTo - inject directly into variable
var extraDB Database
godi.MustInjectTo(c, &extraDB)
```

### Method Chaining

```go
c2 := &godi.Container{}
c2.MustAdd(
    godi.Provide(Database{DSN: "chain-db"}),
).MustAdd(
    godi.Provide(Config{Port: 3306}),
)
```

## Running the Example

```bash
go run main.go
```

## Output

```
=== Error Handling Patterns ===

--- Pattern 1: Add/Inject (Graceful) ---
✓ Registration successful
✓ Expected duplicate error: provider Database already exists
✓ Database: mysql://localhost
✓ Expected not-found error: provider CriticalConfig not found

--- Pattern 2: MustAdd/MustInject (Panic) ---
✓ MustAdd successful (panics on duplicate)
✓ Secret Key: super-secret-key
✓ Database: mysql://localhost
✓ Extra Database (MustInjectTo): mysql://localhost
✓ Method chaining with MustAdd
```

## When to Use Each Pattern

### Add/Inject (Graceful)

- Optional dependencies
- User-provided configurations
- Test scenarios
- When you want to handle errors explicitly
- Recovery is possible

### MustAdd/MustInject (Panic)

- Critical dependencies that must exist
- Application startup (fail fast)
- Production code where missing deps are bugs
- Cleaner code when errors are unexpected
- Method chaining convenience

## Error Types

| Error Type | Description | Example |
|------------|-------------|---------|
| Duplicate Registration | Type already registered | `Add(Provide(Database{}))` twice |
| Dependency Not Found | Type not in container | `Inject[MissingType](c)` |
| Build Error | Factory function returned error | `Build(func() (T, error) {...})` |
| Circular Dependency | A depends on B, B depends on A | Nested container cycles |
