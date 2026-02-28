package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// 泛型和接口示例

type User struct {
	ID   int
	Name string
}

type Product struct {
	ID    int
	Name  string
	Price float64
}

// 泛型仓库实现
type InMemoryRepository[T any] struct {
	data map[int]T
}

func NewInMemoryRepository[T any]() *InMemoryRepository[T] {
	return &InMemoryRepository[T]{data: make(map[int]T)}
}

func (r *InMemoryRepository[T]) Get(id int) T {
	return r.data[id]
}

func (r *InMemoryRepository[T]) Save(entity T) {
	r.data[id] = entity
}

// 服务结构
type UserService struct {
	Repo *InMemoryRepository[User]
}

type ProductService struct {
	Repo *InMemoryRepository[Product]
}

func main() {
	c := &godi.Container{}

	// 注册泛型仓库
	userRepo := NewInMemoryRepository[User]()
	productRepo := NewInMemoryRepository[Product]()

	c.Add(godi.Provide(userRepo))
	c.Add(godi.Provide(productRepo))

	// 注册服务
	c.Add(godi.Provide(&UserService{Repo: userRepo}))
	c.Add(godi.Provide(&ProductService{Repo: productRepo}))

	// 注入并使用
	userSvc, ok := godi.Inject[*UserService](c)
	if !ok {
		panic("failed to inject UserService")
	}

	productSvc, ok := godi.Inject[*ProductService](c)
	if !ok {
		panic("failed to inject ProductService")
	}

	fmt.Printf("用户服务就绪：%v\n", userSvc != nil)
	fmt.Printf("产品服务就绪：%v\n", productSvc != nil)

	// 保存和获取数据
	userRepo.Save(User{ID: 1, Name: "Alice"})
	userRepo.Save(User{ID: 2, Name: "Bob"})

	productRepo.Save(Product{ID: 1, Name: "Laptop", Price: 999.99})
	productRepo.Save(Product{ID: 2, Name: "Mouse", Price: 29.99})

	fmt.Println("\n存储的数据:")
	fmt.Printf("用户 1: %+v\n", userRepo.Get(1))
	fmt.Printf("用户 2: %+v\n", userRepo.Get(2))
	fmt.Printf("产品 1: %+v\n", productRepo.Get(1))
	fmt.Printf("产品 2: %+v\n", productRepo.Get(2))
}
