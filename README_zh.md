# Godi

[![Go Reference](https://pkg.go.dev/badge/github.com/Just-maple/godi.svg)](https://pkg.go.dev/github.com/Just-maple/godi)
[![Go Report Card](https://goreportcard.com/badge/github.com/Just-maple/godi)](https://goreportcard.com/report/github.com/Just-maple/godi)
[![Test](https://github.com/Just-maple/godi/actions/workflows/test.yml/badge.svg)](https://github.com/Just-maple/godi/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/Just-maple/godi/branch/master/graph/badge.svg)](https://codecov.io/gh/Just-maple/godi)

轻量级 Go 依赖注入框架，基于泛型实现。零反射，零代码生成。

## 🚀 特性

| 特性 | 说明 |
|------|------|
| **类型安全** | 完整泛型支持，编译时类型检查 |
| **懒加载** | 依赖首次使用时初始化（单例） |
| **循环检测** | 运行时自动检测循环依赖 |
| **并发安全** | 所有操作线程安全（sync.Map） |
| **接口支持** | 完整支持依赖倒置原则 |
| **Hook 系统** | 生命周期钩子，显式执行 |
| **容器嵌套** | 树形容器结构，冻结保护 |
| **运行时添加** | Build 函数中动态注册容器 |

## 📦 安装

```bash
go get github.com/Just-maple/godi
```

## ⚡ 快速开始

```go
package main

import "github.com/Just-maple/godi"

type Config struct{ DSN string }
type Database struct{ Conn string }

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
| `Build(func) (T, error)` | 注册工厂函数（懒加载单例） | 复杂初始化 |

```go
c := &godi.Container{}

// Provide - 注册实例值
c.Add(godi.Provide(Config{Port: 8080}))

// Build - 注册工厂函数（懒加载单例）
// 模式 1: 单个依赖（自动注入）
c.Add(godi.Build(func(cfg Config) (*Database, error) {
    return NewDatabase(cfg.DSN)
}))

// 模式 2: 容器访问（多个依赖）
c.Add(godi.Build(func(c *godi.Container) (*Service, error) {
    db, _ := godi.Inject[*Database](c)
    cache, _ := godi.Inject[*Cache](c)
    return NewService(db, cache), nil
}))

// 模式 3: 无依赖（使用 struct{}）
c.Add(godi.Build(func(_ struct{}) (*Logger, error) {
    return NewLogger(), nil
}))
```

### 注入依赖

| 方法 | 返回值 | Panic | 使用场景 |
|------|--------|-------|----------|
| `Inject[T](c)` | `(T, error)` | 否 | 标准注入 |
| `MustInject[T](c)` | `T` | 是 | 确定存在 |
| `InjectTo(c, &v)` | `error` | 否 | 注入到现有变量 |
| `InjectAs(c, &v)` | `error` | 否 | 非泛型注入 |
| `c.Inject(&a, &b)` | `error` | 否 | 多重注入 |

```go
// 泛型注入
db, err := godi.Inject[*Database](c)

// 失败时 panic
db := godi.MustInject[*Database](c)

// 注入到现有变量
var db Database
err := godi.InjectTo(c, &db)

// 多重注入
err = c.Inject(&service.DB, &service.Config)
```

### 生命周期钩子

Hook 允许在依赖注入时注册回调函数。**Hook 需要显式执行** - 你必须调用返回的执行器函数。

```go
package main

import (
    "context"
    "fmt"
    "github.com/Just-maple/godi"
)

c := &godi.Container{}

// 在注入依赖前注册 hook
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// 添加并注入依赖（hook 会被注册）
c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
_, _ = godi.Inject[Database](c)

// 显式执行 hook
shutdown.Iterate(context.Background(), false)
```

**Hook 机制：**

| 方面 | 行为 |
|------|------|
| **触发时机** | 依赖注入时注册 Hook |
| **执行方式** | 显式执行 - 必须调用执行器函数 |
| **`provided` 计数器** | 跟踪注入次数（0 = 首次） |
| **HookOnce** | 当 `provided > 0` 时自动跳过 |
| **Hook** | 通过 `provided` 参数手动控制 |
| **嵌套容器** | Hook 在注入路径上的每个容器触发 |

**嵌套容器中的 Hook：**

```go
// 基础设施层
infra := &godi.Container{}
infra.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
infraHook := infra.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) { fmt.Printf("[Infra] %T\n", v) }
})

// 应用层
app := &godi.Container{}
app.MustAdd(infra)
appHook := app.HookOnce("cleanup", func(v any) func(context.Context) {
    return func(ctx context.Context) { fmt.Printf("[App] %T\n", v) }
})

// 从父容器注入 - 触发两个容器的 Hook
_, _ = godi.Inject[Database](app)

// 分别为每个容器执行 hook
ctx := context.Background()
infraHook.Iterate(ctx, false)
appHook.Iterate(ctx, false)

// 输出：
// [Infra] Database
// [App] Database
```

### 容器嵌套

容器可以嵌套创建模块化应用。子容器添加到父容器后会被**冻结**。

```go
// 子容器
child := &godi.Container{}
child.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

// 父容器嵌套子容器
parent := &godi.Container{}
parent.MustAdd(child)  // child 现在被冻结

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
| **重复检测** | Add 检查所有嵌套容器 |
| **容器冻结** | 子容器添加到父容器后被冻结 |
| **Hook 传播** | Hook 在注入路径上的每个容器触发 |
| **运行时添加** | Build 函数可以动态添加容器 |

**容器冻结：**

```go
child := &godi.Container{}
child.MustAdd(godi.Provide(Config{Value: "child"}))

parent := &godi.Container{}
parent.MustAdd(child)  // child 现在被冻结

// 这会失败："container is frozen"
err := child.Add(godi.Provide(struct{ Name string }{Name: "new"}))
```

**Build 中运行时添加（允许）：**

```go
c := &godi.Container{}
nested := &godi.Container{}
nested.MustAdd(godi.Provide("value"))

c.MustAdd(godi.Build(func(c *godi.Container) (string, error) {
    // 这是允许的 - 在 Build 执行期间添加
    c.MustAdd(nested)
    return godi.Inject[string](c)
}))
```

## 📚 使用模式

### 1. 构造函数注入

```go
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

### 2. 接口注入（依赖倒置）

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

### 3. 环境选择（运行时添加）

```go
prodDB := &godi.Container{}
prodDB.MustAdd(godi.Provide(Database{DSN: "mysql://prod-db"}))

devDB := &godi.Container{}
devDB.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

c.MustAdd(
    godi.Provide(Config{Env: "production"}),
    godi.Build(func(c *godi.Container) (Database, error) {
        cfg, _ := godi.Inject[Config](c)
        if cfg.Env == "production" {
            c.MustAdd(prodDB)
        } else {
            c.MustAdd(devDB)
        }
        return godi.Inject[Database](c)
    }),
)
```

### 4. 优雅关闭（Hook）

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
)

// 注入依赖（hook 会被注册）
_, _ = godi.Inject[*Database](c)

// 带超时执行
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
shutdown.Iterate(ctx, true)  // true = 逆序
```

### 5. 测试 Mock

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
| **容器嵌套** | ✅ | ❌ | ❌ | ❌ |
| **运行时添加** | ✅ | ❌ | ❌ | ❌ |
| **学习曲线** | 低 | 中 | 高 | 低 |

### 何时选择 godi

- 你希望**编译时安全**但不想代码生成
- 你需要**最小依赖**和小打包体积
- 你需要资源的**生命周期管理**
- 你重视**简单直观的 API**
- 你需要**容器嵌套**实现模块化架构
- 你需要**运行时容器注册**实现动态场景

## 📁 示例

完整示例请参阅 [`examples/`](examples/)：

| # | 示例 | 说明 |
|---|------|------|
| 01 | [basic](examples/01-basic/) | 基础注入模式 |
| 02 | [error-handling](examples/02-error-handling/) | 错误处理策略 |
| 03 | [all-types](examples/03-all-types/) | 所有支持的类型 + 泛型 |
| 04 | [concurrent](examples/04-concurrent/) | 并发安全 |
| 05 | [testing-mock](examples/05-testing-mock/) | Mock 测试模式 |
| 06 | [web-app](examples/06-web-app/) | 生产级 Web 应用（SOLID 原则） |
| 07 | [lifecycle-cleanup](examples/07-lifecycle-cleanup/) | Hook 资源清理 |
| 08 | [nested-container-hooks](examples/08-nested-container-hooks/) | 多层容器 Hooks |
| 09 | [runtime-container-add](examples/09-runtime-container-add/) | 动态容器注册 |

## 🤝 贡献

欢迎贡献！

1. Fork 仓库
2. 创建特性分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。
