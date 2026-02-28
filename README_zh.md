# Godi - Go 的简单依赖注入容器

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

一个轻量级、类型安全的 Go 依赖注入容器。

## 特性

- **类型安全**: 利用 Go 泛型实现编译时类型检查
- **零反射**: 无反射开销，所有类型解析在编译时完成
- **简单 API**: 最小化、直观的依赖管理接口
- **线程安全**: 内置互斥锁保护并发访问
- **零依赖**: 纯 Go 实现，无外部依赖
- **灵活**: 支持任何类型，包括结构体、基本类型、切片和映射
- **多容器**: 支持跨多个容器注入
- **循环依赖检测**: 自动检测循环依赖
- **懒加载**: 使用 `Lazy` 提供者延迟初始化

## 安装

```bash
go get github.com/Just-maple/godi
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Config struct {
    AppName string
}

func main() {
    // 创建容器
    c := &godi.Container{}
    
    // 注册依赖
    c.Add(godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}))
    c.Add(godi.Provide(Config{AppName: "my-app"}))
    
    // 注入依赖
    db, ok := godi.Inject[Database](c)
    if !ok {
        panic("failed to inject Database")
    }
    
    cfg, ok := godi.Inject[Config](c)
    if !ok {
        panic("failed to inject Config")
    }
    
    fmt.Printf("Connected to %s for %s\n", db.DSN, cfg.AppName)
}
```

## API 参考

### 容器 (Container)

#### `Add(provider Provider) bool`

在容器中注册提供者。如果相同类型的提供者已存在，返回 `false`。

```go
c := &godi.Container{}
success := c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
```

#### `ShouldAdd(provider Provider) error`

注册提供者，如果相同类型的提供者已存在则返回错误。

```go
err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
if err != nil {
    // 处理重复提供者错误
}
```

#### `MustAdd(provider Provider)`

注册提供者，如果相同类型的提供者已存在则触发 panic。

```go
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
```

### 提供者注册

#### `Provide[T any](t T) Provider`

为给定类型创建提供者。

```go
db := Database{DSN: "mysql://localhost"}
provider := godi.Provide(db)
```

#### `Lazy[T any](factory func() (T, error)) Provider`

创建懒加载提供者。工厂函数仅在首次请求依赖时调用。

```go
c.Add(godi.Lazy(func() (*Database, error) {
    return ConnectDB("mysql://localhost")
}))
```

### 依赖注入

#### `Inject[T any](c ...*Container) (v T, ok bool)`

从容器中检索依赖。如果未找到，返回零值和 `false`。支持多容器。

```go
db, ok := godi.Inject[Database](c)
if !ok {
    // 处理缺失依赖
}
```

#### `ShouldInject[T any](c ...*Container) (v T, err error)`

检索依赖，如果未找到则返回错误。包含循环依赖检测。

```go
db, err := godi.ShouldInject[Database](c)
if err != nil {
    // 处理错误（包含循环依赖检测）
}
```

#### `MustInject[T any](c ...*Container) (v T)`

检索依赖，如果未找到则触发 panic。

```go
db := godi.MustInject[Database](c)
```

#### `InjectTo[T any](v *T, c ...*Container) error`

直接将依赖注入到提供的指针中。如果未找到或检测到循环依赖则返回错误。支持多容器。

```go
var db Database
err := godi.InjectTo(&db, c)
if err != nil {
    // 处理错误
}
```

#### `MustInjectTo[T any](v *T, c ...*Container)`

直接将依赖注入到提供的指针中，如果未找到则触发 panic。

```go
var db Database
godi.MustInjectTo(&db, c)
```

## 示例

### 懒加载

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

func main() {
    c := &godi.Container{}
    
    // 懒加载 - 工厂函数在首次使用时调用
    c.Add(godi.Lazy(func() (Database, error) {
        fmt.Println("Initializing database connection...")
        return Database{DSN: "mysql://localhost"}, nil
    }))
    
    // 工厂函数在此处调用
    db, _ := godi.Inject[Database](c)
    fmt.Printf("Connected: %s\n", db.DSN)
    
    // 后续调用使用缓存值
    db2, _ := godi.Inject[Database](c)
    fmt.Printf("Cached: %s\n", db2.DSN)
}
```

### 循环依赖检测

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type A struct {
    B *B
}

type B struct {
    A *A
}

func main() {
    c := &godi.Container{}
    
    // 循环依赖：A 需要 B，B 需要 A
    c.Add(godi.Lazy(func() (A, error) {
        b, err := godi.ShouldInject[B](c)
        return A{B: b}, err
    }))
    
    c.Add(godi.Lazy(func() (B, error) {
        a, err := godi.ShouldInject[A](c)
        return B{A: a}, err
    }))
    
    // 检测循环依赖
    _, err := godi.ShouldInject[A](c)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        // 输出：Error: circular dependency for *main.A
    }
}
```

