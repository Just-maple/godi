// Package interfaces defines all interfaces used across the application
package interfaces

import "context"

// Repository defines the base interface for data access
type Repository interface {
	HealthCheck() bool
}

// Service defines the base interface for business logic
type Service interface {
	Initialize(ctx context.Context) error
}

// Handler defines the interface for request handlers
type Handler interface {
	Handle(ctx context.Context) error
}

// Middleware defines the interface for request processing pipeline
type Middleware interface {
	Process(handler Handler) Handler
}

// Cache defines the interface for caching operations
type Cache interface {
	Get(key string) (interface{}, error)
	Set(key string, value interface{}, ttl int) error
	Delete(key string) error
}

// Database defines the interface for database operations
type Database interface {
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(query string, args ...interface{}) (int64, error)
	Close() error
}
