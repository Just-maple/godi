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

type ProductService struct {
	Repo *InMemoryRepository[Product]
}

func main() {
	c := &godi.Container{}

	// Register generic repositories
	userRepo := NewInMemoryRepository[User]()
	productRepo := NewInMemoryRepository[Product]()

	c.MustAdd(
		godi.Provide(userRepo),
		godi.Provide(productRepo),
		godi.Provide(&UserService{Repo: userRepo}),
		godi.Provide(&ProductService{Repo: productRepo}),
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

	// Save and retrieve data
	userRepo.Save(1, User{ID: 1, Name: "Alice"})
	userRepo.Save(2, User{ID: 2, Name: "Bob"})

	productRepo.Save(1, Product{ID: 1, Name: "Laptop", Price: 999.99})
	productRepo.Save(2, Product{ID: 2, Name: "Mouse", Price: 29.99})

	fmt.Println("\nStored Data:")
	fmt.Printf("User 1: %+v\n", userRepo.Get(1))
	fmt.Printf("User 2: %+v\n", userRepo.Get(2))
	fmt.Printf("Product 1: %+v\n", productRepo.Get(1))
	fmt.Printf("Product 2: %+v\n", productRepo.Get(2))
}
