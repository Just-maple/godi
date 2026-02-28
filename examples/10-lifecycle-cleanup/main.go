package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Just-maple/godi"
)

// Lifecycle Cleanup Example: Demonstrates resource cleanup on shutdown
// Shows how to register shutdown functions during initialization
// and execute them in reverse order when the application exits

// Database represents a database connection
type Database struct {
	name   string
	closed bool
}

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

func (c *Cache) Close() error {
	c.closed = true
	fmt.Printf("[Cache] %s connection closed\n", c.name)
	return nil
}

// Service represents a business service
type Service struct {
	name string
}

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

// Lifecycle manages the application lifecycle
type Lifecycle struct {
	hooks      []func(context.Context) error
	hooksMutex sync.Mutex
}

// NewLifecycle creates a new lifecycle manager
func NewLifecycle() *Lifecycle {
	return &Lifecycle{
		hooks: make([]func(context.Context) error, 0),
	}
}

// AddShutdownHook adds a shutdown hook to be called on cleanup
func (l *Lifecycle) AddShutdownHook(hook func(context.Context) error) {
	l.hooksMutex.Lock()
	defer l.hooksMutex.Unlock()
	l.hooks = append(l.hooks, hook)
}

// Shutdown executes all shutdown hooks in reverse order
func (l *Lifecycle) Shutdown(ctx context.Context) error {
	l.hooksMutex.Lock()
	defer l.hooksMutex.Unlock()

	fmt.Println("\n=== Starting Shutdown ===")

	for i := len(l.hooks) - 1; i >= 0; i-- {
		if err := l.hooks[i](ctx); err != nil {
			fmt.Printf("Shutdown hook %d error: %v\n", i, err)
		}
	}

	fmt.Println("=== Shutdown Complete ===")
	return nil
}

// App represents the main application
type App struct {
	db        *Database
	cache     *Cache
	service   *Service
	lifecycle *Lifecycle
}

func NewApp(db *Database, cache *Cache, service *Service, lifecycle *Lifecycle) *App {
	return &App{
		db:        db,
		cache:     cache,
		service:   service,
		lifecycle: lifecycle,
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
	lifecycle := NewLifecycle()

	c.MustAdd(
		godi.Provide(lifecycle),
		godi.Lazy(func() (*Database, error) {
			db := &Database{name: "main-db"}
			fmt.Printf("[Database] %s connected\n", db.name)
			lifecycle.AddShutdownHook(func(ctx context.Context) error {
				return db.Close()
			})
			return db, nil
		}),
		godi.Lazy(func() (*Cache, error) {
			cache := &Cache{name: "redis-cache"}
			fmt.Printf("[Cache] %s connected\n", cache.name)
			lifecycle.AddShutdownHook(func(ctx context.Context) error {
				return cache.Close()
			})
			return cache, nil
		}),
		godi.Lazy(func() (*Service, error) {
			service := &Service{name: "user-service"}
			fmt.Printf("[Service] %s initialized\n", service.name)
			lifecycle.AddShutdownHook(func(ctx context.Context) error {
				return service.Shutdown(ctx)
			})
			return service, nil
		}),
		godi.Lazy(func() (*App, error) {
			db, _ := godi.Inject[*Database](c)
			cache, _ := godi.Inject[*Cache](c)
			service, _ := godi.Inject[*Service](c)
			lifecycle, _ := godi.Inject[*Lifecycle](c)
			return NewApp(db, cache, service, lifecycle), nil
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app, _ := godi.Inject[*App](c)
	_ = app.Run(ctx)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = lifecycle.Shutdown(shutdownCtx)

	fmt.Println("\n=== Demo Complete ===")
}
