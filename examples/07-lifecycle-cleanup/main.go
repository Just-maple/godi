package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Just-maple/godi"
)

// Lifecycle & Hook Example: Demonstrates resource lifecycle management
// Shows Hook, HookOnce, startup/shutdown sequences, and graceful cleanup

// Database represents a database connection
type Database struct {
	name   string
	closed bool
}

func (d *Database) Close() error {
	d.closed = true
	fmt.Printf("  [Database] %s connection closed\n", d.name)
	return nil
}

// Cache represents a cache connection
type Cache struct {
	name   string
	closed bool
}

func (c *Cache) Close() error {
	c.closed = true
	fmt.Printf("  [Cache] %s connection closed\n", c.name)
	return nil
}

// Service represents a business service
type Service struct {
	name string
}

func (s *Service) Shutdown(ctx context.Context) error {
	fmt.Printf("  [Service] %s shutting down gracefully\n", s.name)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(50 * time.Millisecond):
		fmt.Printf("  [Service] %s shutdown complete\n", s.name)
		return nil
	}
}

// App represents the main application
type App struct {
	db      *Database
	cache   *Cache
	service *Service
}

func NewApp(db *Database, cache *Cache, service *Service) *App {
	return &App{db: db, cache: cache, service: service}
}

func main() {
	fmt.Println("=== Lifecycle & Hook Example ===")
	fmt.Println()

	c := &godi.Container{}

	// Register shutdown hook BEFORE injecting dependencies
	// HookOnce: automatically runs only once per type
	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			// Handle closable resources
			if closer, ok := v.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					fmt.Printf("  [Cleanup Error] %T: %v\n", v, err)
				}
				return
			}
			// Handle shutdownable resources
			if shutdowner, ok := v.(interface{ Shutdown(context.Context) error }); ok {
				if err := shutdowner.Shutdown(ctx); err != nil {
					fmt.Printf("  [Shutdown Error] %T: %v\n", v, err)
				}
			}
		}
	})

	// Hook with provided counter: manual control over execution
	startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil // Skip if already injected before
		}
		return func(ctx context.Context) {
			fmt.Printf("  [Startup] Initializing: %T\n", v)
		}
	})

	// Register dependencies
	c.MustAdd(
		godi.Provide(&Database{name: "main-db"}),
		godi.Provide(&Cache{name: "redis-cache"}),
		godi.Provide(&Service{name: "user-service"}),
		godi.Build(func(db *Database) (*App, error) {
			cache, _ := godi.Inject[*Cache](c)
			service, _ := godi.Inject[*Service](c)
			return NewApp(db, cache, service), nil
		}),
	)

	ctx := context.Background()

	// Inject dependencies (hooks are registered during injection)
	fmt.Println("--- Injecting Dependencies ---")
	app, _ := godi.Inject[*App](c)
	_, _ = godi.Inject[*Database](c)
	_, _ = godi.Inject[*Cache](c)
	_, _ = godi.Inject[*Service](c)

	fmt.Println()
	fmt.Println("--- Running Startup Hooks ---")
	startup.Iterate(ctx, false) // false = forward order

	fmt.Println()
	fmt.Println("--- Application Running ---")
	fmt.Printf("App: DB=%s, Cache=%s, Service=%s\n", app.db.name, app.cache.name, app.service.name)

	fmt.Println()
	fmt.Println("--- Running Shutdown Hooks (Reverse Order) ---")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	shutdown.Iterate(shutdownCtx, true) // true = reverse order (LIFO)

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}

// Hook Patterns:
//
// 1. HookOnce - Automatic single execution
//    shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
//        return func(ctx context.Context) { /* cleanup */ }
//    })
//
// 2. Hook with provided counter - Manual control
//    startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
//        if provided > 0 { return nil } // Skip after first
//        return func(ctx context.Context) { /* init */ }
//    })
//
// 3. Iterate options:
//    hooks.Iterate(ctx, false) // Forward order (FIFO)
//    hooks.Iterate(ctx, true)  // Reverse order (LIFO) - good for cleanup
