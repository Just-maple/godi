package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// Multi-Container Injection Example: Demonstrates cross-container dependency injection

type Database struct {
	DSN string
}

type Cache struct {
	Host string
	Port int
}

type Config struct {
	AppName string
}

func main() {
	// Create multiple containers
	dbContainer := &godi.Container{}
	cacheContainer := &godi.Container{}
	configContainer := &godi.Container{}

	// Register different dependencies in different containers
	dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}))
	cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost", Port: 6379}))
	configContainer.Add(godi.Provide(Config{AppName: "multi-container-app"}))

	// Inject from single container
	db, err := godi.Inject[Database](dbContainer)
	if err != nil {
		panic("failed to inject Database")
	}
	fmt.Printf("Database: %s\n", db.DSN)

	// Inject from multiple containers
	// Inject searches through all provided containers in order
	cache, err := godi.Inject[Cache](dbContainer, cacheContainer)
	if err != nil {
		panic("failed to inject Cache")
	}
	fmt.Printf("Cache: %s:%d\n", cache.Host, cache.Port)

	// Inject from three containers
	cfg, err := godi.Inject[Config](dbContainer, cacheContainer, configContainer)
	if err != nil {
		panic("failed to inject Config")
	}
	fmt.Printf("Application: %s\n", cfg.AppName)

	// Use InjectTo across containers
	var extraDB Database
	err = godi.InjectTo(&extraDB, dbContainer, cacheContainer)
	fmt.Printf("Extra Inject Success: %v\n", err == nil)

	// Use Inject across containers
	extraCache, err := godi.Inject[Cache](cacheContainer, configContainer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Inject Cache: %s:%d\n", extraCache.Host, extraCache.Port)

	fmt.Println("\nMulti-container injection demo complete!")
}
