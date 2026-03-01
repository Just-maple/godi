# Generics Example

Demonstrates dependency injection with generic types.

## What This Example Shows

- Creating generic repositories
- Injecting generic types
- Using type parameters with dependency injection
- Constructor-based dependency assembly

## Key Concepts

```go
// Generic repository
type InMemoryRepository[T any] struct {
    data map[int]T
}

func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
    return &InMemoryRepository[T]{data: make(map[int]T)}
}

// Service with constructor injection
type UserService struct {
    Repo *InMemoryRepository[User]
}

func NewUserService(repo *InMemoryRepository[User]) *UserService {
    return &UserService{Repo: repo}
}

// Register with dependency injection
c.MustAdd(
    godi.Provide(NewInMemoryRepository[User]()),
    godi.Lazy(func(c *godi.Container) (*UserService, error) {
        repo, _ := godi.Inject[*InMemoryRepository[User]](c)
        return NewUserService(repo), nil
    }),
)

// Inject service
userSvc := godi.MustInject[*UserService](c)
```

## Running the Example

```bash
go run main.go
```

## Output

```
User Service Ready: true
Product Service Ready: true

Stored Data:
User 1: {ID:1 Name:Alice}
User 2: {ID:2 Name:Bob}
Product 1: {ID:1 Name:Laptop Price:999.99}
Product 2: {ID:2 Name:Mouse Price:29.99}
```

## When to Use

- Repository pattern with multiple entity types
- Generic data access layers
- Type-safe abstractions
