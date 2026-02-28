package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// 错误处理示例：使用 ShouldInject 和 ShouldAdd

type Database struct {
	DSN string
}

type Config struct {
	Port int
}

func main() {
	c := &godi.Container{}

	// 使用 ShouldAdd 处理重复注册错误
	err := c.ShouldAdd(godi.Provide(Database{DSN: "mysql://localhost"}))
	if err != nil {
		fmt.Printf("注册失败：%v\n", err)
		return
	}
	fmt.Println("第一次注册成功")

	// 重复注册会返回错误
	err = c.ShouldAdd(godi.Provide(Database{DSN: "mysql://remote"}))
	if err != nil {
		fmt.Printf("预期错误：%v\n", err)
	}

	// 使用 ShouldInject 处理注入错误
	db, err := godi.ShouldInject[Database](c)
	if err != nil {
		fmt.Printf("注入失败：%v\n", err)
		return
	}
	fmt.Printf("数据库：%s\n", db.DSN)

	// 注入不存在的依赖会返回错误
	_, err = godi.ShouldInject[Config](c)
	if err != nil {
		fmt.Printf("预期错误：%v\n", err)
	}
}
