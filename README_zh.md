# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)

轻量级 Go 依赖注入框架，基于泛型实现。零反射，零代码生成。

## 🚀 特性

| 特性 | 说明 |
|------|------|
| **类型安全** | 完整泛型支持，编译时类型检查 |
| **懒加载** | 依赖首次使用时初始化 |
| **循环检测** | 运行时自动检测循环依赖 |
| **并发安全** | 所有操作线程安全 |
| **接口支持** | 完整支持依赖倒置原则 |
| **Hook 系统** | 生命周期钩子管理初始化和清理 |
| **容器嵌套** | 树形容器组合，自动重复检测 |

## 📦 安装

```bash
go get github.com/Just-maple/godi
```

## ⚡ 快速开始

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
    
    c.MustAdd(
        godi.Provide(Config{DSN: "mysql://localhost"}),
        godi.Build(func(c *godi.Container) (*Database, error) {
            cfg, _ := godi.Inject[Config](c)
            return &Database{Conn: cfg.DSN}, nil
        }),
    )
    
    db := godi.MustInject[*Database](c)
    println(db.Conn) // 输出：mysql://localhost
}
```

## 📖 核心 API

### 注册依赖

| 方法 | 说明 | 使用场景 |
|------|------|----------|
| `Provide(T)` | 注册实例值 | 简单值、配置 |
| `Build(func) (T, error)` | 注册工厂函数（懒加载，单例） | 复杂初始化 |
| `Chain(func) (T, error)` | 从现有依赖派生 | 类型转换/派生新类型 |

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

| 方法 | 返回值 | Panic | 使用场景 |
|------|--------|-------|----------|
| `Inject[T](c)` | `(T, error)` | 否 | 标准注入 |
| `MustInject[T](c)` | `T` | 是 | 确定存在 |
| `InjectTo(&v, c)` | `error` | 否 | 注入到现有变量 |
| `InjectAs(&v, c)` | `error` | 否 | 非泛型注入 |
| `c.Inject(&a, &b)` | `error` | 否 | 多重注入 |

```go
// 泛型注入
db, err := godi.Inject[*Database](c)

// 失败时 panic
db := godi.MustInject[*Database](c)

// 注入到现有变量
var db Database
err := godi.InjectTo(&db, c)

// 多重注入
service := &Service{}
err = c.Inject(&service.DB, &service.Config)
```

### 生命周期钩子

Hook 允许在依赖注入时注册回调函数。Hook 需要**显式执行** - 你必须调用返回的执行器函数。

```go
package main

import (
    "context"
    "fmt"
    "github.com/Just-maple/godi"
)

c := &godi.Container{}

// 注册依赖
c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// Hook 带执行计数器 - 每次注入都会触发
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // 如果之前已注入过则跳过
    }
    return func(ctx context.Context) {
        fmt.Printf("Starting: %T\n", v)
    }
})

// HookOnce - 仅在首次注入时自动运行
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("Stopping: %T\n", v)
    }
})

// 注入依赖（这会注册 hook）
_, _ = godi.Inject[Database](c)
_, _ = godi.Inject[Cache](c)

// 显式执行 hook
ctx := context.Background()

// 方式 1：手动迭代
startup(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(ctx)
    }
})

// 方式 2：使用 Iterate 辅助方法（推荐）
startup.Iterate(ctx, false) // false = 正序
shutdown.Iterate(ctx, false)

// 输出：
// Starting: Database
// Starting: Cache
// Stopping: Database
// Stopping: Cache
```

**Hook 机制：**

| 方面 | 行为 |
|------|------|
| **触发时机** | 依赖注入时注册 Hook |
| **执行方式** | 显式执行 - 必须调用返回的执行器函数 |
| **`provided` 计数器** | 跟踪类型被注入的次数（0 = 首次） |
| **HookOnce** | 当 `provided > 0` 时自动跳过 |
| **Hook** | 通过 `provided` 参数手动控制 |
| **执行顺序** | Hook 按注册顺序执行 |
| **嵌套容器** | Hook 在注入路径上的每个容器触发 |

**嵌套容器中的 Hook 行为：**

Hook 在**注入路径上的每个容器**上触发。每个容器为每种类型维护独立的 `provided` 计数器：

```go
// 基础设施层
infra := &godi.Container{}
infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

infraHook := infra.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) {
        fmt.Printf("[Infra] Starting: %T\n", v)
    }
})

// 应用层
app := &godi.Container{}
app.MustAdd(infra)

appHook := app.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) {
        fmt.Printf("[App] Starting: %T\n", v)
    }
})

// 从父容器注入 - 触发两个容器的 Hook
_, _ = godi.Inject[Database](app)

// 分别为每个容器执行 hook
// 方式 1：手动迭代
infraHook(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(context.Background())
    }
})

appHook(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(context.Background())
    }
})

// 方式 2：使用 Iterate 辅助方法（推荐）
ctx := context.Background()
infraHook.Iterate(ctx, false)
appHook.Iterate(ctx, false)

