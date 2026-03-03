// Package repository provides data access layer
package repository

import (
	"github.com/Just-maple/godi/examples/09-web-app/internal/model"
	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// UserRepositoryInterface defines the contract for user repository
type UserRepositoryInterface interface {
	GetByID(id int) (*model.User, error)
	Save(user *model.User) error
	HealthCheck() bool
}

// UserRepository handles user data access
// Depends on interfaces.Database (abstraction), not concrete implementation
type UserRepository struct {
	db interfaces.Database
}

// NewUserRepository creates a new user repository
// Injects interfaces.Database instead of concrete DBConnection
func NewUserRepository(db interfaces.Database) *UserRepository {
	return &UserRepository{db: db}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int) (*model.User, error) {
	// Simulate database query
	return &model.User{ID: id, Name: "User from DB", Email: "user@example.com"}, nil
}

// Save persists a user
func (r *UserRepository) Save(user *model.User) error {
	// Simulate database save
	return nil
}

// HealthCheck implements repository health check
func (r *UserRepository) HealthCheck() bool {
	return r.db != nil
}
