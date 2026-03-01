# Chain Example

Demonstrates dependency transformation with Chain and constructor-based injection.

## What This Example Shows

- Using `Chain` to transform dependencies
- Combining Chain with Lazy for complex assembly
- Creating dependency chains (Config → Database → Repository → Service)
- Multiple independent chains in one container

## Key Concepts

### Simple Chain

```go
type Name string
type Length int
type Result string

c := &godi.Container{}
c.MustAdd(
    godi.Provide(Name("hello")),
    godi.Chain(func(s Name) (Length, error) {
        return Length(len(s)), nil
    }),
    godi.Chain(func(n Length) (Result, error) {
        return Result(fmt.Sprintf("len%d", n)), nil
    }),
)

result := godi.MustInject[Result](c) // "len5"
```

### Real-World Chain with Constructor Injection

```go
// Constructors
func NewRepository(db *Database) *Repository {
    return &Repository{DB: db}
}

func NewService(repo *Repository, name string) *Service {
    return &Service{Repo: repo, Name: name}
}

// Register with Chain and Lazy
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost", AppName: "chain-app"}),
    godi.Chain(func(cfg Config) (*Database, error) {
        return &Database{ConnString: cfg.DSN, Connected: true}, nil
    }),
    godi.Lazy(func(c *godi.Container) (*Repository, error) {
        db, _ := godi.Inject[*Database](c)
        return NewRepository(db), nil
    }),
    godi.Lazy(func(c *godi.Container) (*Service, error) {
        repo, _ := godi.Inject[*Repository](c)
        return NewService(repo, "UserService"), nil
    }),
)

svc := godi.MustInject[*Service](c)
```

### Multiple Independent Chains

```go
type BaseInt int
type DoubledInt int
type BaseStr string
type SuffixedStr string

c.MustAdd(
    godi.Provide(BaseInt(10)),
    godi.Provide(BaseStr("prefix")),
    godi.Chain(func(n BaseInt) (DoubledInt, error) {
        return DoubledInt(n * 2), nil
    }),
    godi.Chain(func(s BaseStr) (SuffixedStr, error) {
        return SuffixedStr(s + "-suffix"), nil
    }),
)

num := godi.MustInject[DoubledInt](c)   // 20
str := godi.MustInject[SuffixedStr](c)  // "prefix-suffix"
```

## Running the Example

```bash
go run main.go
```

## Output

```
=== Chain Example ===

Chain 1: 'hello' -> length 5
Chain 2: 5 -> 'len5'
Result: len5

=== Real-world Chain ===

Creating Database from Config: mysql://localhost:3306/mydb
Creating Repository with Database: mysql://localhost:3306/mydb
Creating Service with Repository

Service: UserService
Database: mysql://localhost:3306/mydb (connected: true)

=== Multiple Independent Chains ===

Number chain: 10 -> 20
String chain: prefix -> prefix-suffix
```

## Chain vs Lazy

| Feature | Chain | Lazy |
|---------|-------|------|
| Purpose | Transform existing dependency | Create new dependency |
| Input | Single dependency type | Container reference |
| Use case | Simple transformations | Complex assembly logic |

## When to Use

- **Chain**: Transform config values, wrap dependencies, create derived types
- **Lazy**: Complex assembly, multiple dependencies, conditional logic
