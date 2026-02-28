# Godi - Go 语言的轻量级依赖注入容器

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

一个轻量级、类型安全的 Go 语言依赖注入容器。

## 特性

- **类型安全**: 利用 Go 泛型实现编译时类型安全
- **零反射**: 无反射开销，所有类型解析在编译时完成
- **简单 API**: 最小化、直观的依赖管理接口
- **线程安全**: 内置互斥锁保护并发访问
- **零依赖**: 纯 Go 实现，无外部依赖
- **灵活**: 支持任何类型，包括结构体、基本类型、切片和映射
- **多容器**: 支持跨多个容器注入

## 安装

```bash
go get github.com/Just-maple/godi
```

## 与其他 DI 容器对比

### Godi vs 其他 DI 容器

| 特性 | Godi | dig/fx | wire | Facebook Inject | samber/do |
|------|------|--------|------|-----------------|-----------|
| **类型解析** | 泛型 | 反射 | 代码生成 | 反射 | 泛型 + 反射 |
| **错误检测** | 编译时 | 运行时 | 编译时 | 运行时 | 运行时 |
| **性能** | 零开销 | 反射开销 | 零开销 | 反射开销 | 反射开销 |
| **二进制大小** | 小 | 较大 | 小 | 较大 | 较大 |
| **IDE 支持** | 完整自动补全 | 有限 | 生成代码 | 有限 | 有限 |
| **配置** | 无需配置 | 无需配置 | 需要生成 | 无需配置 | 无需配置 |
| **学习曲线** | 低 | 中等 | 高 | 中等 | 低 |
| **多容器** | 内置支持 | 单容器 | 单容器 | 单容器 | 支持 |
| **外部依赖** | 无 | dig | wire | 无 | 无 |
| **Go 版本** | 1.18+ | 任意 | 任意 | 任意 | 1.21+ |

### 各容器简介

- **dig/fx**: Uber 出品的依赖注入库，基于反射实现，功能丰富
- **wire**: Google 开发的代码生成式 DI 工具，类型安全
- **Facebook Inject**: Facebook 的依赖注入库，基于反射，简单易用
- **samber/do**: 基于泛型的轻量级容器，提供类似功能

### 为什么选择 Godi

1. **零反射**: 所有类型解析在编译时完成，无运行时开销
2. **泛型支持**: 充分利用 Go 1.18+ 泛型特性，类型安全
3. **多容器**: 内置多容器支持，适合模块化架构
4. **零依赖**: 纯 Go 实现，无外部依赖
5. **简单易用**: API 简洁，学习成本低

### 为什么零反射很重要

```go
// 基于反射 (dig/fx) - 可能出现运行时错误
container.Invoke(func(service Service) {})
// ❌ 如果 Service 未注册，只在运行时报错

// Godi - 编译时安全
db, ok := godi.Inject[Database](c)
// ✅ 类型不匹配时编译出错
// ✅ IDE 自动补全
// ✅ 无反射开销
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
    db, _ := godi.Inject[Database](c)
    cfg, _ := godi.Inject[Config](c)
    
    fmt.Printf("Connected to %s for %s\n", db.DSN, cfg.AppName)
}
```

## API 参考

### Container (容器)

`Container` 类型保存所有注册的提供者并管理依赖注入。

#### `Add(provider Provider) bool`

在容器中注册提供者。如果相同类型的提供者已存在，返回 `false`。

```go
c := &godi.Container{}
success := c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
```

#### `ShouldAdd(provider Provider) error`

注册提供者，如果相同类型的提供者已存在则返回错误。

```go
c := &godi.Container{}
err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
if err != nil {
    // 处理重复提供者错误
}
```

#### `MustAdd(provider Provider)`

注册提供者，如果相同类型的提供者已存在则触发 panic。

```go
c := &godi.Container{}
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
```

### 提供者注册

#### `Provide[T any](t T) Provider`

为给定类型创建提供者。

```go
db := Database{DSN: "mysql://localhost"}
provider := godi.Provide(db)
```

### 依赖注入

#### `Inject[T any](containers ...*Container) (v T, ok bool)`

从容器中获取依赖。如果未找到，返回零值和 `false`。支持多个容器。

```go
db, ok := godi.Inject[Database](c)
if !ok {
    // 处理缺失依赖
}
```

#### `ShouldInject[T any](containers ...*Container) (v T, err error)`

获取依赖，如果未找到则返回错误。

```go
db, err := godi.ShouldInject[Database](c)
if err != nil {
    // 处理错误
}
```

#### `MustInject[T any](containers ...*Container) (v T)`

获取依赖，如果未找到则触发 panic。

```go
db := godi.MustInject[Database](c)
```

#### `InjectTo[T any](v *T, containers ...*Container) (ok bool)`

直接将依赖注入到提供的指针中。支持多个容器。

```go
var db Database
ok := godi.InjectTo(&db, c)
```

#### `ShouldInjectTo[T any](v *T, containers ...*Container) error`

注入依赖，如果未找到则返回错误。

```go
var db Database
err := godi.ShouldInjectTo(&db, c)
```

#### `MustInjectTo[T any](v *T, containers ...*Container)`

注入依赖，如果未找到则触发 panic。

```go
var db Database
godi.MustInjectTo(&db, c)
```

## 示例

### 基础用法

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Database struct {
    DSN string
}

type Logger struct {
    Level string
}

type Service struct {
    DB     Database
    Logger Logger
}

