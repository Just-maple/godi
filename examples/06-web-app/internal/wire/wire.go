// Package wire provides dependency injection setup
package wire

import (
	"context"
	"fmt"
	"time"

	"github.com/Just-maple/godi"

	"github.com/Just-maple/godi/examples/09-web-app/internal/app"
	"github.com/Just-maple/godi/examples/09-web-app/internal/config"
	"github.com/Just-maple/godi/examples/09-web-app/internal/handler"
	"github.com/Just-maple/godi/examples/09-web-app/internal/infrastructure"
	"github.com/Just-maple/godi/examples/09-web-app/internal/middleware"
	"github.com/Just-maple/godi/examples/09-web-app/internal/repository"
	"github.com/Just-maple/godi/examples/09-web-app/internal/service"
	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// NewAppContainer creates and configures the DI container
// Following Dependency Inversion Principle: depend on abstractions, not concretions
func NewAppContainer() *godi.Container {
	c := &godi.Container{}

	// Register shutdown hook using HookOnce for automatic single execution
	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			// Execute cleanup for closable resources using interface assertion
			if closer, ok := v.(interface{ Close() error }); ok {
				if err := closer.Close(); err != nil {
					fmt.Printf("[Cleanup] Error closing %T: %v\n", v, err)
				}
			}
		}
	})

	// Register Config (concrete type - no interface needed for config)
	// Register Infrastructure (Build with cleanup hooks)
	// Note: We register concrete types but depend on interfaces in upper layers
	c.MustAdd(
		godi.Provide(config.NewConfig()),
		godi.Build(func(cfg *config.Config) (interfaces.Database, error) {
			return infrastructure.NewDBConnection(cfg.DatabaseDSN), nil
		}),
		godi.Build(func(cfg *config.Config) (interfaces.Cache, error) {
			return infrastructure.NewCacheClient(cfg.CacheAddr), nil
		}),
		godi.Build(func(db interfaces.Database) (repository.UserRepositoryInterface, error) {
			return repository.NewUserRepository(db), nil
		}),
		godi.Build(func(repo repository.UserRepositoryInterface) (service.UserServiceInterface, error) {
			cache, err := godi.Inject[interfaces.Cache](c)
			if err != nil {
				return nil, err
			}
			return service.NewUserService(repo, cache), nil
		}),
		godi.Provide(handler.NewRouter()),
		godi.Build(func(svc service.UserServiceInterface) (interfaces.Handler, error) {
			router, err := godi.Inject[*handler.Router](c)
			if err != nil {
				return nil, err
			}
			return handler.NewUserHandler(svc, router), nil
		}),
		godi.Build(func(cfg *config.Config) (interfaces.Middleware, error) {
			return middleware.NewLoggingMiddleware(cfg.Debug), nil
		}),
		godi.Build(func(cfg *config.Config) (*app.App, error) {
			router := app.NewRouter()
			h, err := godi.Inject[interfaces.Handler](c)
			if err != nil {
				return nil, err
			}
			mw, err := godi.Inject[interfaces.Middleware](c)
			if err != nil {
				return nil, err
			}
			return app.NewApp(cfg, router, h, mw), nil
		}),
	)

	// Store shutdown executor in container for later retrieval
	c.MustAdd(godi.Provide(shutdown))

	return c
}

// Run starts the application and handles graceful shutdown
func Run() error {
	container := NewAppContainer()
	fmt.Println("✓ Container created")
	fmt.Println("✓ Using Dependency Inversion Principle")
	fmt.Println("✓ Shutdown hooks registered via HookOnce")

	appInstance, err := godi.Inject[*app.App](container)
	if err != nil {
		return fmt.Errorf("failed to inject App: %w", err)
	}

	shutdown, err := godi.Inject[godi.Hooks](container)
	if err != nil {
		return fmt.Errorf("failed to inject shutdown: %w", err)
	}

	fmt.Println("✓ All dependencies injected")
	fmt.Println()

	// Start the application
	if err := appInstance.Start(); err != nil {
		return err
	}

	// Perform graceful shutdown using hooks
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	fmt.Println("\n=== Starting Graceful Shutdown ===")
	shutdown.Iterate(shutdownCtx, true) // true = reverse order (LIFO)
	fmt.Println("=== Shutdown Complete ===")

	return nil
}
