package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Error Handling Example: Add/Inject vs MustAdd/MustInject
// Demonstrates both graceful error handling and panic-on-error patterns

type Database struct {
	DSN string
}

type Config struct {
	Port int
}

type CriticalConfig struct {
	SecretKey string
}

func main() {
	fmt.Println("=== Error Handling Patterns ===")
	fmt.Println()

	// Pattern 1: Graceful error handling with Add/Inject
	fmt.Println("--- Pattern 1: Add/Inject (Graceful) ---")
	gracefulErrorHandling()

	fmt.Println()

	// Pattern 2: Panic-on-error with MustAdd/MustInject
	fmt.Println("--- Pattern 2: MustAdd/MustInject (Panic) ---")
	panicOnError()
}

func gracefulErrorHandling() {
	c := &godi.Container{}

	// Use Add to handle duplicate registration errors
	err := c.Add(
		godi.Provide(Database{DSN: "mysql://localhost"}),
		godi.Provide(Config{Port: 8080}),
	)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		return
	}
	fmt.Println("✓ Registration successful")

	// Duplicate registration returns error
	err = c.Add(godi.Provide(Database{DSN: "mysql://remote"}))
	if err != nil {
		fmt.Printf("✓ Expected duplicate error: %v\n", err)
	}

	// Use Inject to handle injection errors
	db, err := godi.Inject[Database](c)
	if err != nil {
		fmt.Printf("Injection failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Database: %s\n", db.DSN)

	// Injecting non-existent dependency returns error
	_, err = godi.Inject[CriticalConfig](c)
	if err != nil {
		fmt.Printf("✓ Expected not-found error: %v\n", err)
	}
}

func panicOnError() {
	c := &godi.Container{}

	// Use MustAdd with method chaining - panics if duplicate
	c.MustAdd(
		godi.Provide(CriticalConfig{SecretKey: "super-secret-key"}),
		godi.Provide(Database{DSN: "mysql://localhost"}),
	)
	fmt.Println("✓ MustAdd successful (panics on duplicate)")

	// Use MustInject - panics if dependency not found
	config := godi.MustInject[CriticalConfig](c)
	db := godi.MustInject[Database](c)

	fmt.Printf("✓ Secret Key: %s\n", config.SecretKey)
	fmt.Printf("✓ Database: %s\n", db.DSN)

	// Use MustInjectTo to inject directly into variable
	var extraDB Database
	godi.MustInjectTo(c, &extraDB)
	fmt.Printf("✓ Extra Database (MustInjectTo): %s\n", extraDB.DSN)

	// Method chaining example
	c2 := &godi.Container{}
	c2.MustAdd(
		godi.Provide(Database{DSN: "chain-db"}),
	).MustAdd(
		godi.Provide(Config{Port: 3306}),
	)
	fmt.Println("✓ Method chaining with MustAdd")
}

// When to use each pattern:
//
// Add/Inject (Graceful):
//   - Optional dependencies
//   - User-provided configurations
//   - Test scenarios
//   - When you want to handle errors explicitly
//
// MustAdd/MustInject (Panic):
//   - Critical dependencies that must exist
//   - Application startup (fail fast)
//   - Production code where missing deps are bugs
//   - Cleaner code when errors are unexpected
