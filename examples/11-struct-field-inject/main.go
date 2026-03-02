package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Struct Field Inject Example
// Demonstrates using Container.Inject to populate struct fields directly

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
	fmt.Println("=== Struct Field Inject Example ===")
	fmt.Println()

	c := &godi.Container{}
	c.MustAdd(
		godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}),
		godi.Provide(Config{AppName: "inject-app"}),
		godi.Provide(Cache{Addr: "redis://localhost:6379"}),
		godi.Build(func(c *godi.Container) (*Service, error) {
			service := &Service{}

			if err := c.Inject(&service.DB, &service.Config, &service.Cache); err != nil {
				return nil, err
			}

			return service, nil
		}),
	)

	svc, err := godi.Inject[*Service](c)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Service created successfully!\n")
	fmt.Printf("  Database: %s\n", svc.DB.DSN)
	fmt.Printf("  App: %s\n", svc.Config.AppName)
	fmt.Printf("  Cache: %s\n", svc.Cache.Addr)

	fmt.Println()
	fmt.Println("=== Compare: Traditional vs Struct Field Inject ===")
	fmt.Println()

	c2 := &godi.Container{}
	c2.MustAdd(
		godi.Provide(Database{DSN: "postgres://localhost"}),
		godi.Provide(Config{AppName: "traditional-app"}),
		godi.Provide(Cache{Addr: "redis://remote:6379"}),
	)

	fmt.Println("Traditional approach (multiple Inject calls):")
	service1 := &Service{}
	service1.DB, _ = godi.Inject[Database](c2)
	service1.Config, _ = godi.Inject[Config](c2)
	service1.Cache, _ = godi.Inject[Cache](c2)
	fmt.Printf("  Got: DB=%s, App=%s, Cache=%s\n", service1.DB.DSN, service1.Config.AppName, service1.Cache.Addr)

	fmt.Println()
	fmt.Println("Struct field inject approach (single call):")
	service2 := &Service{}
	_ = c2.Inject(&service2.DB, &service2.Config, &service2.Cache)
	fmt.Printf("  Got: DB=%s, App=%s, Cache=%s\n", service2.DB.DSN, service2.Config.AppName, service2.Cache.Addr)

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}