// 输出：
// [Infra] Starting: Database
// [App] Starting: Database
```

**关键点：**
- Hook 在**注入路径上的每个容器**上触发
- 每个容器维护**独立的 `provided` 计数器**（按类型）
- 使用 `provided > 0` 检查确保 Hook 只在首次注入时运行
- Hook 在注入时注册，在调用执行器函数时执行
- 为每个容器分别执行 hook

**常见模式：**

```go
// 1. 资源清理（HookOnce）
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// 2. 条件初始化
// 方式 A：使用 HookOnce（推荐用于简单场景）
startup := c.HookOnce("startup", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // 初始化资源 - 自动只运行一次
    }
})

// 方式 B：使用 Hook 带 provided 检查（用于条件逻辑）
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // 仅在首次注入时初始化
    }
    return func(ctx context.Context) {
        // 初始化资源
    }
})

// 3. 基于接口的清理
c.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        switch resource := v.(type) {
        case Database:
            resource.Close()
        case Cache:
            resource.Disconnect()
        }
    }
})

// 4. 逆序优雅关闭
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        // 清理逻辑
    }
})

shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 逆序执行以确保正确的清理顺序
shutdown.Iterate(shutdownCtx, true) // true = 逆序

// 5. 多阶段生命周期
// 为每个阶段使用 HookOnce（推荐）
init := c.HookOnce("init", func(v any) func(context.Context) {
    return func(ctx context.Context) { /* 初始化 */ }
})

start := c.HookOnce("start", func(v any) func(context.Context) {
    return func(ctx context.Context) { /* 启动 */ }
})

// 或者为每个阶段使用 Hook 带 provided 检查的条件逻辑
init := c.Hook("init", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) { /* 初始化 */ }
})

start := c.Hook("start", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil
    }
    return func(ctx context.Context) { /* 启动 */ }
})

// 按顺序执行各阶段
ctx := context.Background()
init.Iterate(ctx, false)
start.Iterate(ctx, false)
```

### 容器嵌套

容器嵌套支持构建模块化、树形结构的应用，具有自动重复检测和容器冻结功能。

```go
// 子容器
child := &godi.Container{}
child.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

// 父容器嵌套子容器
parent := &godi.Container{}
parent.MustAdd(child)

// 从父容器注入（在子容器中查找 Database）
db, _ := godi.Inject[Database](parent)

// 重复类型会被阻止
err := parent.Add(godi.Provide(Database{DSN: "other"}))
// err: provider *godi.Database already exists
```

**核心机制：**

| 机制 | 行为 |
|------|------|
| **树形搜索** | Inject 深度优先遍历子容器 |
| **重复检测** | Add 检查所有嵌套容器中是否已存在该类型 |
| **容器冻结** | 子容器添加到父容器后被冻结，不能添加新提供者 |
| **Hook 传播** | Hook 在注入路径上的每个容器上触发 |
| **每容器计数器** | 每个容器跟踪自己每个类型的 `provided` 计数 |

**容器冻结机制：**

当容器被添加到父容器后，它会被**冻结**，无法再接受新的提供者：

```go
child := &godi.Container{}
child.MustAdd(godi.Provide(Config{Value: "child"}))

parent := &godi.Container{}
parent.MustAdd(child)  // child 现在被冻结

// 这会失败："container frozen cause added as provider"
err := child.Add(godi.Provide(struct{ Name string }{Name: "new"}))
```

**Provide 方法：**

检查容器（包括嵌套容器）是否提供某类型：

```go
c := &godi.Container{}
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

db := Database{}
_, ok := c.Provide(&db)
fmt.Println(ok)  // true

other := struct{ Other string }{}
_, ok = c.Provide(&other)
fmt.Println(ok)  // false
```

**嵌套容器中的 Hook 行为：**

参见 [生命周期钩子](#生命周期钩子) 章节了解嵌套容器中 Hook 的详细行为。总结：
- Hook 在**注入路径上的每个容器**上触发
- 每个容器维护**独立的 `provided` 计数器**（按类型）
- 使用每个容器的执行器函数分别执行 hook

## 📚 使用模式

### 1. 构造函数注入

```go
type Service struct {
    DB   Database
    Config Config
}

c.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Config{AppName: "my-app"}),
    godi.Build(func(c *godi.Container) (*Service, error) {
        db, _ := godi.Inject[Database](c)
        cfg, _ := godi.Inject[Config](c)
        return &Service{DB: db, Config: cfg}, nil
    }),
)
```

### 2. 字段注入

```go
service := &Service{}
err := c.Inject(&service.DB, &service.Config, &service.Cache)
```

### 3. 接口注入（依赖倒置）

```go
type Database interface {
    Query(string) ([]Row, error)
    Close() error
}

c.Add(godi.Build(func(c *godi.Container) (Database, error) {
    return NewMySQLDatabase(dsn), nil
}))

