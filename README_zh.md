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
| **多容器** | 跨容器依赖查找 |
| **并发安全** | 所有操作线程安全 |
| **接口支持** | 完整支持依赖倒置原则 |
| **Hook 系统** | 生命周期钩子管理初始化和清理 |

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
    println(db.Conn) // 输出：mysql://localhost
}
```

## 📖 核心 API

### 注册依赖

| 方法 | 说明 | 使用场景 |
|------|------|----------|
| `Provide(T)` | 注册实例值 | 简单值、配置 |
| `Build(func) (T, error)` | 注册工厂函数（懒加载，单例） | 复杂初始化、懒加载 |
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

```go
c := &godi.Container{}

// Hook 带执行计数器
startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
    if provided > 0 {
        return nil // 已调用则跳过
    }
    return func(ctx context.Context) {
        fmt.Printf("Starting: %T\n", v)
    }
})

// HookOnce - 自动单次执行
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// 执行钩子
ctx := context.Background()
shutdown(func(hooks []func(context.Context)) {
    for _, fn := range hooks {
        fn(ctx)
    }
})
```

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

// 注册关闭钩子
shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
    return func(ctx context.Context) {
        if closer, ok := v.(interface{ Close() error }); ok {
            closer.Close()
        }
    }
})

// 注册资源
c.MustAdd(
    godi.Build(func(c *godi.Container) (*Database, error) {
        return NewDatabase("dsn")
    }),
    godi.Build(func(c *godi.Container) (*Cache, error) {
        return NewCache("redis://localhost")
    }),
)

// 执行关闭
shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

shutdown(func(hooks []func(context.Context)) {
    for i := len(hooks) - 1; i >= 0; i-- {
        hooks[i](shutdownCtx)
    }
})
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

### 7. 多容器注入

```go
c1 := &godi.Container{}
c2 := &godi.Container{}

c1.MustAdd(godi.Provide(Database{DSN: "db1"}))
c2.MustAdd(godi.Provide(Config{AppName: "app2"}))

// 按顺序查找容器
db, err := godi.Inject[Database](c1, c2)
cfg, err := godi.Inject[Config](c1, c2)
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
| **多容器支持** | ✅ | ✅ (Scope) | ❌ | ✅ (Scope) |
| **循环依赖检测** | ✅ | ✅ | ✅ | ✅ |
| **懒加载** | ✅ | ✅ | ❌ | ✅ |
| **项目状态** | 活跃 | 活跃 | ⚠️ 已归档 | 活跃 |

### 何时选择 godi

- 你希望**编译时安全**但不想代码生成
- 你需要**最小依赖**和小打包体积
- 你需要资源的**生命周期管理**
- 你重视**简单直观的 API**
- 你使用**多容器**或模块化架构

## 📁 示例

完整示例请参阅 [`examples/`](examples/)：

| # | 示例 | 说明 |
|---|------|------|
| 01 | [basic](examples/01-basic/) | 基础注入模式 |
| 02 | [error-handling](examples/02-error-handling/) | 错误处理策略 |
| 03 | [must-inject](examples/03-must-inject/) | Panic 模式注入 |
| 04 | [all-types](examples/04-all-types/) | 所有支持的类型 |
| 05 | [multi-container](examples/05-multi-container/) | 跨容器注入 |
| 06 | [concurrent](examples/06-concurrent/) | 并发安全 |
| 07 | [generics](examples/07-generics/) | 高级泛型 |
| 08 | [testing-mock](examples/08-testing-mock/) | Mock 测试模式 |
| 09 | [web-app](examples/09-web-app/) | 生产级 Web 应用结构 |
| 10 | [lifecycle-cleanup](examples/10-lifecycle-cleanup/) | Hook 资源清理 |
| 11 | [chain](examples/11-chain/) | 依赖转换 |
| 12 | [struct-field-inject](examples/12-struct-field-inject/) | 结构体字段注入 |
| 13 | [hook](examples/13-hook/) | Hook 生命周期管理 |

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

1. Fork 仓库
2. 创建特性分支（`git checkout -b feature/amazing-feature`）
3. 提交更改（`git commit -m 'Add amazing feature'`）
4. 推送到分支（`git push origin feature/amazing-feature`）
5. 提交 Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件。
