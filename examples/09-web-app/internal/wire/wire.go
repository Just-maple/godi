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
		godi.Build(func(c *godi.Container) (interfaces.Database, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			db := infrastructure.NewDBConnection(cfg.DatabaseDSN)
			return db, nil
		}),
		godi.Build(func(c *godi.Container) (interfaces.Cache, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			cache := infrastructure.NewCacheClient(cfg.CacheAddr)
			return cache, nil
		}),
		godi.Build(func(c *godi.Container) (repository.UserRepositoryInterface, error) {
			db, err := godi.Inject[interfaces.Database](c)
			if err != nil {
				return nil, err
			}
			return repository.NewUserRepository(db), nil
		}),
		godi.Build(func(c *godi.Container) (service.UserServiceInterface, error) {
			repo, err := godi.Inject[repository.UserRepositoryInterface](c)
			if err != nil {
				return nil, err
			}
			cache, err := godi.Inject[interfaces.Cache](c)
			if err != nil {
				return nil, err
			}
			return service.NewUserService(repo, cache), nil
		}),
		godi.Provide(handler.NewRouter()),
		godi.Build(func(c *godi.Container) (interfaces.Handler, error) {
			svc, err := godi.Inject[service.UserServiceInterface](c)
			if err != nil {
				return nil, err
			}
			router, err := godi.Inject[*handler.Router](c)
			if err != nil {
				return nil, err
			}
			return handler.NewUserHandler(svc, router), nil
		}),
		godi.Build(func(c *godi.Container) (interfaces.Middleware, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			return middleware.NewLoggingMiddleware(cfg.Debug), nil
		}),
		godi.Build(func(c *godi.Container) (*app.App, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
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

	shutdown, err := godi.Inject[func(func([]func(context.Context)))](container)
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
	shutdown(func(hooks []func(context.Context)) {
		// Execute hooks in reverse order (LIFO)
		for i := len(hooks) - 1; i >= 0; i-- {
			hooks[i](shutdownCtx)
		}
	})
	fmt.Println("=== Shutdown Complete ===")

	return nil
}
