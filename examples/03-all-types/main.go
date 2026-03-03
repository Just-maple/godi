package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// All Types & Generics Example: Demonstrates all supported types including generics

func main() {
	fmt.Println("=== All Types & Generics Example ===")
	fmt.Println()

	// Section 1: Basic Types
	fmt.Println("--- Basic Types ---")
	basicTypes()

	fmt.Println()

	// Section 2: Generics & Interfaces
	fmt.Println("--- Generics & Interfaces ---")
	genericsAndInterfaces()
}

func basicTypes() {
	c := &godi.Container{}

	// Basic types
	c.MustAdd(
		godi.Provide("application-name"),
		godi.Provide(42),
		godi.Provide(int8(8)),
		godi.Provide(int16(16)),
		godi.Provide(int32(32)),
		godi.Provide(int64(64)),
		godi.Provide(uint(100)),
		godi.Provide(uint8(8)),
		godi.Provide(uint16(16)),
		godi.Provide(uint32(32)),
		godi.Provide(uint64(64)),
		godi.Provide(float32(3.14)),
		godi.Provide(float64(3.14159)),
		godi.Provide(true),
	)

	// Slices
	c.MustAdd(
		godi.Provide([]string{"a", "b", "c"}),
		godi.Provide([]int{1, 2, 3}),
		godi.Provide([]byte{0x01, 0x02, 0x03}),
	)

	// Maps
	c.MustAdd(
		godi.Provide(map[string]int{"a": 1, "b": 2}),
		godi.Provide(map[string]string{"key": "value"}),
	)

	// Arrays
	c.MustAdd(
		godi.Provide([3]int{1, 2, 3}),
		godi.Provide([2]string{"x", "y"}),
	)

	// Pointers
	type User struct {
		Name string
	}
	c.MustAdd(godi.Provide(&User{Name: "Alice"}))

	// Structs
	type Config struct {
		Host string
		Port int
	}
	c.MustAdd(godi.Provide(Config{Host: "localhost", Port: 8080}))

	// Channels
	c.MustAdd(
		godi.Provide(make(chan int, 10)),
		godi.Provide(make(chan string, 5)),
	)

	// Functions
	c.MustAdd(
		godi.Provide(func() string { return "hello" }),
		godi.Provide(func(x int) int { return x * 2 }),
	)

	// Interfaces
	c.MustAdd(godi.Provide(any("interface value")))

	// Inject and verify
	str := godi.MustInject[string](c)
	num := godi.MustInject[int](c)
	f32 := godi.MustInject[float32](c)
	f64 := godi.MustInject[float64](c)
	boolean := godi.MustInject[bool](c)
	strSlice := godi.MustInject[[]string](c)
	intMap := godi.MustInject[map[string]int](c)
	arr := godi.MustInject[[3]int](c)
	user := godi.MustInject[*User](c)
	config := godi.MustInject[Config](c)
	fn := godi.MustInject[func() string](c)

	fmt.Printf("String: %s\n", str)
	fmt.Printf("Number: %d\n", num)
	fmt.Printf("Float32: %.2f\n", f32)
	fmt.Printf("Float64: %.5f\n", f64)
	fmt.Printf("Boolean: %v\n", boolean)
	fmt.Printf("String Slice: %v\n", strSlice)
	fmt.Printf("Int Map: %v\n", intMap)
	fmt.Printf("Array: %v\n", arr)
	fmt.Printf("User Pointer: %v\n", user)
	fmt.Printf("Config: %+v\n", config)
	fmt.Printf("Function Call: %s\n", fn())
}

// Generics and Interfaces Section

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

func genericsAndInterfaces() {
	c := &godi.Container{}

	// Register generic repositories and services using dependency injection
	c.MustAdd(
		godi.Provide(NewInMemoryRepository[User]()),
		godi.Provide(NewInMemoryRepository[Product]()),
		godi.Build(func(repo *InMemoryRepository[User]) (*UserService, error) {
			return NewUserService(repo), nil
		}),
		godi.Build(func(repo *InMemoryRepository[Product]) (*ProductService, error) {
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

// Supported Types Summary:
//
// ✅ Structs: Database, Config
// ✅ Primitives: string, int, bool, float64
// ✅ Pointers: *Database
// ✅ Slices: []string
// ✅ Maps: map[string]int
// ✅ Interfaces: any, custom interfaces
// ✅ Arrays: [3]int
// ✅ Channels: chan int
// ✅ Functions: func() error
// ✅ Generics: Repository[T], Service[T]