db, err := godi.Inject[Database](c)
```

### 4. 依赖链

```go
c.MustAdd(
    godi.Provide(Config{DSN: "mysql://localhost"}),
    godi.Build(func(c *godi.Container) (*Database, error) {
        cfg, _ := godi.Inject[Config](c)
        return NewDatabase(cfg.DSN)
    }),
    godi.Build(func(c *godi.Container) (*Repository, error) {
        db, _ := godi.Inject[*Database](c)
        return NewRepository(db)
    }),
    godi.Build(func(c *godi.Container) (*Service, error) {
        repo, _ := godi.Inject[*Repository](c)
        return NewService(repo)
    }),
)
```

### 5. 优雅关闭（Hook）

```go
c := &godi.Container{}

// 在注入依赖前注册 shutdown hook
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

c.MustAdd(
    godi.Build(func(c *godi.Container) (*Database, error) {
        return NewDatabase("dsn")
    }),
    godi.Build(func(c *godi.Container) (*Cache, error) {
        return NewCache("redis://localhost")
    }),
)

// 注入依赖（hook 会被注册）
_, _ = godi.Inject[*Database](c)
_, _ = godi.Inject[*Cache](c)

// 带超时的优雅关闭
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// 逆序执行 hook 以确保正确的清理顺序
shutdown.Iterate(shutdownCtx, true) // true = 逆序
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

### 7. 容器嵌套

```go
// 基础设施层
infra := &godi.Container{}
infra.MustAdd(
    godi.Provide(Database{DSN: "mysql://localhost"}),
    godi.Provide(Cache{Addr: "redis://localhost"}),
)

// 在 infra 容器上注册 hook
infraShutdown := infra.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("Infra cleanup: %T\n", v)
    }
})

// 应用层
app := &godi.Container{}
app.MustAdd(infra, godi.Provide(Config{AppName: "my-app"}))

// 在 app 容器上注册 hook
appShutdown := app.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        fmt.Printf("App cleanup: %T\n", v)
    }
})

// 从父容器注入所有依赖
db, _ := godi.Inject[Database](app)
cache, _ := godi.Inject[Cache](app)
cfg, _ := godi.Inject[Config](app)

// 为每个容器执行 hook
ctx := context.Background()
infraShutdown.Iterate(ctx, false)
appShutdown.Iterate(ctx, false)
```

### 8. Chain 转换

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

## 🔧 支持的类型

- ✅ 结构体：`Database`, `Config`
- ✅ 基础类型：`string`, `int`, `bool`, `float64`
- ✅ 指针：`*Database`
- ✅ 切片：`[]string`
- ✅ 映射：`map[string]int`
- ✅ 接口：`any`, 自定义接口
- ✅ 数组：`[3]int`
- ✅ 通道：`chan int`
- ✅ 函数：`func() error`

## 📊 框架对比

| 特性 | godi | dig/fx | wire | samber/do |
|------|------|--------|------|-----------|
| **类型系统** | 泛型 | 反射 | 代码生成 | 泛型 |
| **运行时错误** | 否 | 可能 | 否 | 可能 |
| **构建步骤** | 否 | 否 | 需要 | 否 |
| **API 风格** | 函数式 | 函数式 | 代码生成 | 函数式 |
| **学习曲线** | 低 | 中 | 高 | 低 |
| **打包体积** | 最小 | 中 | 大 | 小 |
| **生命周期钩子** | ✅ | ✅ | ❌ | ✅ |
| **循环依赖检测** | ✅ | ✅ | ✅ | ✅ |
| **懒加载** | ✅ | ✅ | ❌ | ✅ |
| **容器嵌套** | ✅ | ❌ | ❌ | ❌ |
| **项目状态** | 活跃 | 活跃 | ⚠️ 已归档 | 活跃 |

### 何时选择 godi

- 你希望**编译时安全**但不想代码生成
- 你需要**最小依赖**和小打包体积
- 你需要资源的**生命周期管理**
- 你重视**简单直观的 API**
- 你需要**容器嵌套**实现模块化架构

## 📁 示例

完整示例请参阅 [`examples/`](examples/)：

| # | 示例 | 说明 |
|---|------|------|
| 01 | [basic](examples/01-basic/) | 基础注入模式 |
| 02 | [error-handling](examples/02-error-handling/) | 错误处理策略 |
| 03 | [must-inject](examples/03-must-inject/) | Panic 模式注入 |
| 04 | [all-types](examples/04-all-types/) | 所有支持的类型 |
| 05 | [concurrent](examples/05-concurrent/) | 并发安全 |
| 06 | [generics](examples/06-generics/) | 高级泛型 |
| 07 | [testing-mock](examples/07-testing-mock/) | Mock 测试模式 |
| 08 | [web-app](examples/08-web-app/) | 生产级 Web 应用结构 |
| 09 | [lifecycle-cleanup](examples/09-lifecycle-cleanup/) | Hook 资源清理 |
| 10 | [chain](examples/10-chain/) | 依赖转换 |
| 11 | [struct-field-inject](examples/11-struct-field-inject/) | 结构体字段注入 |
| 12 | [hook](examples/12-hook/) | Hook 生命周期管理 |
| 13 | [nested-container-hooks](examples/13-nested-container-hooks/) | 多层容器 Hooks |

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 仓库
2. 创建特性分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。
