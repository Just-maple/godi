package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// 基础示例：注册和注入简单依赖

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
