# Must Functions Example

Demonstrates panic-on-error injection patterns for critical dependencies.

## What This Example Shows

- Using `MustAdd` for method chaining
- Using `MustInject` for direct injection
- Using `MustInjectTo` for injecting into variables
- When to use panic-on-error patterns

## Key Concepts

```go
// MustAdd returns *Container for method chaining
c.MustAdd(
    godi.Provide(CriticalConfig{SecretKey: "super-secret-key"}),
    godi.Provide(Database{DSN: "mysql://localhost"}),
)

// MustInject - panics if dependency not found
config := godi.MustInject[CriticalConfig](c)
db := godi.MustInject[Database](c)

// MustInjectTo - inject directly into variable
var extraDB Database
godi.MustInjectTo(&extraDB, c)
```

## Running the Example

```bash
go run main.go
```

## Output

```
Secret Key: super-secret-key
Database: mysql://localhost
Extra Database: mysql://localhost
```

## When to Use

- Application startup configuration
- Critical dependencies that must exist
- When you want cleaner code without error checking
- Testing scenarios
