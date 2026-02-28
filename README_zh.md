# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)

## 概述

Godi 是一个使用 Go 泛型的依赖注入容器。

## 生态系统背景

Go 依赖注入领域有几种不同的方案：

| 项目 | 方案 | 主要特点 |
|------|------|---------|
| **dig/fx** (Uber) | 反射 | 运行时依赖解析 |
| **wire** (Google) | 代码生成 | 编译时依赖解析 |
| **samber/do** | 泛型 + 反射 | 函数式容器 API |
| **godi** | 泛型 | 直接基于类型的注入 |

各方案有不同的取舍：

- **基于反射** (dig): 灵活，无需设置，可能出现运行时错误
- **代码生成** (wire): 类型安全，需要构建步骤，设置较复杂
- **基于泛型** (godi, samber/do): 类型安全，无需代码生成，运行时解析

Godi 专注于使用 Go 泛型的最小化 API，提供自动循环依赖检测和多容器支持。

## 安装

```bash
go get github.com/Just-maple/godi
```

## 核心概念

### 容器 (Container)

`Container` 保存已注册的提供者并管理依赖注入。

```go
c := &godi.Container{}
```

### 提供者注册

**Provide** - 注册具体值：

```go
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
```

**Lazy** - 注册工厂函数，在首次请求时执行：

```go
c.Add(godi.Lazy(func() (*Database, error) {
    return sql.Open("mysql", dsn)
}))
```

### 获取依赖

```go
// 返回 (值，ok)
db, ok := godi.Inject[Database](c)

// 返回 (值，error)
db, err := godi.ShouldInject[Database](c)

// 找不到则 panic
db := godi.MustInject[Database](c)
```

---

## 使用场景

### 1. 基础注入

注册和检索简单依赖：

```go
package main

import "github.com/Just-maple/godi"

type Config struct {
    DSN string
}

func main() {
    c := &godi.Container{}
    c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))
    
    cfg, ok := godi.Inject[Config](c)
    if !ok {
        panic("Config not found")
    }
}
```

### 2. 懒加载

延迟昂贵资源的初始化直到首次使用：

```go
c.Add(godi.Lazy(func() (*Database, error) {
    // 此代码仅在首次请求 Database 时执行
    return sql.Open("mysql", dsn)
}))

// 工厂函数在此处执行
db, err := godi.Inject[*Database](c)
```

### 3. 带依赖的懒加载

懒加载工厂可以注入自己的依赖：

```go
c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))

c.Add(godi.Lazy(func() (*Database, error) {
    // 在工厂内部注入依赖
    cfg, err := godi.Inject[Config](c)
    if err != nil {
        return nil, err
    }
    return sql.Open("mysql", cfg.DSN)
}))
```

### 4. 依赖链

构建依赖链：

```go
// 第 1 层：Config
c.Add(godi.Provide(Config{DSN: "mysql://localhost"}))

// 第 2 层：Database 依赖 Config
c.Add(godi.Lazy(func() (*Database, error) {
    cfg, _ := godi.Inject[Config](c)
    return NewDatabase(cfg.DSN)
}))

// 第 3 层：Repository 依赖 Database
c.Add(godi.Lazy(func() (*UserRepository, error) {
    db, _ := godi.Inject[*Database](c)
    return NewUserRepository(db)
}))

// 第 4 层：Service 依赖 Repository
c.Add(godi.Lazy(func() (*UserService, error) {
    repo, _ := godi.Inject[*UserRepository](c)
    return NewUserService(repo)
}))

// 注入顶层服务（触发整个依赖链）
svc, err := godi.Inject[*UserService](c)
```

### 5. 循环依赖检测

循环依赖在运行时检测：

```go
type A struct{ B *B }
type B struct{ A *A }

c.Add(godi.Lazy(func() (A, error) {
    b, err := godi.Inject[B](c)
    return A{B: b}, err
}))

c.Add(godi.Lazy(func() (B, error) {
    a, err := godi.Inject[A](c)
    return B{A: a}, err
}))

// 返回错误："circular dependency for main.A"
_, err := godi.Inject[A](c)
```

### 6. 多容器注入

从多个容器注入：

```go
dbContainer := &godi.Container{}
cacheContainer := &godi.Container{}

dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost"}))
cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost"}))

// 搜索两个容器
db, _ := godi.Inject[Database](dbContainer, cacheContainer)
cache, _ := godi.Inject[Cache](dbContainer, cacheContainer)
```

### 7. InjectTo - 注入到现有变量

```go
var db Database
err := godi.InjectTo(&db, c)
if err != nil {
    // 处理错误
}
```

### 8. 基于接口的注入

注册和注入接口：

```go
// 定义接口
type Database interface {
    Query(string) ([]Row, error)
}

// 注册实现
c.Add(godi.Lazy(func() (Database, error) {
    return NewMySQLDatabase(dsn)
}))

// 注入接口
db, err := godi.Inject[Database](c)
```

### 9. 测试时使用 Mock

测试时替换实现：

```go
// 生产环境
prod := &godi.Container{}
prod.Add(godi.Lazy(func() (Database, error) {
    return NewMySQLDatabase(prodDSN)
}))

// 测试环境
test := &godi.Container{}
test.Add(godi.Provide(&MockDatabase{Data: testdata}))

// 相同的服务可用于两种环境
svc := NewUserService(db)
```

### 10. 分组相关依赖

使用结构体分组配置：

