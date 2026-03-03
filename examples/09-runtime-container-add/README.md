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

c.MustAdd(
    godi.Provide(Config{Env: Production}),
    godi.Build(func(c *godi.Container) (Database, error) {
        cfg, _ := godi.Inject[Config](c)
        
        switch cfg.Env {
        case Production:
            c.MustAdd(prodDB)
        case Development:
            c.MustAdd(devDB)
        }
        
        return godi.Inject[Database](c)
    }),
)
```

### Pattern 2: Conditional Feature Registration

```go
c.MustAdd(
    godi.Provide(Config{EnableCache: true}),
    godi.Build(func(c *godi.Container) (Service, error) {
        cfg, _ := godi.Inject[Config](c)
        
        if cfg.EnableCache {
            c.MustAdd(cacheContainer)
            cache, _ := godi.Inject[Cache](c)
            return NewServiceWithCache(cache)
        }
        
        return NewServiceWithoutCache()
    }),
)
```

### Pattern 3: Interface Implementation Selection

```go
mockRepo := &godi.Container{}
mockRepo.MustAdd(godi.Provide(func() UserRepository { return &MockUserRepository{} }()))

mysqlRepo := &godi.Container{}
mysqlRepo.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Build(func(db Database) (UserRepository, error) {
        return &MySQLUserRepository{db: db}, nil
    }),
)

c.MustAdd(
    godi.Provide(Config{Env: Test}),
    godi.Build(func(c *godi.Container) (UserRepository, error) {
        cfg, _ := godi.Inject[Config](c)
        
        if cfg.Env == Test {
            c.MustAdd(mockRepo)
        } else {
            c.MustAdd(mysqlRepo)
        }
        
        return godi.Inject[UserRepository](c)
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
nested := &godi.Container{}
nested.MustAdd(godi.Provide("value"))

c.MustAdd(
    godi.Build(func(c *godi.Container) (string, error) {
        // This is OK - adding during Build execution
        c.MustAdd(nested)
        return godi.Inject[string](c)
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
