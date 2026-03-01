# Multi-Container Example

Demonstrates dependency injection across multiple containers.

## What This Example Shows

- Creating multiple isolated containers
- Injecting from a single container
- Injecting across multiple containers
- Using `InjectTo` across containers

## Key Concepts

```go
// Create multiple containers
dbContainer := &godi.Container{}
cacheContainer := &godi.Container{}

// Register dependencies in different containers
dbContainer.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
cacheContainer.MustAdd(godi.Provide(Cache{Host: "redis://localhost"}))

// Inject from single container
db, _ := godi.Inject[Database](dbContainer)

// Inject across multiple containers (searches in order)
cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
```

## Running the Example

```bash
go run main.go
```

## Output

```
Database: mysql://localhost:3306/mydb
Cache: redis://localhost:6379
Application: multi-container-app
```

## When to Use

- Modular applications with separate components
- Testing with isolated dependency scopes
- Microservices with shared configuration
- Plugin architectures
