# Struct Field Inject Example

Demonstrates using `Container.Inject` to inject dependencies directly into struct fields.

## Features

- Inject directly into struct fields
- Cleaner code when building complex dependencies
- Error handling for any failed injection

## Usage

```go
type Service struct {
    DB     Database
    Config Config
    Cache  Cache
}

c := &godi.Container{}
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// Inject directly into struct fields
service := &Service{}
err := c.Inject(&service.DB, &service.Config, &service.Cache)
```

## Run

```bash
go run main.go
```

## Output

```
=== Container.Inject Example ===

Service created successfully!
  Database: mysql://localhost:3306/mydb
  App: inject-app
  Cache: redis://localhost:6379

=== Compare: Traditional vs Container.Inject ===

Traditional approach (multiple Inject calls):
  Got: DB=postgres://localhost, App=traditional-app, Cache=redis://remote:6379

Container.Inject approach (single call):
  Got: DB=postgres://localhost, App=traditional-app, Cache=redis://remote:6379

=== Demo Complete ===
```
