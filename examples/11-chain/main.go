package main

import (
	"fmt"

	"github.com/Just-maple/godi"
)

// Chain Example: Demonstrates dependency transformation with Chain
// Chain creates a new provider by transforming an existing dependency
// No need to pass Container - it's automatically available in Lazy!

type Config struct {
	DSN     string
	AppName string
}

type Database struct {
	ConnString string
	Connected  bool
}

type Repository struct {
	DB *Database
}

type Service struct {
	Repo *Repository
	Name string
}

func main() {
	fmt.Println("=== Chain Example ===")
	fmt.Println()

	// Simple Chain: Name -> Length -> Result
	type Name string
	type Length int
	type Result string

	c1 := &godi.Container{}
	c1.MustAdd(
		godi.Provide(Name("hello")),
		godi.Chain(func(s Name) (Length, error) {
			fmt.Printf("Chain 1: '%s' -> length %d\n", s, len(s))
			return Length(len(s)), nil
		}),
		godi.Chain(func(n Length) (Result, error) {
			result := Result(fmt.Sprintf("len%d", n))
			fmt.Printf("Chain 2: %d -> '%s'\n", n, result)
			return result, nil
		}),
	)

	result, _ := godi.Inject[Result](c1)
	fmt.Printf("Result: %s\n\n", result)

	// Real-world: Config -> Database -> Repository -> Service
	fmt.Println("=== Real-world Chain ===")
	fmt.Println()

	c2 := &godi.Container{}
	c2.MustAdd(
		godi.Provide(Config{
			DSN:     "mysql://localhost:3306/mydb",
			AppName: "chain-app",
		}),
		godi.Chain(func(cfg Config) (*Database, error) {
			fmt.Printf("Creating Database from Config: %s\n", cfg.DSN)
			return &Database{ConnString: cfg.DSN, Connected: true}, nil
		}),
		godi.Chain(func(db *Database) (*Repository, error) {
			fmt.Printf("Creating Repository with Database: %s\n", db.ConnString)
			return &Repository{DB: db}, nil
		}),
		godi.Chain(func(repo *Repository) (*Service, error) {
			fmt.Printf("Creating Service with Repository\n")
			return &Service{Repo: repo, Name: "UserService"}, nil
		}),
	)

	svc, _ := godi.Inject[*Service](c2)
	fmt.Printf("\nService: %s\n", svc.Name)
	fmt.Printf("Database: %s (connected: %v)\n", svc.Repo.DB.ConnString, svc.Repo.DB.Connected)

	fmt.Println()
	fmt.Println("=== Multiple Independent Chains ===")
	fmt.Println()

	// Multiple independent chains using type aliases
	type BaseInt int
	type DoubledInt int
	type BaseStr string
	type SuffixedStr string

	c3 := &godi.Container{}
	c3.MustAdd(
		godi.Provide(BaseInt(10)),
		godi.Provide(BaseStr("prefix")),
		godi.Chain(func(n BaseInt) (DoubledInt, error) {
			return DoubledInt(n * 2), nil
		}),
		godi.Chain(func(s BaseStr) (SuffixedStr, error) {
			return SuffixedStr(s + "-suffix"), nil
		}),
	)

	num, _ := godi.Inject[DoubledInt](c3)
	str, _ := godi.Inject[SuffixedStr](c3)
	fmt.Printf("Number chain: 10 -> %d\n", num)
	fmt.Printf("String chain: prefix -> %s\n", str)

	fmt.Println()
	fmt.Println("=== Demo Complete ===")
}
