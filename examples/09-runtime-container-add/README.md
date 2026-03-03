# Runtime Container Add Example

This example demonstrates how to dynamically add containers at runtime based on configuration, environment, or feature flags.

## Key Concept: Runtime Add vs Frozen Container

**Important**: When you add a container to a parent, the child container becomes **frozen** and cannot accept new providers directly. However, **during Build execution**, you CAN add containers at runtime without triggering frozen errors.

```
┌─────────────────────────────────────────────────────────┐
│  Frozen Container (Cannot Add After Parent Addition)   │
│  ❌ child.Add(Provide(...))  // Error: frozen          │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  Runtime Add in Build (Allowed)                         │
│  ✅ c.MustAdd(container)  // OK: inside Build function  │
└─────────────────────────────────────────────────────────┘
```

## Real-World Scenarios

### 1. Environment-Based Database Selection

Different environments (production, development, test) require different database configurations. Instead of complex conditional logic throughout your code, you can register the appropriate database container at runtime.

### 2. Conditional Feature Registration

Features like caching, monitoring, or analytics can be enabled/disabled via configuration. Runtime container add allows you to register these dependencies only when needed.

### 3. Interface Implementation Selection

Choose between real implementations and mocks based on environment - perfect for testing scenarios where you want the same interface but different implementations.

## Container Hierarchy

```
Main Container (c)
├── Config (provided directly)
├── Database/Cache/Repository (built at runtime)
└── Conditional containers (added during build)
    ├── prodDB / devDB / testDB
    ├── redisCache (optional)
    └── mockRepo / mysqlRepo
```

## Running the Example

```bash
cd examples/09-runtime-container-add
go run main.go
```

## Expected Output

```
=== Runtime Container Add - Real World Example ===

--- Scenario 1: Environment-Based Database Selection ---
Build: Detected environment = production
Build: Registering production database container
Build: Connected to mysql://prod-db.company.com:3306/app
Result: Database{DSN: mysql://prod-db.company.com:3306/app}

--- Scenario 2: Conditional Cache Registration ---
Build: Config.EnableCache = true
Build: Registering Redis cache container
Result: Connected to cache: redis://prod-redis.company.com:6379

--- Scenario 3: Interface Implementation Selection ---
Build: Environment = test
Build: Using mock repository for testing
Build: Repository type = *main.MockUserRepository
Result: Mock user 123

--- Cleanup Hooks Demo ---
Build: Adding database container at runtime
Injected: Database{DSN: mysql://localhost:3306/app}

Executing cleanup hooks...
[Cleanup] Database connection closed (call #1)
Verification: Cleanup called 1 time(s) (expected: 1)
```

## Key Patterns

### Pattern 1: Environment-Based Selection

```go
prodDB := &godi.Container{}
prodDB.MustAdd(godi.Provide(Database{DSN: "mysql://prod-db:3306/app"}))

devDB := &godi.Container{}
devDB.MustAdd(godi.Provide(Database{DSN: "mysql://localhost:3306/dev"}))

type EnvConfig struct {
    Env string
}

c.MustAdd(
    godi.Provide(EnvConfig{Env: "production"}),
    godi.Build(func(cfg EnvConfig) (Database, error) {
        // Choose appropriate database based on environment
        if cfg.Env == "production" {
            db, _ := godi.Inject[Database](prodDB)
            return db, nil
        }
        db, _ := godi.Inject[Database](devDB)
        return db, nil
    }),
)
```

### Pattern 2: Conditional Feature Registration

```go
type AppConfig struct {
    EnableCache bool
}

cacheContainer := &godi.Container{}
cacheContainer.MustAdd(godi.Provide(Cache{Addr: "redis://localhost"}))

c.MustAdd(
    godi.Provide(AppConfig{EnableCache: true}),
    godi.Build(func(cfg AppConfig) (Service, error) {
        var cache *Cache
        if cfg.EnableCache {
            cache, _ = godi.Inject[Cache](cacheContainer)
        }
        return NewService(cache), nil
    }),
)
```

### Pattern 3: Interface Implementation Selection

```go
type EnvConfig struct {
    Env string
}

// Mock repository container
mockRepo := &godi.Container{}
mockRepo.MustAdd(godi.Provide(func() UserRepository { return &MockUserRepository{} }()))

// MySQL repository container
mysqlRepo := &godi.Container{}
mysqlRepo.MustAdd(godi.Provide(func() UserRepository { 
    return &MySQLUserRepository{dsn: "mysql://localhost"} 
}()))

c.MustAdd(
    godi.Provide(EnvConfig{Env: "test"}),
    godi.Build(func(cfg EnvConfig) (UserRepository, error) {
        if cfg.Env == "test" {
            repo, _ := godi.Inject[func() UserRepository](mockRepo)
            return repo(), nil
        }
        repo, _ := godi.Inject[func() UserRepository](mysqlRepo)
        return repo(), nil
    }),
)
```

## Benefits

| Benefit | Description |
|---------|-------------|
| **Separation of Concerns** | Configuration logic isolated in Build functions |
| **Lazy Loading** | Containers only added when needed |
| **Type Safety** | Compile-time checking for all dependencies |
| **Testability** | Easy to swap implementations for testing |
| **Flexibility** | Runtime decisions without code changes |
| **No Frozen Error** | Build function can add containers safely |

## Best Practices

1. **Pre-register containers**: Create containers for each variant before the Build function
2. **Use interfaces**: Depend on abstractions, not concrete types
3. **Document decisions**: Comment why certain containers are conditionally added
4. **Combine with hooks**: Use hooks for cleanup of runtime-added resources
5. **Keep it simple**: Don't overuse - consider if a simpler pattern would work

## Understanding Frozen vs Runtime Add

### Frozen Container (Cannot Add)

```go
child := &godi.Container{}
child.MustAdd(godi.Provide(Config{Value: "child"}))

parent := &godi.Container{}
parent.MustAdd(child)  // child is now frozen

// This will FAIL: "container is frozen"
err := child.Add(godi.Provide(struct{ Name string }{Name: "new"}))
```

### Runtime Add in Build (Allowed)

```go
c := &godi.Container{}

// Pre-create containers with different values
containerA := &godi.Container{}
containerA.MustAdd(godi.Provide(Database{DSN: "mysql://db-a"}))

containerB := &godi.Container{}
containerB.MustAdd(godi.Provide(Database{DSN: "mysql://db-b"}))

type Config struct {
    UseA bool
}

c.MustAdd(
    godi.Provide(Config{UseA: true}),
    godi.Build(func(cfg Config) (Database, error) {
        // Choose container at runtime based on config
        if cfg.UseA {
            db, _ := godi.Inject[Database](containerA)
            return db, nil
        }
        db, _ := godi.Inject[Database](containerB)
        return db, nil
    }),
)
```

### Why This Works

- **Build functions execute lazily** - containers are added at injection time, not registration time
- **The Build function receives the container** - it has access to add new providers
- **Frozen check happens at Add time** - Build execution is a special case

## Best Practices

1. **Pre-register containers**: Create containers for each variant before the Build function
2. **Use interfaces**: Depend on abstractions, not concrete types
3. **Document decisions**: Comment why certain containers are conditionally added
4. **Combine with hooks**: Use hooks for cleanup of runtime-added resources
5. **Keep it simple**: Don't overuse - consider if a simpler pattern would work

## Related Examples

- [08-nested-container-hooks](../08-nested-container-hooks/) - Multi-level container hooks
- [07-lifecycle-cleanup](../07-lifecycle-cleanup/) - Resource cleanup with hooks
- [06-web-app](../06-web-app/) - Production web app architecture
