package main

import (
	"fmt"
	"sync"

	"github.com/Just-maple/godi"
)

// Concurrency Example: Demonstrates thread-safe container operations

type Counter struct {
	Value int
}

func main() {
	c := &godi.Container{}
	c.MustAdd(godi.Provide(Counter{Value: 0}))

	var wg sync.WaitGroup
	numGoroutines := 10

	// Multiple goroutines injecting concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			counter, err := godi.Inject[Counter](c)
			if err != nil {
				fmt.Printf("Goroutine %d: Injection failed\n", id)
				return
			}

			fmt.Printf("Goroutine %d: Injection successful, value=%d\n", id, counter.Value)
		}(i)
	}

	wg.Wait()
	fmt.Println("All goroutines completed")

	// Demonstrate concurrent registration
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

	// Verify all values are registered
	val, err := godi.Inject[string](c2)
	if err == nil {
		fmt.Printf("Injected value: %s\n", val)
	}
	fmt.Println("Concurrent registration demo complete!")
}
