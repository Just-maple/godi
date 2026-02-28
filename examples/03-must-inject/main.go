package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// Must Functions Example: Using MustInject and MustAdd
// Panics on error - use when dependencies are critical

type CriticalConfig struct {
	SecretKey string
}

type Database struct {
	DSN string
}

func main() {
	c := &godi.Container{}

	// Register critical dependencies
	c.MustAdd(godi.Provide(CriticalConfig{SecretKey: "super-secret-key"}))
	c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

	// Use MustInject - panics if dependency not found
	config := godi.MustInject[CriticalConfig](c)
	db := godi.MustInject[Database](c)

	fmt.Printf("Secret Key: %s\n", config.SecretKey)
	fmt.Printf("Database: %s\n", db.DSN)

	// Use MustInjectTo to inject directly into variable
	var extraDB Database
	godi.MustInjectTo(&extraDB, c)
	fmt.Printf("Extra Database: %s\n", extraDB.DSN)
}
