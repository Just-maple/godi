# Testing with Mocks Example

Demonstrates using dependency injection for testable code with interfaces and mocks.

## What This Example Shows

- Defining interfaces for dependencies
- Creating mock implementations
- Swapping real and mock implementations via containers
- Constructor-based dependency assembly

## Key Concepts

```go
// Define interface
type Database interface {
    Query(sql string) ([]map[string]interface{}, error)
}

// Real implementation
type RealDatabase struct {
    DSN string
}

// Mock implementation
type MockDatabase struct {
    Data []map[string]interface{}
}

// Service depends on interface
type UserService struct {
    DB Database
}

func NewUserService(db Database) *UserService {
    return &UserService{DB: db}
}

// Production container with real database
prodContainer.MustAdd(
    godi.Provide(func() Database {
        return &RealDatabase{DSN: "mysql://localhost/prod"}
    }()),
    godi.Build(func(c *godi.Container) (*UserService, error) {
        db, _ := godi.Inject[Database](c)
        return NewUserService(db), nil
    }),
)

// Test container with mock
testContainer.MustAdd(
    godi.Provide(func() Database {
        return &MockDatabase{Data: mockData}
    }()),
    godi.Build(func(c *godi.Container) (*UserService, error) {
        db, _ := godi.Inject[Database](c)
        return NewUserService(db), nil
    }),
)
```

## Running the Example

```bash
go run main.go
```

## Output

```
Test user: map[email:test@example.com id:1 name:Test User]
Production service ready: true

Demo complete!
- Test environment uses MockDatabase
- Production environment uses RealDatabase
```

## Benefits

- Easy to swap implementations
- Testable without real dependencies
- Follows Dependency Inversion Principle
- No code changes needed for testing
