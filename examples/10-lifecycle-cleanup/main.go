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
	// Simulate graceful shutdown
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

	// Execute hooks in reverse order (LIFO)
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

	// Simulate application running
	<-ctx.Done()
	return ctx.Err()
}

func main() {
	fmt.Println("=== Lifecycle Cleanup Example ===")
	fmt.Println()

	// Create container
	c := &godi.Container{}

	// Register lifecycle manager
	lifecycle := NewLifecycle()
	c.Add(godi.Provide(lifecycle))

	// Register Database with cleanup hook
	c.Add(godi.Lazy(func() (*Database, error) {
		db := &Database{name: "main-db"}
		fmt.Printf("[Database] %s connected\n", db.name)

		// Register cleanup hook
		lifecycle.AddShutdownHook(func(ctx context.Context) error {
			return db.Close()
		})

		return db, nil
	}))

	// Register Cache with cleanup hook
	c.Add(godi.Lazy(func() (*Cache, error) {
		cache := &Cache{name: "redis-cache"}
		fmt.Printf("[Cache] %s connected\n", cache.name)

		// Register cleanup hook
		lifecycle.AddShutdownHook(func(ctx context.Context) error {
			return cache.Close()
		})

		return cache, nil
	}))

	// Register Service with shutdown hook
	c.Add(godi.Lazy(func() (*Service, error) {
		service := &Service{name: "user-service"}
		fmt.Printf("[Service] %s initialized\n", service.name)

		// Register graceful shutdown hook
		lifecycle.AddShutdownHook(func(ctx context.Context) error {
			return service.Shutdown(ctx)
		})

		return service, nil
	}))

	// Register App
	c.Add(godi.Lazy(func() (*App, error) {
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

		lifecycle, err := godi.Inject[*Lifecycle](c)
		if err != nil {
			return nil, err
		}

		return NewApp(db, cache, service, lifecycle), nil
	}))

	// Run application
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	app, err := godi.Inject[*App](c)
	if err != nil {
		panic(err)
	}

	// Run app (will timeout after 2 seconds)
	_ = app.Run(ctx)

	// Perform cleanup
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := lifecycle.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Shutdown error: %v\n", err)
	}

	fmt.Println("\n=== Demo Complete ===")
}
