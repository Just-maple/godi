# Concurrent Access Example

Demonstrates thread-safe operations in GoDI containers.

## What This Example Shows

- Concurrent injection from multiple goroutines
- Concurrent registration from multiple goroutines
- Thread-safe container operations
- Using sync.WaitGroup with dependency injection

## Key Concepts

```go
c := &godi.Container{}
c.MustAdd(godi.Provide(Counter{Value: 0}))

var wg sync.WaitGroup

// Multiple goroutines injecting concurrently
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        counter, _ := godi.Inject[Counter](c)
        fmt.Printf("Goroutine %d: value=%d\n", id, counter.Value)
    }(i)
}

wg.Wait()

// Concurrent registration
for i := 0; i < 5; i++ {
    wg.Add(1)
    go func(id int) {
        defer wg.Done()
        c.Add(godi.Provide(fmt.Sprintf("value-%d", id)))
    }(i)
}
wg.Wait()
```

## Running the Example

```bash
go run main.go
```

## Thread Safety

GoDI containers are designed for concurrent access:

- Multiple goroutines can safely call `Inject` simultaneously
- Multiple goroutines can safely call `Add` simultaneously
- No external locking required

## When to Use

- HTTP servers handling concurrent requests
- Background job processors
- Event-driven architectures
