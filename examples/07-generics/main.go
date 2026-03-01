package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Generics and Interfaces Example

type User struct {
	ID   int
	Name string
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

// Generic repository implementation
type InMemoryRepository[T any] struct {
	data map[int]T
}

func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
	return &InMemoryRepository[T]{data: make(map[int]T)}
}

func (r *InMemoryRepository[T]) Get(id int) T {
	return r.data[id]
}

func (r *InMemoryRepository[T]) Save(id int, entity T) {
	r.data[id] = entity
}

// Service structures
type UserService struct {
	Repo *InMemoryRepository[User]
}

func NewUserService(repo *InMemoryRepository[User]) *UserService {
	return &UserService{Repo: repo}
}

type ProductService struct {
	Repo *InMemoryRepository[Product]
}

func NewProductService(repo *InMemoryRepository[Product]) *ProductService {
	return &ProductService{Repo: repo}
}

func main() {
	c := &godi.Container{}

	// Register generic repositories and services using dependency injection
	c.MustAdd(
		godi.Provide(NewInMemoryRepository[User]()),
		godi.Provide(NewInMemoryRepository[Product]()),
		godi.Build(func(c *godi.Container) (*UserService, error) {
			repo, err := godi.Inject[*InMemoryRepository[User]](c)
			if err != nil {
				return nil, err
			}
			return NewUserService(repo), nil
		}),
		godi.Build(func(c *godi.Container) (*ProductService, error) {
			repo, err := godi.Inject[*InMemoryRepository[Product]](c)
			if err != nil {
				return nil, err
			}
			return NewProductService(repo), nil
		}),
	)

	// Inject and use
	userSvc, err := godi.Inject[*UserService](c)
	if err != nil {
		panic("failed to inject UserService")
	}

	productSvc, err := godi.Inject[*ProductService](c)
	if err != nil {
		panic("failed to inject ProductService")
	}

	fmt.Printf("User Service Ready: %v\n", userSvc != nil)
	fmt.Printf("Product Service Ready: %v\n", productSvc != nil)

	// Save and retrieve data through injected services
	userSvc.Repo.Save(1, User{ID: 1, Name: "Alice"})
	userSvc.Repo.Save(2, User{ID: 2, Name: "Bob"})

	productSvc.Repo.Save(1, Product{ID: 1, Name: "Laptop", Price: 999.99})
	productSvc.Repo.Save(2, Product{ID: 2, Name: "Mouse", Price: 29.99})

	fmt.Println("\nStored Data:")
	fmt.Printf("User 1: %+v\n", userSvc.Repo.Get(1))
	fmt.Printf("User 2: %+v\n", userSvc.Repo.Get(2))
	fmt.Printf("Product 1: %+v\n", productSvc.Repo.Get(1))
	fmt.Printf("Product 2: %+v\n", productSvc.Repo.Get(2))
}
