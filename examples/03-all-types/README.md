# All Types & Generics Example

Demonstrates all types supported by GoDI for dependency injection, including generics and interfaces.

## What This Example Shows

- Basic types: string, int, float, bool
- Slices and arrays
- Maps
- Pointers to structs
- Channels
- Functions
- Interfaces
- Generic types and repositories

## Key Concepts

### Basic Types

```go
c := &godi.Container{}

// Basic types
c.MustAdd(
    godi.Provide("application-name"),
    godi.Provide(42),
    godi.Provide(float64(3.14159)),
    godi.Provide(true),
)

// Slices
c.MustAdd(godi.Provide([]string{"a", "b", "c"}))

// Maps
c.MustAdd(godi.Provide(map[string]int{"key": 1}))

// Pointers
c.MustAdd(godi.Provide(&User{Name: "Alice"}))

// Functions
c.MustAdd(godi.Provide(func() string { return "hello" }))
```

### Generics and Interfaces

```go
// Generic repository
type InMemoryRepository[T any] struct {
    data map[int]T
}

func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
    return &InMemoryRepository[T]{data: make(map[int]T)}
}

// Register generic types
c.MustAdd(
    godi.Provide(NewInMemoryRepository[User]()),
    godi.Provide(NewInMemoryRepository[Product]()),
    godi.Build(func(repo *InMemoryRepository[User]) (*UserService, error) {
        return NewUserService(repo), nil
    }),
)

// Inject generic services
userSvc, _ := godi.Inject[*UserService](c)
```

## Running the Example

```bash
go run main.go
```

## Supported Types

| Category | Types |
|----------|-------|
| **Basic** | string, int, int8-64, uint, uint8-64, float32, float64, bool |
| **Collections** | slices, maps, arrays |
| **Reference** | pointers, channels |
| **Advanced** | functions, interfaces, structs |
| **Generics** | Repository[T], Service[T], any type parameter |

## When to Use

| Type | Use Case |
|------|----------|
| **Configuration values** | strings, numbers, bools |
| **Feature flags** | bool values |
| **Shared resources** | channels |
| **Utility functions** | func() types |
| **Generic repositories** | Data access layers with type safety |
| **Interfaces** | Dependency inversion, testing with mocks |

## Generics Benefits

- Type-safe repositories without code generation
- Single implementation for multiple entity types
- Compile-time type checking
- No reflection overhead
