package main

import (
	"context"
	"fmt"

	"github.com/Just-maple/godi"
)

type Environment string

const (
	Production  Environment = "production"
	Development Environment = "development"
	Test        Environment = "test"
)

type Database struct {
	DSN string
}

type Cache struct {
	Addr string
}

type Config struct {
	Env         Environment
	AppName     string
	EnableCache bool
}

type UserRepository interface {
	GetUser(id int) string
}

type MySQLUserRepository struct {
	db Database
}

func (m *MySQLUserRepository) GetUser(id int) string {
	return fmt.Sprintf("MySQL user %d from %s", id, m.db.DSN)
}

type MockUserRepository struct{}

func (m *MockUserRepository) GetUser(id int) string {
	return fmt.Sprintf("Mock user %d", id)
}

func main() {
	fmt.Println("=== Runtime Container Add - Real World Example ===\n")

	fmt.Println("--- Scenario 1: Environment-Based Database Selection ---")
	demoEnvironmentBasedDB()

	fmt.Println("\n--- Scenario 2: Conditional Cache Registration ---")
	demoConditionalCache()

	fmt.Println("\n--- Scenario 3: Interface Implementation Selection ---")
	demoInterfaceSelection()

	fmt.Println("\n--- Scenario 4: Frozen vs Runtime Add ---")
	demoFrozenVsRuntimeAdd()
}

func demoEnvironmentBasedDB() {
	prodDB := &godi.Container{}
	prodDB.MustAdd(godi.Provide(Database{DSN: "mysql://prod-db.company.com:3306/app"}))

	devDB := &godi.Container{}
	devDB.MustAdd(godi.Provide(Database{DSN: "mysql://localhost:3306/dev"}))

	testDB := &godi.Container{}
	testDB.MustAdd(godi.Provide(Database{DSN: "mysql://localhost:3306/test"}))

	c := &godi.Container{}
	c.MustAdd(
		godi.Provide(Config{Env: Production, AppName: "MyApp"}),
		godi.Build(func(c *godi.Container) (string, error) {
			cfg, err := godi.Inject[Config](c)
			if err != nil {
				return "", err
			}

			fmt.Printf("Build: Detected environment = %s\n", cfg.Env)

			switch cfg.Env {
			case Production:
				fmt.Println("Build: Registering production database container")
				c.MustAdd(prodDB)
			case Development:
				fmt.Println("Build: Registering development database container")
				c.MustAdd(devDB)
			case Test:
				fmt.Println("Build: Registering test database container")
				c.MustAdd(testDB)
			}

			db, err := godi.Inject[Database](c)
			if err != nil {
				return "", err
			}

			fmt.Printf("Build: Connected to %s\n", db.DSN)
			return fmt.Sprintf("Connected to %s", db.DSN), nil
		}),
	)

	result, err := godi.Inject[string](c)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Result: %s\n", result)
}

func demoConditionalCache() {
	redisCache := &godi.Container{}
	redisCache.MustAdd(godi.Provide(Cache{Addr: "redis://prod-redis.company.com:6379"}))

	c := &godi.Container{}
	c.MustAdd(
		godi.Provide(Config{Env: Production, AppName: "MyApp", EnableCache: true}),
		godi.Build(func(c *godi.Container) (string, error) {
			cfg, err := godi.Inject[Config](c)
			if err != nil {
				return "", err
			}

			fmt.Printf("Build: Config.EnableCache = %v\n", cfg.EnableCache)

			if cfg.EnableCache {
				fmt.Println("Build: Registering Redis cache container")
				c.MustAdd(redisCache)

				cache, err := godi.Inject[Cache](c)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("Connected to cache: %s", cache.Addr), nil
			}

			return "Cache disabled, running without caching", nil
		}),
	)

	result, err := godi.Inject[string](c)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Result: %s\n", result)
}

