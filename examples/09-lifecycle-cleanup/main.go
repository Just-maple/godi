package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Just-maple/godi"
)

// Lifecycle Cleanup Example: Demonstrates resource cleanup using Hook system
// Shows how to register shutdown functions during initialization
// and execute them in reverse order when the application exits

// Database represents a database connection
type Database struct {
	name   string
	closed bool
}

// Close implements io.Closer interface
func (d *Database) Close() error {
	d.closed = true
	fmt.Printf("[Database] %s connection closed\n", d.name)
	return nil
}

// Cache represents a cache connection
type Cache struct {
	name   string
	closed bool
}

// Close implements io.Closer interface
func (c *Cache) Close() error {
	c.closed = true
	fmt.Printf("[Cache] %s connection closed\n", c.name)
	return nil
}

// Service represents a business service
type Service struct {
	name string
}

// Shutdown implements shutdown interface
func (s *Service) Shutdown(ctx context.Context) error {
	fmt.Printf("[Service] %s shutting down gracefully\n", s.name)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(100 * time.Millisecond):
		fmt.Printf("[Service] %s shutdown complete\n", s.name)
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
	return &App{
		db:      db,
		cache:   cache,
		service: service,
	}
}

// Run starts the application
func (a *App) Run(ctx context.Context) error {
	fmt.Println("Application is running...")
	fmt.Println("Press Ctrl+C or wait for timeout to shutdown")
	<-ctx.Done()
	return ctx.Err()
}

func main() {
	fmt.Println("=== Lifecycle Cleanup Example ===")
	fmt.Println()

	c := &godi.Container{}

	// Use HookOnce for automatic single execution cleanup
	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			// Handle closable resources
			if closer, ok := v.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					fmt.Printf("[Cleanup] Error closing %T: %v\n", v, err)
				}
				return
			}
			// Handle shutdownable resources
			if shutdowner, ok := v.(interface{ Shutdown(context.Context) error }); ok {
				if err := shutdowner.Shutdown(ctx); err != nil {
					fmt.Printf("[Cleanup] Error shutting down %T: %v\n", v, err)
				}
			}
		}
	})

	c.MustAdd(
		godi.Build(func(c *godi.Container) (*Database, error) {
			db := &Database{name: "main-db"}
			fmt.Printf("[Database] %s connected\n", db.name)
			return db, nil
		}),
		godi.Build(func(c *godi.Container) (*Cache, error) {
			cache := &Cache{name: "redis-cache"}
			fmt.Printf("[Cache] %s connected\n", cache.name)
			return cache, nil
		}),
		godi.Build(func(c *godi.Container) (*Service, error) {
			service := &Service{name: "user-service"}
			fmt.Printf("[Service] %s initialized\n", service.name)
			return service, nil
		}),
		godi.Build(func(c *godi.Container) (*App, error) {
			db, err := godi.Inject[*Database](c)
			if err != nil {
				return nil, err
			}
			cache, err := godi.Inject[*Cache](c)
			if err != nil {
				return nil, err
			}
			service, err := godi.Inject[*Service](c)
			if err != nil {
				return nil, err
			}
			return NewApp(db, cache, service), nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app, _ := godi.Inject[*App](c)
	_ = app.Run(ctx)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	fmt.Println("\n=== Starting Shutdown ===")
	shutdown.Iterate(shutdownCtx, true) // true = reverse order (LIFO)
	fmt.Println("=== Shutdown Complete ===")

	fmt.Println("\n=== Demo Complete ===")
}
