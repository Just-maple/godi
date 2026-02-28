// Package wire provides dependency injection setup
package wire

import (
	"fmt"

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

	// Register Config (concrete type - no interface needed for config)
	c.Add(godi.Provide(config.NewConfig()))

	// Register Infrastructure (Lazy loading)
	// Note: We register concrete types but depend on interfaces in upper layers
	c.Add(godi.Lazy(func() (interfaces.Database, error) {
		cfg, err := godi.ShouldInject[*config.Config](c)
		if err != nil {
			return nil, err
		}
		return infrastructure.NewDBConnection(cfg.DatabaseDSN), nil
	}))

	c.Add(godi.Lazy(func() (interfaces.Cache, error) {
		cfg, err := godi.ShouldInject[*config.Config](c)
		if err != nil {
			return nil, err
		}
		return infrastructure.NewCacheClient(cfg.CacheAddr), nil
	}))

	// Register Repository layer - depends on interfaces.Database
	c.Add(godi.Lazy(func() (repository.UserRepositoryInterface, error) {
		db, err := godi.ShouldInject[interfaces.Database](c)
		if err != nil {
			return nil, err
		}
		return repository.NewUserRepository(db), nil
	}))

	// Register Service layer - depends on interfaces
	c.Add(godi.Lazy(func() (service.UserServiceInterface, error) {
		repo, err := godi.ShouldInject[repository.UserRepositoryInterface](c)
		if err != nil {
			return nil, err
		}
		cache, err := godi.ShouldInject[interfaces.Cache](c)
		if err != nil {
			return nil, err
		}
		return service.NewUserService(repo, cache), nil
	}))

	// Register Handler layer - depends on interfaces
	c.Add(godi.Provide(handler.NewRouter()))

	c.Add(godi.Lazy(func() (interfaces.Handler, error) {
		svc, err := godi.ShouldInject[service.UserServiceInterface](c)
		if err != nil {
			return nil, err
		}
		router, err := godi.ShouldInject[*handler.Router](c)
		if err != nil {
			return nil, err
		}
		return handler.NewUserHandler(svc, router), nil
	}))

	// Register Middleware - depends on interfaces
	c.Add(godi.Lazy(func() (interfaces.Middleware, error) {
		cfg, err := godi.ShouldInject[*config.Config](c)
		if err != nil {
			return nil, err
		}
		return middleware.NewLoggingMiddleware(cfg.Debug), nil
	}))

	// Register App - all dependencies are interfaces
	c.Add(godi.Lazy(func() (*app.App, error) {
		cfg, err := godi.ShouldInject[*config.Config](c)
		if err != nil {
			return nil, err
		}
		router := app.NewRouter()
		h, err := godi.ShouldInject[interfaces.Handler](c)
		if err != nil {
			return nil, err
		}
		mw, err := godi.ShouldInject[interfaces.Middleware](c)
		if err != nil {
			return nil, err
		}
		return app.NewApp(cfg, router, h, mw), nil
	}))

	return c
}

// Run starts the application
func Run() error {
	container := NewAppContainer()
	fmt.Println("✓ Container created (Lazy loading)")
	fmt.Println("✓ Using Dependency Inversion Principle")

	appInstance, err := godi.ShouldInject[*app.App](container)
	if err != nil {
		return fmt.Errorf("failed to inject App: %w", err)
	}

	fmt.Println("✓ All dependencies injected")
	fmt.Println()

	return appInstance.Start()
}
