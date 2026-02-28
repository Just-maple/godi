package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Error Handling Example: Using Inject and ShouldAdd
// Demonstrates graceful error handling without panics

type Database struct {
	DSN string
}

type Config struct {
	Port int
}

func main() {
	c := &godi.Container{}

	// Use ShouldAdd to handle duplicate registration errors
	err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}
	fmt.Println("First registration successful")

	// Duplicate registration returns error
	err = c.ShouldAdd(godi.Provide(Database{DSN: "mysql://remote"}))
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}

	// Use Inject to handle injection errors
	db, err := godi.Inject[Database](c)
	if err != nil {
		fmt.Printf("Injection failed: %v\n", err)
		return
	}
	fmt.Printf("Database: %s\n", db.DSN)

	// Injecting non-existent dependency returns error
	_, err = godi.Inject[Config](c)
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
	}
}
