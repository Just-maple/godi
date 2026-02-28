// Package app provides the main application structure
package app

import (
	"context"
	"fmt"

	"github.com/Just-maple/godi/examples/09-web-app/internal/config"
	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// App represents the main application
// All dependencies are interfaces (abstractions)
type App struct {
	config     *config.Config
	router     *Router
	handler    interfaces.Handler
	middleware []interfaces.Middleware
}

// NewApp creates a new application instance
// Injects interfaces.Handler and interfaces.Middleware (abstractions)
func NewApp(
	cfg *config.Config,
	router *Router,
	handler interfaces.Handler,
	mw interfaces.Middleware,
) *App {
	return &App{
		config:     cfg,
		router:     router,
		handler:    mw.Process(handler),
		middleware: []interfaces.Middleware{mw},
	}
}

// Start starts the application
func (a *App) Start() error {
	fmt.Printf("Starting %s on port %d\n", a.config.AppName, a.config.Port)
	if a.config.Debug {
		fmt.Println("Debug mode: enabled")
	}

	ctx := context.Background()
	return a.handler.Handle(ctx)
}

// Router holds routing configuration
type Router struct {
	Routes []string
}

// NewRouter creates a new router
func NewRouter() *Router {
	return &Router{
		Routes: []string{"/users", "/posts", "/comments"},
	}
}
