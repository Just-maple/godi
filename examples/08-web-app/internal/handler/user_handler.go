// Package handler provides HTTP handlers
package handler

import (
	"context"
	"fmt"

	"github.com/Just-maple/godi/examples/09-web-app/internal/service"
)

// UserHandler handles user-related requests
// Depends on service.UserServiceInterface (abstraction)
type UserHandler struct {
	service service.UserServiceInterface
	router  *Router
}

// NewUserHandler creates a new user handler
// Injects service.UserServiceInterface (abstraction)
func NewUserHandler(service service.UserServiceInterface, router *Router) *UserHandler {
	return &UserHandler{
		service: service,
		router:  router,
	}
}

// Handle implements interfaces.Handler
func (h *UserHandler) Handle(ctx context.Context) error {
	user, err := h.service.GetUser(1)
	if err != nil {
		return err
	}
	fmt.Printf("Handler: Got user %s\n", user.Name)
	return nil
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
