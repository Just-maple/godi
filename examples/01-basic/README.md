# Basic Example

Demonstrates the core functionality of GoDI dependency injection framework including Provide, Build patterns, struct field injection, and dependency chains.

## What This Example Shows

- Creating a container
- Registering dependencies with `Provide`
- Build patterns (single dependency, container access, no dependency)
- Struct field injection with `Inject`
- Dependency chains
- Injecting dependencies with `Inject`

## Key Concepts

### Pattern 1: Provide - Register Simple Values

```go
c := &godi.Container{}
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
)

db, _ := godi.Inject[Database](c)
cfg, _ := godi.Inject[Config](c)
```

### Pattern 2a: Build with Single Dependency (Auto-Injected)

```go
c.MustAdd(
    godi.Provide(Config{DSN: "postgres://localhost"}),
    godi.Build(func(cfg Config) (Database, error) {
        return Database{DSN: cfg.DSN}, nil
    }),
)
```

### Pattern 2b: Build with Container Access (Multiple Dependencies)

```go
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "multi-dep"}),
    godi.Build(func(c *godi.Container) (Service, error) {
        db, _ := godi.Inject[Database](c)
        cfg, _ := godi.Inject[Config](c)
        return Service{DB: db, Config: cfg}, nil
    }),
)
```

### Pattern 2c: Build with No Dependency (struct{})

```go
c.MustAdd(
    godi.Build(func(_ struct{}) (string, error) {
        return "no-dependency", nil
    }),
)
```

### Pattern 3: Struct Field Injection

```go
service := &Service{}
err := c.Inject(&service.DB, &service.Config, &service.Cache)
```

### Pattern 4: Dependency Chain

```go
type Name string
type Length int
type Result string

c.MustAdd(
    godi.Provide(Name("hello")),
    godi.Build(func(s Name) (Length, error) {
        return Length(len(s)), nil
    }),
    godi.Build(func(n Length) (Result, error) {
        return Result(fmt.Sprintf("len:%d", n)), nil
    }),
)

result, _ := godi.Inject[Result](c) // "len:5"
```

## Running the Example

```bash
go run main.go
```

## Output

```
=== Basic Example ===

Pattern 1 (Provide): mysql://localhost:3306/mydb for my-app, cache: redis://localhost:6379

=== Build Patterns ===

Pattern 2a (Build single dep): postgres://localhost
Pattern 2b (Build multi-dep): mysql://localhost for multi-dep
Pattern 2c (Build no dep): no-dependency

=== Struct Field Injection ===

Pattern 3 (Field Inject): DB=field-inject-db, App=field-inject-app, Cache=field-inject-cache

=== Dependency Chain ===

Pattern 4 (Chain): hello -> len:5

=== Demo Complete ===
```

## When to Use Each Pattern

| Pattern | Use Case |
|---------|----------|
| **Provide** | Simple values, configuration, already-initialized instances |
| **Build (single dep)** | Transform one dependency, type conversions, wrapping |
| **Build (container)** | Multiple dependencies, complex assembly logic |
| **Build (no dep)** | Pure factories, logging, configuration-free services |
| **Field Inject** | Populating struct fields from container |
| **Dependency Chain** | Derived types, transformations, type aliases |
