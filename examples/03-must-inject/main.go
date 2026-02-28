package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// Must 系列示例：使用 MustInject 和 MustAdd

type CriticalConfig struct {
	SecretKey string
}

type Database struct {
	DSN string
}

func main() {
	c := &godi.Container{}

	// 注册关键依赖
	c.MustAdd(godi.Provide(CriticalConfig{SecretKey: "super-secret-key"}))
	c.MustAdd(godi.Provide(Database{DSN: "mysql://localhost"}))

	// 使用 MustInject - 如果依赖不存在会 panic
	config := godi.MustInject[CriticalConfig](c)
	db := godi.MustInject[Database](c)

	fmt.Printf("密钥：%s\n", config.SecretKey)
	fmt.Printf("数据库：%s\n", db.DSN)

	// 使用 MustInjectTo 直接注入到变量
	var extraDB Database
	godi.MustInjectTo(&extraDB, c)
	fmt.Printf("额外数据库：%s\n", extraDB.DSN)
}