func main() {
    c := &godi.Container{}
    
    // 注册依赖
    c.Add(godi.Provide(Database{DSN: "postgres://localhost/mydb"}))
    c.Add(godi.Provide(Logger{Level: "info"}))
    
    // 注入到结构体
    db, _ := godi.Inject[Database](c)
    logger, _ := godi.Inject[Logger](c)
    
    service := Service{
        DB:     db,
        Logger: logger,
    }
    
    fmt.Printf("Service ready: %v\n", service)
}
```

### 使用 ShouldInject 进行错误处理

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
    
    // 使用 ShouldInject 进行适当的错误处理
    config, err := godi.ShouldInject[Config](c)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Server starting on port %d\n", config.Port)
}
```

### 支持不同类型

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

func main() {
    c := &godi.Container{}
    
    // 注册各种类型
    c.Add(godi.Provide("application-name"))
    c.Add(godi.Provide(42))
    c.Add(godi.Provide(3.14))
    c.Add(godi.Provide(true))
    c.Add(godi.Provide([]string{"a", "b", "c"}))
    c.Add(godi.Provide(map[string]int{"x": 1}))
    
    str, _ := godi.Inject[string](c)
    num, _ := godi.Inject[int](c)
    slice, _ := godi.Inject[[]string](c)
    
    fmt.Printf("String: %s, Number: %d, Slice: %v\n", str, num, slice)
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
    // Inject 会在所有提供的容器中搜索
    db, _ := godi.Inject[Database](dbContainer, cacheContainer)
    cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
    
    fmt.Printf("DB: %s, Cache: %s\n", db.DSN, cache.Host)
}
```

### 使用 InjectTo

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    AppName string
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(Config{AppName: "my-app"}))
    
    // 使用 InjectTo 直接注入到变量
    var cfg Config
    ok := godi.InjectTo(&cfg, c)
    
    fmt.Printf("Injected: %v, App: %s\n", ok, cfg.AppName)
}
```

### 实际示例：Web 应用

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type DBConnection struct {
    DSN string
}

type Config struct {
    DSN  string
    Port int
}

type UserRepository struct {
    DB DBConnection
}

type UserService struct {
    Repo UserRepository
}

func main() {
    c := &godi.Container{}
    
    // 注册配置
    c.Add(godi.Provide(Config{
        DSN:  "postgres://localhost/mydb",
        Port: 8080,
    }))
    
    // 注册数据库连接
    c.Add(godi.Provide(DBConnection{DSN: "postgres://localhost/mydb"}))
    
    // 注册仓库和服务
    c.Add(godi.Provide(UserRepository{}))
    c.Add(godi.Provide(UserService{}))
    
    // 注入并使用
    config, _ := godi.Inject[Config](c)
    fmt.Printf("App configured on port %d\n", config.Port)
}
```

### 使用 MustInject 处理关键依赖

```go
package main

import (
    "github.com/Just-maple/godi"
)

type CriticalConfig struct {
    SecretKey string
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(CriticalConfig{SecretKey: "my-secret"}))
    
    // 当依赖至关重要时使用 MustInject
    config := godi.MustInject[CriticalConfig](c)
    
    // 继续使用有保证的配置
    _ = config
}
```

### 使用指针类型

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
    
    // 注册指针类型
    dbPtr := &Database{DSN: "mysql://localhost"}
    c.Add(godi.Provide(dbPtr))
    
    // 注入指针类型
    got, _ := godi.Inject[*Database](c)
    fmt.Printf("DB: %s\n", got.DSN)
}
```

### 处理重复提供者

```go
package main

import (
    "fmt"
    "github.com/Just-maple/godi"
)

type Config struct {
    Value string
}

func main() {
    c := &godi.Container{}
    
    // 第一次注册
    err := c.ShouldAdd(godi.Provide(Config{Value: "first"}))
    fmt.Printf("First add: %v\n", err) // <nil>
    
    // 重复注册会返回错误
    err = c.ShouldAdd(godi.Provide(Config{Value: "second"}))
    fmt.Printf("Second add: %v\n", err) // provider *di.Config already exists
}
```

## 为什么零反射很重要

与依赖反射的其他 DI 容器（如 `dig` 或 `fx`）不同，Godi 使用 Go 泛型进行类型解析：

| 方面 | 基于反射 | Godi (泛型) |
|------|---------|-------------|
| 性能 | 运行时开销 | 编译时解析 |
| 类型安全 | 运行时错误 | 编译时错误 |
| 二进制大小 | 较大 | 更小 |
| IDE 支持 | 有限 | 完整自动补全 |
| 代码混淆 | 容易失效 | 正常工作 |

```go
// 基于反射（可能出现运行时错误）
container.Invoke(func(service Service) {}) // 如果未注册，运行时出错

// Godi（编译时安全）
db, ok := godi.Inject[Database](c) // 类型不匹配时编译出错
```

## 支持的类型

Godi 支持注入任何 Go 类型：

- **结构体**: `Database`、`Config`、自定义类型
- **基本类型**: `string`、`int`、`bool`、`float64`
- **指针**: `*Database`、`*Config`
- **切片**: `[]string`、`[]int`
- **映射**: `map[string]int`、`map[string]any`
- **接口**: `any`、自定义接口
- **数组**: `[3]int`、`[5]string`
- **通道**: `chan int`、`chan string`
- **函数**: `func()`, `func(int) string`

## 线程安全

所有容器操作都由互斥锁保护，可安全并发使用：

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// 从多个 goroutine 安全使用
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

1. **尽早注册依赖**: 在应用程序启动时设置容器
2. **使用 ShouldInject 进行错误处理**: 当需要优雅地处理错误时，优先使用 `ShouldInject` 而不是 `MustInject`
3. **分组相关依赖**: 考虑使用结构体来分组相关配置
4. **避免循环依赖**: 仔细设计依赖图
5. **使用多个容器实现模块化**: 使用不同的容器来分离不同模块的关注点

## 许可证

MIT License - 详见 LICENSE 文件
