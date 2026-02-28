package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// 多容器注入示例：展示如何跨多个容器注入依赖

type Database struct {
	DSN string
}

type Cache struct {
	Host string
	Port int
}

type Config struct {
	AppName string
}

func main() {
	// 创建多个容器
	dbContainer := &godi.Container{}
	cacheContainer := &godi.Container{}
	configContainer := &godi.Container{}

	// 在不同容器中注册不同的依赖
	dbContainer.Add(godi.Provide(Database{DSN: "mysql://localhost:3306/mydb"}))
	cacheContainer.Add(godi.Provide(Cache{Host: "redis://localhost", Port: 6379}))
	configContainer.Add(godi.Provide(Config{AppName: "multi-container-app"}))

	// 从单个容器注入
	db, ok := godi.Inject[Database](dbContainer)
	if !ok {
		panic("failed to inject Database")
	}
	fmt.Printf("数据库：%s\n", db.DSN)

	// 从多个容器注入
	// Inject 会按顺序在所有提供的容器中搜索
	cache, ok := godi.Inject[Cache](dbContainer, cacheContainer)
	if !ok {
		panic("failed to inject Cache")
	}
	fmt.Printf("缓存：%s:%d\n", cache.Host, cache.Port)

	// 从三个容器注入
	cfg, ok := godi.Inject[Config](dbContainer, cacheContainer, configContainer)
	if !ok {
		panic("failed to inject Config")
	}
	fmt.Printf("应用：%s\n", cfg.AppName)

	// 使用 InjectTo 跨容器注入
	var extraDB Database
	ok = godi.InjectTo(&extraDB, dbContainer, cacheContainer)
	fmt.Printf("额外注入：%v\n", ok)

	// 使用 ShouldInject 跨容器
	extraCache, err := godi.ShouldInject[Cache](cacheContainer, configContainer)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ShouldInject 缓存：%s:%d\n", extraCache.Host, extraCache.Port)

	fmt.Println("\n多容器注入演示完成！")
}
