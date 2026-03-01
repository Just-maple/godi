# All Types Example

Demonstrates all types supported by GoDI for dependency injection.

## What This Example Shows

- Basic types: string, int, float, bool
- Slices and arrays
- Maps
- Pointers to structs
- Channels
- Functions
- Interfaces

## Key Concepts

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

// Inject any type
str := godi.MustInject[string](c)
fn := godi.MustInject[func() string](c)
```

## Running the Example

```bash
go run main.go
```

## Supported Types

| Category | Types |
|----------|-------|
| Basic | string, int, int8-64, uint, uint8-64, float32, float64, bool |
| Collections | slices, maps, arrays |
| Reference | pointers, channels |
| Advanced | functions, interfaces, structs |

## When to Use

- Configuration values (strings, numbers)
- Feature flags (bools)
- Shared resources (channels)
- Utility functions
