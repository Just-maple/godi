# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)

轻量级 Go 依赖注入框架，基于泛型实现，零反射，零代码生成。

## 特性

- ✅ **泛型支持** - 类型安全，无需反射
- ✅ **懒加载** - 依赖首次使用时初始化
- ✅ **循环依赖检测** - 运行时自动检测
- ✅ **多容器注入** - 支持跨容器查找
- ✅ **并发安全** - 所有操作线程安全
- ✅ **接口注入** - 支持依赖倒置原则

## 快速开始

```go
package main

import "github.com/Just-maple/godi"

type Config struct {
    DSN string
}

type Database struct {
    Conn string
}

func main() {
    c := &godi.Container{}
    
    // 注册依赖
    c.MustAdd(
        godi.Provide(Config{DSN: "mysql://localhost"}),
        godi.Build(func(c *godi.Container) (*Database, error) {
            cfg, _ := godi.Inject[Config](c)
            return &Database{Conn: cfg.DSN}, nil
        }),
    )
    
    // 注入依赖
    db := godi.MustInject[*Database](c)
    println(db.Conn)
}
```

## 核心 API

### 注册依赖

```go
c := &godi.Container{}

// Provide - 注册实例值
c.Add(godi.Provide(Config{Port: 8080}))

// Build - 注册工厂函数（懒加载，单例）
c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    return NewDatabase("dsn")
}))

// Chain - 从现有依赖派生新依赖
c.Add(godi.Chain(func(cfg Config) (*Connection, error) {
    return NewConnection(cfg.DSN)
}))
```

### 注入依赖

```go
// Inject - 返回 (值，错误)
db, err := godi.Inject[*Database](c)

// MustInject - 失败时 panic
db := godi.MustInject[*Database](c)

// InjectTo - 注入到现有变量
var db Database
err := godi.InjectTo(&db, c)

// MustInjectTo - 注入到现有变量，失败时 panic
godi.MustInjectTo(&db, c)
```

### 多容器注入

```go
db, err := godi.Inject[*Database](container1, container2, container3)
```

按顺序查找，返回第一个匹配的依赖。

## 使用场景

### 1. 基础注入

```go
c := &godi.Container{}
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    godi.Provide(Database{DSN: "mysql://localhost"}),
)

cfg, err := godi.Inject[Config](c)
```

### 2. 懒加载

工厂函数仅在首次请求时执行，结果缓存：

```go
c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    // 首次调用时执行
    return sql.Open("mysql", dsn)
}))

// 工厂函数在此执行
db, err := godi.Inject[*Database](c)
```

### 3. 依赖链

```go
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    
    godi.Build(func(c *godi.Container) (*Database, error) {
        cfg, _ := godi.Inject[Config](c)
        return NewDatabase(cfg.DSN)
    }),
    
    godi.Build(func(c *godi.Container) (*UserRepository, error) {
        db, _ := godi.Inject[*Database](c)
        return NewUserRepository(db)
    }),
    
    godi.Build(func(c *godi.Container) (*UserService, error) {
        repo, _ := godi.Inject[*UserRepository](c)
        return NewUserService(repo)
    }),
)

svc := godi.MustInject[*UserService](c)
```

### 4. 循环依赖检测

```go
type A struct{ B *B }
type B struct{ A *A }

c.MustAdd(
    godi.Build(func(c *godi.Container) (A, error) {
        b, _ := godi.Inject[B](c)
        return A{B: b}, nil
    }),
    godi.Build(func(c *godi.Container) (B, error) {
        a, _ := godi.Inject[A](c)
        return B{A: a}, nil
    }),
)

// 返回错误："circular dependency for main.A"
_, err := godi.Inject[A](c)
```

### 5. 接口注入

```go
type Database interface {
    Query(string) ([]Row, error)
}

c.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(dsn)
}))

db, err := godi.Inject[Database](c)
```

### 6. 测试 Mock

```go
// 生产环境
prod := &godi.Container{}
prod.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(prodDSN)
}))

// 测试环境
test := &godi.Container{}
test.Add(godi.Provide(&MockDatabase{Data: testData}))

// 相同的服务代码，不同的实现
svc := NewUserService(db)
```

### 7. 生命周期管理

```go
lifecycle := NewLifecycle()
c.MustAdd(godi.Provide(lifecycle))

c.Add(godi.Build(func(c *godi.Container) (*Database, error) {
    db := NewDatabase(dsn)
    lifecycle.AddShutdownHook(func(ctx context.Context) error {
        return db.Close()
    })
    return db, nil
}))

// 应用退出时
lifecycle.Shutdown(context.Background())
```

### 8. Chain 转换依赖

```go
type Name string
type Length int

c.MustAdd(
    godi.Provide(Name("hello")),
    godi.Chain(func(s Name) (Length, error) {
        return Length(len(s)), nil
    }),
)

len := godi.MustInject[Length](c) // 5
```

## 支持的类型

- 结构体：`Database`, `Config`
- 基础类型：`string`, `int`, `bool`, `float64`
- 指针：`*Database`
- 切片：`[]string`
- 映射：`map[string]int`
- 接口：`any`, 自定义接口
- 数组：`[3]int`
- 通道：`chan int`
- 函数：`func() error`

## 并发安全

所有容器操作都是线程安全的：

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// 并发注入安全
go func() {
    db, _ := godi.Inject[Database](c)
}()
```

## 示例

查看 [examples/](examples/) 获取完整示例：

| 示例 | 说明 |
|------|------|
| [01-basic](examples/01-basic/) | 基础注入 |
| [02-error-handling](examples/02-error-handling/) | 错误处理 |
| [03-must-inject](examples/03-must-inject/) | Panic 模式 |
| [04-all-types](examples/04-all-types/) | 所有支持的类型 |
| [05-multi-container](examples/05-multi-container/) | 多容器注入 |
| [06-concurrent](examples/06-concurrent/) | 并发安全 |
| [07-generics](examples/07-generics/) | 泛型注入 |
| [08-testing-mock](examples/08-testing-mock/) | Mock 测试 |
| [09-web-app](examples/09-web-app/) | Web 应用最佳实践 |
| [10-lifecycle-cleanup](examples/10-lifecycle-cleanup/) | 生命周期管理 |
| [11-chain](examples/11-chain/) | 依赖转换 |

## 与其他框架对比

| 框架 | 方式 | 特点 |
|------|------|------|
| **dig/fx** (Uber) | 反射 | 运行时解析，可能运行时错误 |
| **wire** (Google) | 代码生成 | 编译时解析，需要额外构建步骤 |
| **samber/do** | 泛型 + 反射 | 函数式 API |
| **godi** | 纯泛型 | 类型安全，无代码生成，API 简洁 |

## 许可证

MIT License
