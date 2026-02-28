package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Basic Example: Registering and injecting simple dependencies
// Demonstrates the core functionality of godi

type Database struct {
	DSN string
}

type Config struct {
	AppName string
}

func main() {
	// Create container
	c := &godi.Container{}

	// Register multiple dependencies at once
	c.Add(
		godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}),
		godi.Provide(Config{AppName: "my-app"}),
	)

	// Inject dependencies
	db, err := godi.Inject[Database](c)
	if err != nil {
		panic("failed to inject Database")
	}

	cfg, err := godi.Inject[Config](c)
	if err != nil {
		panic("failed to inject Config")
	}

	fmt.Printf("Connected to %s for %s\n", db.DSN, cfg.AppName)
	// Output: Connected to mysql://localhost:3306/mydb for my-app
}
