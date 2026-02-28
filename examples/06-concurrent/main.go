package main

import (
	"fmt"
	"sync"

	"github.com/Just-maple/godi"
)

// 并发安全示例：展示容器的线程安全性

type Counter struct {
	Value int
}

func main() {
	c := &godi.Container{}
	c.Add(godi.Provide(Counter{Value: 0}))

	var wg sync.WaitGroup
	numGoroutines := 10

	// 多个 goroutine 同时注入
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			counter, ok := godi.Inject[Counter](c)
			if !ok {
				fmt.Printf("Goroutine %d: 注入失败\n", id)
				return
			}

			fmt.Printf("Goroutine %d: 注入成功，值=%d\n", id, counter.Value)
		}(i)
	}

	wg.Wait()
	fmt.Println("所有 goroutine 完成")

	// 演示并发注册
	c2 := &godi.Container{}
	var wg2 sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			c2.Add(godi.Provide(fmt.Sprintf("value-%d", id)))
		}(i)
	}

	wg2.Wait()

	// 验证所有值都已注册
	val, ok := godi.Inject[string](c2)
	if ok {
		fmt.Printf("注入的值：%s\n", val)
	}
	fmt.Println("并发注册演示完成！")
}
