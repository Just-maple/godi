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
	"github.com/Just-maple/godi/examples/09-web-app/internal/lifecycle"
	"github.com/Just-maple/godi/examples/09-web-app/internal/middleware"
	"github.com/Just-maple/godi/examples/09-web-app/internal/repository"
	"github.com/Just-maple/godi/examples/09-web-app/internal/service"
	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// NewAppContainer creates and configures the DI container
// Following Dependency Inversion Principle: depend on abstractions, not concretions
func NewAppContainer() *godi.Container {
	c := &godi.Container{}

	// Register lifecycle manager first (used by all other components)
	appLifecycle := lifecycle.New("WebApp")

	// Register Config (concrete type - no interface needed for config)
	// Register Infrastructure (Lazy loading with cleanup hooks)
	// Note: We register concrete types but depend on interfaces in upper layers
	c.MustAdd(
		godi.Provide(appLifecycle),
		godi.Provide(config.NewConfig()),
		godi.Lazy(func() (interfaces.Database, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			db := infrastructure.NewDBConnection(cfg.DatabaseDSN)

			// Register cleanup hook - will be called on shutdown
			appLifecycle.AddShutdownHook(func(ctx context.Context) error {
				if closer, ok := db.(interface{ Close() error }); ok {
					return closer.Close()
				}
				return nil
			})

			return db, nil
		}),
		godi.Lazy(func() (interfaces.Cache, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			cache := infrastructure.NewCacheClient(cfg.CacheAddr)

			// Register cleanup hook - will be called on shutdown
			appLifecycle.AddShutdownHook(func(ctx context.Context) error {
				if closer, ok := cache.(interface{ Close() error }); ok {
					return closer.Close()
				}
				return nil
			})

			return cache, nil
		}),
		godi.Lazy(func() (repository.UserRepositoryInterface, error) {
			db, err := godi.Inject[interfaces.Database](c)
			if err != nil {
				return nil, err
			}
			return repository.NewUserRepository(db), nil
		}),
		godi.Lazy(func() (service.UserServiceInterface, error) {
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
		godi.Lazy(func() (interfaces.Handler, error) {
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
		godi.Lazy(func() (interfaces.Middleware, error) {
			cfg, err := godi.Inject[*config.Config](c)
			if err != nil {
				return nil, err
			}
			return middleware.NewLoggingMiddleware(cfg.Debug), nil
		}),
		godi.Lazy(func() (*app.App, error) {
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
			lc, err := godi.Inject[*lifecycle.Lifecycle](c)
			if err != nil {
				return nil, err
			}
			return app.NewApp(cfg, router, h, mw, lc), nil
		}),
	)

	return c
}

// Run starts the application and handles graceful shutdown
func Run() error {
	container := NewAppContainer()
	fmt.Println("✓ Container created (Lazy loading)")
	fmt.Println("✓ Using Dependency Inversion Principle")
	fmt.Println("✓ Lifecycle hooks registered")

	appInstance, err := godi.Inject[*app.App](container)
	if err != nil {
		return fmt.Errorf("failed to inject App: %w", err)
	}

	fmt.Println("✓ All dependencies injected")
	fmt.Println()

	// Start the application
	if err := appInstance.Start(); err != nil {
		return err
	}

	// Perform graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return appInstance.Shutdown(shutdownCtx)
}
