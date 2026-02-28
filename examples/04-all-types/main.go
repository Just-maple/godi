package main

import (
	"fmt"
	"github.com/Just-maple/godi"
)

// All Types Example: Demonstrates all supported types

func main() {
	c := &godi.Container{}

	// Basic types
	c.MustAdd(
		godi.Provide("application-name"),
		godi.Provide(42),
		godi.Provide(int8(8)),
		godi.Provide(int16(16)),
		godi.Provide(int32(32)),
		godi.Provide(int64(64)),
		godi.Provide(uint(100)),
		godi.Provide(uint8(8)),
		godi.Provide(uint16(16)),
		godi.Provide(uint32(32)),
		godi.Provide(uint64(64)),
		godi.Provide(float32(3.14)),
		godi.Provide(float64(3.14159)),
		godi.Provide(true),
	)

	// Slices
	c.MustAdd(
		godi.Provide([]string{"a", "b", "c"}),
		godi.Provide([]int{1, 2, 3}),
		godi.Provide([]byte{0x01, 0x02, 0x03}),
	)

	// Maps
	c.MustAdd(
		godi.Provide(map[string]int{"a": 1, "b": 2}),
		godi.Provide(map[string]string{"key": "value"}),
	)

	// Arrays
	c.MustAdd(
		godi.Provide([3]int{1, 2, 3}),
		godi.Provide([2]string{"x", "y"}),
	)

	// Pointers
	type User struct {
		Name string
	}
	c.MustAdd(godi.Provide(&User{Name: "Alice"}))

	// Structs
	type Config struct {
		Host string
		Port int
	}
	c.MustAdd(godi.Provide(Config{Host: "localhost", Port: 8080}))

	// Channels
	c.MustAdd(
		godi.Provide(make(chan int, 10)),
		godi.Provide(make(chan string, 5)),
	)

	// Functions
	c.MustAdd(
		godi.Provide(func() string { return "hello" }),
		godi.Provide(func(x int) int { return x * 2 }),
	)

	// Interfaces
	c.MustAdd(godi.Provide(any("interface value")))

	// Inject and verify
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

	fmt.Printf("String: %s\n", str)
	fmt.Printf("Number: %d\n", num)
	fmt.Printf("Float32: %f\n", f32)
	fmt.Printf("Float64: %f\n", f64)
	fmt.Printf("Boolean: %v\n", boolean)
	fmt.Printf("String Slice: %v\n", strSlice)
	fmt.Printf("Int Map: %v\n", intMap)
	fmt.Printf("Array: %v\n", arr)
	fmt.Printf("User Pointer: %v\n", user)
	fmt.Printf("Config: %+v\n", config)
	fmt.Printf("Function Call: %s\n", fn())
}
