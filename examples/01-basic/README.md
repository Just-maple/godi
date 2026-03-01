# Basic Example

Demonstrates the core functionality of GoDI dependency injection framework.

## What This Example Shows

- Creating a container
- Registering dependencies with `Provide`
- Injecting dependencies with `Inject`
- Registering multiple dependencies at once with `MustAdd`

## Key Concepts

```go
// Create container
c := &godi.Container{}

// Register dependencies
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}),
    godi.Provide(Config{AppName: "my-app"}),
)

// Inject dependencies
db, err := godi.Inject[Database](c)
cfg, err := godi.Inject[Config](c)
```

## Running the Example

```bash
go run main.go
```

## Output

```
Connected to mysql://localhost:3306/mydb for my-app
```

## When to Use

Use this pattern for simple dependencies that don't require complex initialization logic.
