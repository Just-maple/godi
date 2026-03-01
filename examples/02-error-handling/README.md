# Error Handling Example

Demonstrates graceful error handling patterns in GoDI.

## What This Example Shows

- Using `Add` instead of `MustAdd` for recoverable errors
- Handling duplicate registration errors
- Handling injection errors for missing dependencies
- Checking errors from `Inject` instead of panicking

## Key Concepts

```go
// Use Add to handle errors gracefully
err := c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
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

## Running the Example

```bash
go run main.go
```

## Output

```
Registration successful
Expected error: dependency already registered
Database: mysql://localhost
Expected error: dependency not found
```

## When to Use

- Use `Add`/`Inject` when errors are recoverable
- Use `MustAdd`/`MustInject` when dependencies are critical and should panic on failure