```go
type AppConfig struct {
    Database DatabaseConfig
    HTTP     HTTPConfig
    Cache    CacheConfig
}

c.Add(godi.Provide(AppConfig{
    Database: DatabaseConfig{DSN: "mysql://localhost"},
    HTTP:     HTTPConfig{Port: 8080},
    Cache:    CacheConfig{Host: "redis://localhost"},
}))

cfg, _ := godi.Inject[AppConfig](c)
```

### 11. 生命周期管理和清理

在初始化时注册清理钩子，实现优雅关闭：

```go
// 创建生命周期管理器
lifecycle := lifecycle.New("MyApp")
c.Add(godi.Provide(lifecycle))

// 注册数据库并添加清理钩子
c.Add(godi.Lazy(func() (*Database, error) {
    db := NewDatabase(dsn)
    
    // 注册清理钩子
    lifecycle.AddShutdownHook(func(ctx context.Context) error {
        return db.Close()
    })
    
    return db, nil
}))

// 注册服务并添加关闭钩子
c.Add(godi.Lazy(func() (*Service, error) {
    svc := NewService()
    
    // 注册优雅关闭钩子
    lifecycle.AddShutdownHook(func(ctx context.Context) error {
        return svc.Shutdown(ctx)
    })
    
    return svc, nil
}))

// 应用退出时，按反向顺序关闭
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
lifecycle.Shutdown(shutdownCtx)
```

钩子按反向顺序执行（LIFO - 后进先出），确保正确的清理顺序。

---

## API 参考

### 容器方法

| 方法 | 描述 |
|------|------|
| `Add(p Provider) bool` | 注册提供者。重复时返回 false |
| `ShouldAdd(p Provider) error` | 注册提供者。重复时返回 error |
| `MustAdd(p Provider)` | 注册提供者。重复时 panic |

### 注入函数

| 函数 | 签名 | 行为 |
|------|------|------|
| `Inject[T](c)` | `(T, error)` | 找不到或循环依赖时返回 error |
| `MustInject[T](c)` | `T` | 找不到时 panic |
| `InjectTo[T](&v, c)` | `error` | 注入到提供的指针 |
| `MustInjectTo[T](&v, c)` | - | 注入到指针，失败时 panic |

### 提供者函数

| 函数 | 描述 |
|------|------|
| `Provide[T](v T)` | 注册具体值 |
| `Lazy[T](func() (T, error))` | 注册工厂函数，延迟执行 |

---

## Lazy 懒加载模式

### 模式 1: 简单懒加载

```go
c.Add(godi.Lazy(func() (*Database, error) {
    return sql.Open("mysql", dsn)
}))
```

### 模式 2: 懒加载带错误处理

```go
c.Add(godi.Lazy(func() (*Database, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}))
```

### 模式 3: 懒加载带依赖

```go
c.Add(godi.Lazy(func() (*UserService, error) {
    db, err := godi.Inject[*Database](c)
    if err != nil {
        return nil, err
    }
    cache, err := godi.Inject[*Cache](c)
    if err != nil {
        return nil, err
    }
    return NewUserService(db, cache)
}))
```

### 模式 4: 懒加载单例

懒加载提供者是单例 - 工厂函数只执行一次：

```go
c.Add(godi.Lazy(func() (*ExpensiveResource, error) {
    fmt.Println("Initializing...") // 仅打印一次
    return NewExpensiveResource()
}))

// 工厂函数在此处执行
r1, _ := godi.Inject[*ExpensiveResource](c)

// 返回缓存值
r2, _ := godi.Inject[*ExpensiveResource](c)
```

### 模式 5: 条件懒加载

```go
c.Add(godi.Lazy(func() (Database, error) {
    cfg, _ := godi.Inject[Config](c)
    
    if cfg.Environment == "test" {
        return NewMockDatabase(), nil
    }
    return NewMySQLDatabase(cfg.DSN)
}))
```

### 模式 6: 多层依赖链

```go
// 5 层依赖链
c.Add(godi.Lazy(func() (Level1, error) {
    return Level1{Value: 1}, nil
}))

c.Add(godi.Lazy(func() (Level2, error) {
    l1, _ := godi.ShouldInject[Level1](c)
    return Level2{Value: l1.Value + 1}, nil
}))

c.Add(godi.Lazy(func() (Level3, error) {
    l2, _ := godi.Inject[Level2](c)
    return Level3{Value: l2.Value + 1}, nil
}))

// 注入 Level3 会触发整个链
l3, err := godi.Inject[Level3](c)
```

---

## 线程安全

所有容器操作都是线程安全的：

```go
c := &godi.Container{}
c.Add(godi.Provide(Database{DSN: "mysql://localhost"}))

// 并发使用安全
go func() {
    db, _ := godi.Inject[Database](c)
}()

go func() {
    db, _ := godi.Inject[Database](c)
}()
```

---

## 支持的类型

Godi 支持任何 Go 类型：

- 结构体：`Database`, `Config`
- 基本类型：`string`, `int`, `bool`
- 指针：`*Database`
- 切片：`[]string`
- 映射：`map[string]int`
- 接口：`any`, 自定义接口
- 数组：`[3]int`
- 通道：`chan int`
- 函数：`func() error`

---

## 示例

完整示例见 [examples/](examples/)：

| 示例 | 描述 |
|------|------|
| 01-basic | 基础注入 |
| 02-error-handling | 错误处理 |
| 03-must-inject | 错误时 panic |
| 04-all-types | 所有支持的类型 |
| 05-multi-container | 多容器 |
| 06-concurrent | 并发访问 |
| 07-generics | 泛型类型 |
| 08-testing-mock | Mock 测试 |
| 09-web-app | 生产级 Web 应用 |
| 10-lifecycle-cleanup | 生命周期管理和清理 |

## 许可证

MIT License
