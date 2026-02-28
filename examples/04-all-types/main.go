package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// 多种类型示例：展示支持的所有类型

func main() {
	c := &godi.Container{}

	// 基本类型
	c.Add(godi.Provide("application-name"))
	c.Add(godi.Provide(42))
	c.Add(godi.Provide(int8(8)))
	c.Add(godi.Provide(int16(16)))
	c.Add(godi.Provide(int32(32)))
	c.Add(godi.Provide(int64(64)))
	c.Add(godi.Provide(uint(100)))
	c.Add(godi.Provide(uint8(8)))
	c.Add(godi.Provide(uint16(16)))
	c.Add(godi.Provide(uint32(32)))
	c.Add(godi.Provide(uint64(64)))
	c.Add(godi.Provide(float32(3.14)))
	c.Add(godi.Provide(float64(3.14159)))
	c.Add(godi.Provide(true))
	c.Add(godi.Provide(byte('A')))
	c.Add(godi.Provide(rune('中')))

	// 切片
	c.Add(godi.Provide([]string{"a", "b", "c"}))
	c.Add(godi.Provide([]int{1, 2, 3}))
	c.Add(godi.Provide([]byte{0x01, 0x02, 0x03}))

	// 映射
	c.Add(godi.Provide(map[string]int{"a": 1, "b": 2}))
	c.Add(godi.Provide(map[string]string{"key": "value"}))

	// 数组
	c.Add(godi.Provide([3]int{1, 2, 3}))
	c.Add(godi.Provide([2]string{"x", "y"}))

	// 指针
	type User struct {
		Name string
	}
	c.Add(godi.Provide(&User{Name: "Alice"}))

	// 结构体
	type Config struct {
		Host string
		Port int
	}
	c.Add(godi.Provide(Config{Host: "localhost", Port: 8080}))

	// 通道
	c.Add(godi.Provide(make(chan int, 10)))
	c.Add(godi.Provide(make(chan string, 5)))

	// 函数
	c.Add(godi.Provide(func() string { return "hello" }))
	c.Add(godi.Provide(func(x int) int { return x * 2 }))

	// 接口
	c.Add(godi.Provide(any("interface value")))

	// 注入并验证
	str, _ := godi.Inject[string](c)
	num, _ := godi.Inject[int](c)
	f32, _ := godi.Inject[float32](c)
	f64, _ := godi.Inject[float64](c)
	boolean, _ := godi.Inject[bool](c)
	strSlice, _ := godi.Inject[[]string](c)
	intMap, _ := godi.Inject[map[string]int](c)
	arr, _ := godi.Inject[[3]int](c)
	user, _ := godi.Inject[*User](c)
	config, _ := godi.Inject[Config](c)
	fn, _ := godi.Inject[func() string](c)

	fmt.Printf("字符串：%s\n", str)
	fmt.Printf("数字：%d\n", num)
	fmt.Printf("Float32: %f\n", f32)
	fmt.Printf("Float64: %f\n", f64)
	fmt.Printf("布尔：%v\n", boolean)
	fmt.Printf("字符串切片：%v\n", strSlice)
	fmt.Printf("整数映射：%v\n", intMap)
	fmt.Printf("数组：%v\n", arr)
	fmt.Printf("用户指针：%v\n", user)
	fmt.Printf("配置：%+v\n", config)
	fmt.Printf("函数调用：%s\n", fn())
}
