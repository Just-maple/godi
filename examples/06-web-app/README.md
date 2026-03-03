# Web Application Example - Best Practices

This example demonstrates a production-ready Go web application structure using the godi dependency injection framework, following **SOLID principles** especially **Dependency Inversion Principle (DIP)**.

## Key Principles

### 1. Dependency Inversion Principle (DIP)
- High-level modules depend on abstractions (interfaces)
- Low-level modules implement abstractions
- No direct dependencies on concrete implementations

### 2. Separation of Concerns
- Each layer has a single responsibility
- Clear boundaries between layers
- Easy to test and maintain

### 3. Interface-Based Design
- All cross-layer dependencies use interfaces
- Easy to swap implementations
- Facilitates mocking for tests

### 4. Hook-based Lifecycle Management
- Automatic cleanup via `HookOnce`
- No manual lifecycle manager needed
- Type-safe resource cleanup

## Directory Structure

```
examples/06-web-app/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── app/                 # Application orchestration
│   │   └── app.go           # Depends on interfaces.Handler
│   ├── config/              # Configuration management
│   │   └── config.go
│   ├── handler/             # HTTP handlers
│   │   └── user_handler.go  # Depends on service.UserServiceInterface
│   ├── infrastructure/      # External service connections
│   │   └── database.go      # Implements interfaces.Database, interfaces.Cache
│   ├── middleware/          # Request processing pipeline
│   │   └── logging.go       # Implements interfaces.Middleware
│   ├── model/               # Domain models
│   │   └── user.go
│   ├── repository/          # Data access layer
│   │   └── user_repository.go  # Depends on interfaces.Database
│   ├── service/             # Business logic layer
│   │   └── user_service.go  # Depends on interfaces
│   └── wire/                # Dependency injection setup
│       └── wire.go          # Registers interfaces, not concretions
└── pkg/
    └── interfaces/          # Interface definitions (abstractions)
        └── interfaces.go    # All interfaces defined here
```

## Architecture Layers & Dependencies

```
┌─────────────────────────────────────────────────────────┐
│  pkg/interfaces/ (Abstractions)                         │
│  - Database, Cache, Repository, Service, Handler, MW   │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ implements
┌─────────────────────────────────────────────────────────┐
│  internal/infrastructure/ (Concrete Implementations)    │
│  - DBConnection implements interfaces.Database          │
│  - CacheClient implements interfaces.Cache              │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ depends on interfaces
┌─────────────────────────────────────────────────────────┐
│  internal/repository/ (Data Access)                     │
│  - UserRepository depends on interfaces.Database        │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ depends on interfaces
┌─────────────────────────────────────────────────────────┐
│  internal/service/ (Business Logic)                     │
│  - UserService depends on interfaces                    │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ depends on interfaces
┌─────────────────────────────────────────────────────────┐
│  internal/handler/ (HTTP Layer)                         │
│  - UserHandler depends on service.UserServiceInterface  │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ depends on interfaces
┌─────────────────────────────────────────────────────────┐
│  internal/middleware/ (Pipeline)                        │
│  - LoggingMiddleware implements interfaces.Middleware   │
└─────────────────────────────────────────────────────────┘
                          ↑
                          │ depends on interfaces
┌─────────────────────────────────────────────────────────┐
│  internal/app/ (Application)                            │
│  - App depends on interfaces.Handler, interfaces.MW     │
└─────────────────────────────────────────────────────────┘
```

## Dependency Injection Setup

```go
// wire.go - Register abstractions and shutdown hooks

// Register shutdown hook using HookOnce
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // Execute cleanup for closable resources using interface
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// Infrastructure layer - returns interfaces
c.Add(godi.Build(func() (interfaces.Database, error) {
    cfg, _ := godi.ShouldInject[*config.Config](c)
    return infrastructure.NewDBConnection(cfg.DatabaseDSN), nil
}))

// Repository layer - depends on interfaces.Database
c.Add(godi.Build(func() (repository.UserRepositoryInterface, error) {
    db, _ := godi.ShouldInject[interfaces.Database](c)
    return repository.NewUserRepository(db), nil
}))
```

## Running the Example

```bash
cd examples/09-web-app
go run cmd/main.go
```

## Example Output

```
=== Web Application Example ===
Best Practices: Separation of Concerns

✓ Container created
✓ Using Dependency Inversion Principle
✓ Shutdown hooks registered via HookOnce
[Infrastructure] Database connection established: postgres://localhost:5432/mydb
[Infrastructure] Cache client connected: redis://localhost:6379
✓ All dependencies injected

Starting WebApp on port 8080
Debug mode: enabled
[DEBUG] Request started
Handler: Got user User from DB
[DEBUG] Request completed in 15.291µs

=== Starting Graceful Shutdown ===
[Infrastructure] Cache client disconnected: redis://localhost:6379
[Infrastructure] Database connection closed: postgres://localhost:5432/mydb
=== Shutdown Complete ===

=== Demo Complete ===
```

## Benefits of This Architecture

### 1. Testability
```go
// Easy to mock for tests
mockDB := &MockDatabase{}
mockCache := &MockCache{}
repo := repository.NewUserRepository(mockDB)  // Uses interface
```

### 2. Flexibility
```go
// Swap implementations without changing business logic
c.Add(godi.Provide(&RealDatabase{}))   // Production
// OR
c.Add(godi.Provide(&MockDatabase{}))   // Testing
```

### 3. Maintainability
- Changes in infrastructure don't affect business logic
- Each layer can evolve independently
- Clear contracts between layers

### 4. SOLID Compliance
- **S**ingle Responsibility - Each layer has one job
- **O**pen/Closed - Easy to extend without modification
- **L**iskov Substitution - Any implementation works
- **I**nterface Segregation - Small, focused interfaces
- **D**ependency Inversion - Depend on abstractions

## Interface Definitions

```go
// pkg/interfaces/interfaces.go

type Database interface {
    Query(query string, args ...interface{}) ([]map[string]interface{}, error)
    Execute(query string, args ...interface{}) (int64, error)
    Close() error
}

type Cache interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, ttl int) error
    Delete(key string) error
    Close() error
}

type Handler interface {
    Handle(ctx context.Context) error
}

type Middleware interface {
    Process(handler Handler) Handler
}
```

## Best Practices Demonstrated

1. ✅ **Interface Segregation** - Each layer defines its own interfaces
2. ✅ **Dependency Inversion** - High-level modules don't depend on low-level modules
3. ✅ **Single Responsibility** - Each package has one reason to change
4. ✅ **Explicit Dependencies** - All dependencies are clearly declared
5. ✅ **Build Initialization** - Resources created only when needed
6. ✅ **Hook-based Cleanup** - Automatic resource cleanup via `HookOnce`
7. ✅ **English Documentation** - All code documented in English
