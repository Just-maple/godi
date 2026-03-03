// Package service provides business logic layer
package service

import (
	"fmt"

	"github.com/Just-maple/godi/examples/09-web-app/internal/model"
	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// UserServiceInterface defines the contract for user service
type UserServiceInterface interface {
	GetUser(id int) (*model.User, error)
	CreateUser(name, email string) (*model.User, error)
}

// UserService handles user business logic
// Depends on interfaces (abstractions), not concrete implementations
type UserService struct {
	repo  UserRepositoryInterface
	cache interfaces.Cache
}

// UserRepositoryInterface is the interface for user repository
type UserRepositoryInterface interface {
	GetByID(id int) (*model.User, error)
	Save(user *model.User) error
	HealthCheck() bool
}

// NewUserService creates a new user service
// Injects interfaces instead of concrete types
func NewUserService(repo UserRepositoryInterface, cache interfaces.Cache) *UserService {
	return &UserService{
		repo:  repo,
		cache: cache,
	}
}

// GetUser retrieves a user by ID with caching
func (s *UserService) GetUser(id int) (*model.User, error) {
	key := fmt.Sprintf("user:%d", id)

	// Try cache first
	if cached, err := s.cache.Get(key); err == nil && cached != nil {
		return cached.(*model.User), nil
	}

	// Fallback to database
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Populate cache
	s.cache.Set(key, user, 3600)

	return user, nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(name, email string) (*model.User, error) {
	user := &model.User{Name: name, Email: email}
	if err := s.repo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}