### 多容器注入

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Cache struct {
    Host string
}

func main() {
    // 创建多个容器
    dbContainer := &godi.Container{}
    cacheContainer := &godi.Container{}
    
    // 在不同容器中注册
    dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
    cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost"}))
    
    // 从多个容器注入
    db, _ := godi.Inject[Database](dbContainer, cacheContainer)
    cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
    
    fmt.Printf("DB: %s, Cache: %s\n", db.DSN, cache.Host)
}
```

### 使用 ShouldInject 处理错误

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    Port int
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(Config{Port: 8080}))
    
    // 优雅的错误处理
    config, err := godi.ShouldInject[Config](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Server starting on port %d\n", config.Port)
}
```

### 复杂依赖图

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    DSN string
}

type Database struct {
    DSN string
}

type Repository struct {
    DB Database
}

type Service struct {
    Repo Repository
}

func main() {
    c := &godi.Container{}
    
    // 注册配置
    c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))
    
    // 带依赖的懒加载
    c.Add(godi.Lazy(func() (Database, error) {
        cfg, err := godi.ShouldInject[Config](c)
        if err != nil {
            return Database{}, err
        }
        return Database{DSN: cfg.DSN}, nil
    }))
    
    c.Add(godi.Lazy(func() (Repository, error) {
        db, err := godi.ShouldInject[Database](c)
        if err != nil {
            return Repository{}, err
        }
        return Repository{DB: db}, nil
    }))
    
    c.Add(godi.Lazy(func() (Service, error) {
        repo, err := godi.ShouldInject[Repository](c)
        if err != nil {
            return Service{}, err
        }
        return Service{Repo: repo}, nil
    }))
    
    // 注入顶层服务
    svc, err := godi.ShouldInject[Service](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Service ready: DB=%s\n", svc.Repo.DB.DSN)
}
```

## 支持的类型

Godi 支持注入任何 Go 类型：

- **结构体**: `Database`, `Config`, 自定义类型
- **基本类型**: `string`, `int`, `bool`, `float64`
- **指针**: `*Database`, `*Config`
- **切片**: `[]string`, `[]int`
- **映射**: `map[string]int`, `map[string]any`
- **接口**: `any`, 自定义接口
- **数组**: `[3]int`, `[5]string`
- **通道**: `chan int`, `chan string`
- **函数**: `func()`, `func(int) string`

## 线程安全

所有容器操作都由互斥锁保护，可安全并发使用：

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// 可安全地从多个 goroutine 使用
go func() {
    db, _ := godi.Inject[Database](c)
    // 使用 db...
}()

go func() {
    db, _ := godi.Inject[Database](c)
    // 使用 db...
}()
```

## 最佳实践

1. **尽早注册依赖**: 在应用启动时设置容器
2. **使用 ShouldInject 处理错误**: 想要优雅处理错误时，优先使用 `ShouldInject` 而非 `MustInject`
3. **对昂贵资源使用 Lazy**: 数据库连接、HTTP 客户端等
4. **避免循环依赖**: 仔细设计依赖图
5. **使用多容器实现模块化**: 使用不同容器分离不同模块的关注点
6. **依赖接口**: 跨层依赖使用接口（参见 examples/09-web-app）

## 示例目录

| 示例 | 描述 |
|------|------|
| [01-basic](examples/01-basic/) | 基础依赖注入 |
| [02-error-handling](examples/02-error-handling/) | 错误处理模式 |
| [03-must-inject](examples/03-must-inject/) | 错误时 panic 的注入 |
| [04-all-types](examples/04-all-types/) | 所有支持的类型 |
| [05-multi-container](examples/05-multi-container/) | 多容器 |
| [06-concurrent](examples/06-concurrent/) | 并发访问 |
| [07-generics](examples/07-generics/) | 泛型类型注入 |
| [08-testing-mock](examples/08-testing-mock/) | 使用 Mock 测试 |
| [09-web-app](examples/09-web-app/) | 生产级 Web 应用 (SOLID 原则) |

## 对比

| 特性 | Godi | dig/fx | wire |
|------|------|--------|------|
| **类型解析** | 泛型 | 反射 | 代码生成 |
| **错误检测** | 编译时 | 运行时 | 编译时 |
| **性能** | 零开销 | 反射开销 | 零开销 |
| **设置** | 无需设置 | 无需设置 | 代码生成 |
| **多容器** | ✅ 内置 | ❌ 单一 | ❌ 单一 |
| **循环检测** | ✅ 运行时 | ✅ 运行时 | ✅ 编译时 |

## 许可证

MIT License - 详见 LICENSE 文件