func demoInterfaceSelection() {
	mockRepo := &godi.Container{}
	mockRepo.MustAdd(godi.Provide(func() UserRepository { return &MockUserRepository{} }()))

	mysqlRepo := &godi.Container{}
	mysqlRepo.MustAdd(
		godi.Provide(Database{DSN: "mysql://localhost:3306/app"}),
		godi.Build(func(c *godi.Container) (UserRepository, error) {
			db, err := godi.Inject[Database](c)
			if err != nil {
				return nil, err
			}
			return &MySQLUserRepository{db: db}, nil
		}),
	)

	c := &godi.Container{}
	c.MustAdd(
		godi.Provide(Config{Env: Test, AppName: "MyApp"}),
		godi.Build(func(c *godi.Container) (string, error) {
			cfg, err := godi.Inject[Config](c)
			if err != nil {
				return "", err
			}

			fmt.Printf("Build: Environment = %s\n", cfg.Env)

			if cfg.Env == Test {
				fmt.Println("Build: Using mock repository container")
				c.MustAdd(mockRepo)
			} else {
				fmt.Println("Build: Using MySQL repository container")
				c.MustAdd(mysqlRepo)
			}

			repo, err := godi.Inject[UserRepository](c)
			if err != nil {
				return "", err
			}

			fmt.Printf("Build: Repository type = %T\n", repo)
			return repo.GetUser(123), nil
		}),
	)

	result, err := godi.Inject[string](c)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Result: %s\n", result)

	fmt.Println("\n--- Cleanup Hooks Demo ---")
	demoCleanupHooks()
}

func demoCleanupHooks() {
	dbContainer := &godi.Container{}
	dbContainer.MustAdd(godi.Provide(Database{DSN: "mysql://localhost:3306/runtime"}))

	c := &godi.Container{}

	cleanupCalls := 0
	cleanup := c.HookOnce("cleanup", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				cleanupCalls++
				fmt.Printf("[Cleanup] Database connection closed (call #%d)\n", cleanupCalls)
			}
		}
	})

	c.MustAdd(
		godi.Provide(Config{Env: Production}),
		godi.Build(func(c *godi.Container) (string, error) {
			cfg, err := godi.Inject[Config](c)
			if err != nil {
				return "", err
			}

			if cfg.Env == Production {
				fmt.Println("Build: Adding production database container at runtime")
				c.MustAdd(dbContainer)
			}

			db, err := godi.Inject[Database](c)
			if err != nil {
				return "", err
			}
			fmt.Printf("Build: Connected to %s\n", db.DSN)
			return fmt.Sprintf("Connected to %s", db.DSN), nil
		}),
	)

	result, err := godi.Inject[string](c)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Result: %s\n", result)

	db, _ := godi.Inject[Database](c)
	fmt.Printf("Injected Database: {DSN: %s}\n", db.DSN)

	ctx := context.Background()
	fmt.Println("\nExecuting cleanup hooks...")
	cleanup.Iterate(ctx, true)

	fmt.Printf("Verification: Cleanup called %d time(s) (expected: 1)\n", cleanupCalls)
}

func demoFrozenVsRuntimeAdd() {
	fmt.Println("\n[1] Direct add to frozen container (will fail):")
	child := &godi.Container{}
	child.MustAdd(godi.Provide("child-value"))

	parent := &godi.Container{}
	parent.MustAdd(child)

	err := child.Add(godi.Provide("new-value"))
	if err != nil {
		fmt.Printf("✓ Expected error: %v\n", err)
	}

	fmt.Println("\n[2] Runtime add in Build function (will succeed):")
	nested := &godi.Container{}
	nested.MustAdd(godi.Provide("runtime-value"))

	c := &godi.Container{}
	c.MustAdd(
		godi.Provide(42),
		godi.Build(func(c *godi.Container) (string, error) {
			fmt.Println("Build: Adding nested container at runtime...")
			c.MustAdd(nested)

			i, _ := godi.Inject[int](c)
			v, _ := godi.Inject[string](c)
			return fmt.Sprintf("value=%s, num=%d", v, i), nil
		}),
	)

	result, err := godi.Inject[string](c)
	if err != nil {
		fmt.Printf("✗ Unexpected error: %v\n", err)
		return
	}
	fmt.Printf("✓ Success: %s\n", result)

	fmt.Println("\n[3] Create fresh container for another use case:")
	freshNested := &godi.Container{}
	freshNested.MustAdd(godi.Provide("fresh-value"))

	anotherParent := &godi.Container{}
	anotherParent.MustAdd(freshNested)

	v, _ := godi.Inject[string](anotherParent)
	fmt.Printf("✓ Fresh container works: %s\n", v)

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Frozen containers cannot accept new providers directly")
	fmt.Println("✓ Build functions CAN add containers at runtime")
	fmt.Println("✓ Each container should only be added to one parent")
}
