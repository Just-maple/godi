package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Basic Example: Registering and injecting dependencies
// Demonstrates Provide, Build patterns, and struct field injection

type Database struct {
	DSN string
}

type Config struct {
	AppName string
}

type Cache struct {
	Addr string
}

type Service struct {
	DB     Database
	Config Config
	Cache  Cache
}

func main() {
	fmt.Println("=== Basic Example ===")
	fmt.Println()

	// Create container
	c := &godi.Container{}

	// Pattern 1: Provide - Register simple values
	c.MustAdd(
		godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}),
		godi.Provide(Config{AppName: "my-app"}),
		godi.Provide(Cache{Addr: "redis://localhost:6379"}),
	)

	// Inject dependencies
	db, _ := godi.Inject[Database](c)
	cfg, _ := godi.Inject[Config](c)
	cache, _ := godi.Inject[Cache](c)

	fmt.Printf("Pattern 1 (Provide): %s for %s, cache: %s\n", db.DSN, cfg.AppName, cache.Addr)

	// Pattern 2: Build with single dependency (auto-injected)
	fmt.Println()
	fmt.Println("=== Build Patterns ===")
	fmt.Println()

	c2 := &godi.Container{}
	type DBConfig struct {
		DSN string
	}
	c2.MustAdd(
		godi.Provide(DBConfig{DSN: "postgres://localhost"}),
		godi.Build(func(cfg DBConfig) (Database, error) {
			return Database{DSN: cfg.DSN}, nil
		}),
	)

	db2, _ := godi.Inject[Database](c2)
	fmt.Printf("Pattern 2a (Build single dep): %s\n", db2.DSN)

	// Pattern 2b: Build with Container access (multiple dependencies)
	c3 := &godi.Container{}
	c3.MustAdd(
		godi.Provide(Database{DSN: "mysql://localhost"}),
		godi.Provide(Config{AppName: "multi-dep"}),
		godi.Build(func(c *godi.Container) (Service, error) {
			db, _ := godi.Inject[Database](c)
			cfg, _ := godi.Inject[Config](c)
			return Service{DB: db, Config: cfg}, nil
		}),
	)

	svc, _ := godi.Inject[Service](c3)
	fmt.Printf("Pattern 2b (Build multi-dep): %s for %s\n", svc.DB.DSN, svc.Config.AppName)

	// Pattern 2c: Build with no dependency (struct{})
	c4 := &godi.Container{}
	c4.MustAdd(
		godi.Build(func(_ struct{}) (string, error) {
			return "no-dependency", nil
		}),
	)

	result, _ := godi.Inject[string](c4)
	fmt.Printf("Pattern 2c (Build no dep): %s\n", result)

	// Pattern 3: Struct Field Injection
	fmt.Println()
	fmt.Println("=== Struct Field Injection ===")
	fmt.Println()

	c5 := &godi.Container{}
	c5.MustAdd(
		godi.Provide(Database{DSN: "field-inject-db"}),
		godi.Provide(Config{AppName: "field-inject-app"}),
		godi.Provide(Cache{Addr: "field-inject-cache"}),
	)

	service := &Service{}
	err := c5.Inject(&service.DB, &service.Config, &service.Cache)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Pattern 3 (Field Inject): DB=%s, App=%s, Cache=%s\n",
		service.DB.DSN, service.Config.AppName, service.Cache.Addr)

	// Pattern 4: Dependency Chain
	fmt.Println()
	fmt.Println("=== Dependency Chain ===")
	fmt.Println()

	type Name string
	type Length int
	type Result string

	c6 := &godi.Container{}
	c6.MustAdd(
		godi.Provide(Name("hello")),
		godi.Build(func(s Name) (Length, error) {
			return Length(len(s)), nil
		}),
		godi.Build(func(n Length) (Result, error) {
			return Result(fmt.Sprintf("len:%d", n)), nil
		}),
	)

	chainResult, _ := godi.Inject[Result](c6)
	fmt.Printf("Pattern 4 (Chain): hello -> %s\n", chainResult)

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}
