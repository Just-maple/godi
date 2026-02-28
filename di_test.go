package godi

import (
	"fmt"
	"math/rand"
	"testing"
)

type Database struct {
	DSN string
}

type Config struct {
	AppName string
}

type Service struct {
	Name string
	DB   Database
	Cfg  Config
}

func TestProvide(t *testing.T) {
	db := Database{DSN: "mysql://localhost:3306/test"}
	provider := Provide(db)

	var got Database
	ok := provider.Inject(&got)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got.DSN != db.DSN {
		t.Errorf("expected DSN %s, got %s", db.DSN, got.DSN)
	}

	id := provider.ID()
	if _, ok := id.(*Database); !ok {
		t.Errorf("expected ID to be *Database, got %T", id)
	}
}

func TestContainer_AddAndInject(t *testing.T) {
	c := &Container{}
	_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))
	_ = c.Add(Provide(Config{AppName: "test-app"}))

	gotDB, err := ShouldInject[Database](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDB.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", gotDB.DSN)
	}

	gotCfg, err := ShouldInject[Config](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", gotCfg.AppName)
	}
}

func TestContainer_Add(t *testing.T) {
	t.Run("add unique provider", func(t *testing.T) {
		c := &Container{}
		err := c.ShouldAdd(Provide(Database{DSN: "mysql://localhost"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("add duplicate provider returns error", func(t *testing.T) {
		c := &Container{}
		_ = c.Add(Provide(Database{DSN: "mysql://localhost"}))
		err := c.ShouldAdd(Provide(Database{DSN: "mysql://remote"}))
		if err == nil {
			t.Fatal("expected error for duplicate provider")
		}
	})
}

type L0 int
type L1 int
type L2 int
type L3 int
type L4 int
type L5 int
type L6 int
type L7 int
type L8 int
type L9 int

func TestLazyDependencyOrderIndependent_Shuffle(t *testing.T) {
	tests := []struct {
		name      string
		nodeCount int
	}{
		{"Count5", 5},
		{"Count8", 8},
		{"Count10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iterations := 20
			for iter := 0; iter < iterations; iter++ {
				c := &Container{}
				order := rand.Perm(tt.nodeCount)

				for _, idx := range order {
					switch idx {
					case 0:
						c.Add(Lazy(func() (L0, error) { return 1, nil }))
					case 1:
						c.Add(Lazy(func() (L1, error) {
							v, err := ShouldInject[L0](c)
							return L1(v) + 1, err
						}))
					case 2:
						c.Add(Lazy(func() (L2, error) {
							v, err := ShouldInject[L1](c)
							return L2(v) + 1, err
						}))
					case 3:
						c.Add(Lazy(func() (L3, error) {
							v, err := ShouldInject[L2](c)
							return L3(v) + 1, err
						}))
					case 4:
						c.Add(Lazy(func() (L4, error) {
							v, err := ShouldInject[L3](c)
							return L4(v) + 1, err
						}))
					case 5:
						c.Add(Lazy(func() (L5, error) {
							v, err := ShouldInject[L4](c)
							return L5(v) + 1, err
						}))
					case 6:
						c.Add(Lazy(func() (L6, error) {
							v, err := ShouldInject[L5](c)
							return L6(v) + 1, err
						}))
					case 7:
						c.Add(Lazy(func() (L7, error) {
							v, err := ShouldInject[L6](c)
							return L7(v) + 1, err
						}))
					case 8:
						c.Add(Lazy(func() (L8, error) {
							v, err := ShouldInject[L7](c)
							return L8(v) + 1, err
						}))
					case 9:
						c.Add(Lazy(func() (L9, error) {
							v, err := ShouldInject[L8](c)
							return L9(v) + 1, err
						}))
					}
				}

				var err error
				switch tt.nodeCount {
				case 5:
					_, err = ShouldInject[L4](c)
				case 8:
					_, err = ShouldInject[L7](c)
				case 10:
					_, err = ShouldInject[L9](c)
				}
				if err != nil {
					t.Fatalf("iter %d: injection failed: %v", iter, err)
				}
			}
		})
	}
}

type NodeResult struct {
	ID       int
	Value    int
	Computed bool
}

var (
	_ = L0(0)
	_ = L1(0)
	_ = L2(0)
	_ = L3(0)
	_ = L4(0)
	_ = L5(0)
	_ = L6(0)
	_ = L7(0)
	_ = L8(0)
	_ = L9(0)
)

func TestLazyLargeDependencyGraph(t *testing.T) {
	tests := []struct {
		name       string
		nodeCount  int
		setupDeps  func() map[int][]int
		validateFn func() error
	}{
		{
			name:      "LinearChain",
			nodeCount: 5,
		},
		{
			name:      "BinaryTree",
			nodeCount: 5,
		},
		{
			name:      "Diamond",
			nodeCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shuffleRuns := 10

			for run := 0; run < shuffleRuns; run++ {
				var deps map[int][]int
				switch tt.name {
				case "LinearChain":
					deps = map[int][]int{0: {}, 1: {0}, 2: {1}, 3: {2}, 4: {3}}
				case "BinaryTree":
					deps = map[int][]int{0: {}, 1: {0}, 2: {0}, 3: {1}, 4: {1}}
				case "Diamond":
					deps = map[int][]int{0: {}, 1: {0}, 2: {0}, 3: {1, 2}, 4: {3}}
				}

				c := &Container{}

				for i := 0; i < tt.nodeCount; i++ {
					nodeDeps := deps[i]
					switch i {
					case 0:
						c.Add(Lazy(func() (L0, error) { return 100, nil }))
					case 1:
						c.Add(Lazy(func(d []int) func() (L1, error) {
							return func() (L1, error) {
								v, _ := ShouldInject[L0](c)
								return L1(int(v) + 10), nil
							}
						}(nodeDeps)))
					case 2:
						c.Add(Lazy(func(d []int) func() (L2, error) {
							return func() (L2, error) {
								v, _ := ShouldInject[L0](c)
								return L2(int(v) + 20), nil
							}
						}(nodeDeps)))
					case 3:
						c.Add(Lazy(func(d []int) func() (L3, error) {
							return func() (L3, error) {
								v1, _ := ShouldInject[L1](c)
								v2, _ := ShouldInject[L2](c)
								return L3(int(v1) + int(v2) + 30), nil
							}
						}(nodeDeps)))
					case 4:
						c.Add(Lazy(func(d []int) func() (L4, error) {
							return func() (L4, error) {
								v, _ := ShouldInject[L3](c)
								return L4(int(v) + 40), nil
							}
						}(nodeDeps)))
					}
				}

				order := rand.Perm(tt.nodeCount)
				providers := c.Providers()
				c = &Container{}
				for _, idx := range order {
					c.providers = append(c.providers, providers[idx])
				}

				_, err := ShouldInject[L4](c)
				if err != nil {
					t.Fatalf("run %d: injection failed: %v", run, err)
				}
			}
			t.Logf("%s: all %d shuffle runs passed", tt.name, shuffleRuns)
		})
	}
}

func (c *Container) Providers() []Provider {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]Provider(nil), c.providers...)
}

func TestProvide_DifferentTypes(t *testing.T) {
	c := &Container{}
	_ = c.Add(Provide(Database{DSN: "mysql://localhost"}))
	_ = c.Add(Provide(Config{AppName: "app"}))
	_ = c.Add(Provide("string-value"))
	_ = c.Add(Provide(42))

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)
	str, _ := Inject[string](c)
	num, _ := Inject[int](c)

	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
	if cfg.AppName != "app" {
		t.Errorf("expected app, got %s", cfg.AppName)
	}
	if str != "string-value" {
		t.Errorf("expected string-value, got %s", str)
	}
	if num != 42 {
		t.Errorf("expected 42, got %d", num)
	}
}

func TestInject_ServiceWithDependencies(t *testing.T) {
	c := &Container{}
	_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))
	_ = c.Add(Provide(Config{AppName: "test-app"}))

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)

	svc := Service{
		Name: "my-service",
		DB:   db,
		Cfg:  cfg,
	}

	if svc.DB.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", svc.DB.DSN)
	}
	if svc.Cfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", svc.Cfg.AppName)
	}
}

func TestInjectTo(t *testing.T) {
	c := &Container{}
	_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))

	var db Database
	err := InjectTo(&db, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
	}
}

func ExampleProvide() {
	c := &Container{}
	_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))
	_ = c.Add(Provide(Config{AppName: "my-app"}))

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)

	fmt.Printf("DB: %s, App: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost:3306/test, App: my-app
}

func ExampleContainer_Add() {
	c := &Container{}

	db := Provide(Database{DSN: "mysql://localhost"})
	cfg := Provide(Config{AppName: "test"})

	_ = c.Add(db)
	_ = c.Add(cfg)

	gotDB, _ := Inject[Database](c)
	gotCfg, _ := Inject[Config](c)

	fmt.Printf("Providers: %d\n", len(c.providers))
	fmt.Printf("DB: %s, Config: %s\n", gotDB.DSN, gotCfg.AppName)
	// Output: Providers: 2
	// DB: mysql://localhost, Config: test
}

func TestContainer_MustAdd(t *testing.T) {
	t.Run("MustAdd succeeds for unique provider", func(t *testing.T) {
		c := &Container{}
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	})

	t.Run("MustAdd panics for duplicate provider", func(t *testing.T) {
		c := &Container{}
		c.Add(Provide(Database{DSN: "mysql://localhost"}))
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for duplicate provider")
			}
		}()
		c.MustAdd(Provide(Database{DSN: "mysql://remote"}))
	})
}

func TestMustInjectTo(t *testing.T) {
	t.Run("MustInjectTo succeeds for existing provider", func(t *testing.T) {
		c := &Container{}
		c.Add(Provide(Database{DSN: "mysql://localhost"}))

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()

		var db Database
		MustInjectTo(&db, c)
		if db.DSN != "mysql://localhost" {
			t.Errorf("expected mysql://localhost, got %s", db.DSN)
		}
	})

	t.Run("MustInjectTo panics for non-existent provider", func(t *testing.T) {
		c := &Container{}

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for non-existent provider")
			}
		}()

		var db Database
		MustInjectTo(&db, c)
	})
}

func TestInjectTo_ErrorCase(t *testing.T) {
	c := &Container{}
	var db Database
	err := InjectTo(&db, c)
	if err == nil {
		t.Fatal("expected error for non-existent provider")
	}
}

func TestMustInject(t *testing.T) {
	t.Run("MustInject succeeds for existing provider", func(t *testing.T) {
		c := &Container{}
		c.Add(Provide(Database{DSN: "mysql://localhost"}))

		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("unexpected panic: %v", r)
			}
		}()

		db := MustInject[Database](c)
		if db.DSN != "mysql://localhost" {
			t.Errorf("expected mysql://localhost, got %s", db.DSN)
		}
	})

	t.Run("MustInject panics for non-existent provider", func(t *testing.T) {
		c := &Container{}

		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for non-existent provider")
			}
		}()

		_ = MustInject[Database](c)
	})
}

func TestContainer_ProviderID(t *testing.T) {
	c := &Container{}
	p := Provide(Database{DSN: "test"})
	c.Add(p)

	id := p.ID()
	if id == nil {
		t.Fatal("expected non-nil ID")
	}
}

func TestInjectTo_StructPointer(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "mysql://localhost"}))

	var db Database
	err := InjectTo(&db, c)
	if err != nil {
		t.Fatal("expected InjectTo to succeed")
	}
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestInjectTo_WrongType(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "mysql://localhost"}))

	var cfg Config
	err := InjectTo(&cfg, c)
	if err == nil {
		t.Fatal("expected InjectTo to fail for wrong type")
	}
}

func TestProvide_PointerType(t *testing.T) {
	c := &Container{}
	dbPtr := &Database{DSN: "mysql://localhost"}
	c.Add(Provide(dbPtr))

	got, ok := Inject[*Database](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", got.DSN)
	}
}

func TestProvide_SliceType(t *testing.T) {
	c := &Container{}
	slice := []string{"a", "b", "c"}
	c.Add(Provide(slice))

	got, ok := Inject[[]string](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if len(got) != 3 {
		t.Errorf("expected length 3, got %d", len(got))
	}
}

func TestProvide_MapType(t *testing.T) {
	c := &Container{}
	m := map[string]int{"a": 1, "b": 2}
	c.Add(Provide(m))

	got, ok := Inject[map[string]int](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got["a"] != 1 || got["b"] != 2 {
		t.Errorf("expected map values, got %v", got)
	}
}

func TestProvide_BoolType(t *testing.T) {
	c := &Container{}
	c.Add(Provide(true))

	got, ok := Inject[bool](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if !got {
		t.Errorf("expected true, got %v", got)
	}
}

func TestProvide_FloatType(t *testing.T) {
	c := &Container{}
	c.Add(Provide(3.14))

	got, ok := Inject[float64](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got != 3.14 {
		t.Errorf("expected 3.14, got %v", got)
	}
}

func TestContainer_MultipleProviders(t *testing.T) {
	c := &Container{}
	for i := 0; i < 10; i++ {
		c.Add(Provide(Config{AppName: fmt.Sprintf("app-%d", i)}))
	}

	if len(c.providers) != 1 {
		t.Errorf("expected 1 provider (duplicates rejected), got %d", len(c.providers))
	}
}

func TestInject_EmptyContainer(t *testing.T) {
	c := &Container{}

	_, ok := Inject[Database](c)
	if ok {
		t.Fatal("expected Inject to return false for empty container")
	}
}

func TestShouldInject_ErrorMessage(t *testing.T) {
	c := &Container{}
	_, err := ShouldInject[Database](c)
	if err == nil {
		t.Fatal("expected error")
	}
	expectedMsg := "provider *godi.Database not found"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message %q, got %q", expectedMsg, err.Error())
	}
}

func TestContainer_ShouldAdd_ErrorMessage(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "test"}))
	err := c.ShouldAdd(Provide(Database{DSN: "test2"}))
	if err == nil {
		t.Fatal("expected error")
	}
	expectedPrefix := "provider *godi.Database already exists"
	if err.Error() != expectedPrefix {
		t.Errorf("expected error message %q, got %q", expectedPrefix, err.Error())
	}
}

func ExampleMustInject() {
	c := &Container{}
	c.Add(Provide(Database{DSN: "mysql://localhost"}))

	db := MustInject[Database](c)
	fmt.Printf("DB: %s\n", db.DSN)
	// Output: DB: mysql://localhost
}

func ExampleShouldInject() {
	c := &Container{}
	c.Add(Provide(Config{AppName: "my-app"}))

	cfg, err := ShouldInject[Config](c)
	fmt.Printf("App: %s, Error: %v\n", cfg.AppName, err)
	// Output: App: my-app, Error: <nil>
}

func ExampleInjectTo() {
	c := &Container{}
	c.Add(Provide(Database{DSN: "mysql://localhost"}))

	var db Database
	err := InjectTo(&db, c)
	fmt.Printf("Injected: %v, DSN: %s\n", err == nil, db.DSN)
	// Output: Injected: true, DSN: mysql://localhost
}

func ExampleContainer_ShouldAdd() {
	c := &Container{}
	err := c.ShouldAdd(Provide(Service{Name: "test"}))
	fmt.Printf("Error: %v\n", err)
	// Output: Error: <nil>
}

func ExampleContainer_MustAdd() {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	db, _ := Inject[Database](c)
	fmt.Printf("DB: %s\n", db.DSN)
	// Output: DB: mysql://localhost
}

func TestProvide_IntTypes(t *testing.T) {
	c := &Container{}
	c.Add(Provide(int8(8)))
	c.Add(Provide(int16(16)))
	c.Add(Provide(int32(32)))
	c.Add(Provide(int64(64)))

	i8, ok := Inject[int8](c)
	if !ok || i8 != 8 {
		t.Errorf("expected int8(8), got %v", i8)
	}

	i16, ok := Inject[int16](c)
	if !ok || i16 != 16 {
		t.Errorf("expected int16(16), got %v", i16)
	}

	i32, ok := Inject[int32](c)
	if !ok || i32 != 32 {
		t.Errorf("expected int32(32), got %v", i32)
	}

	i64, ok := Inject[int64](c)
	if !ok || i64 != 64 {
		t.Errorf("expected int64(64), got %v", i64)
	}
}

func TestProvide_UIntTypes(t *testing.T) {
	c := &Container{}
	c.Add(Provide(uint(100)))
	c.Add(Provide(uint8(8)))
	c.Add(Provide(uint16(16)))
	c.Add(Provide(uint32(32)))
	c.Add(Provide(uint64(64)))

	u, ok := Inject[uint](c)
	if !ok || u != 100 {
		t.Errorf("expected uint(100), got %v", u)
	}

	u8, ok := Inject[uint8](c)
	if !ok || u8 != 8 {
		t.Errorf("expected uint8(8), got %v", u8)
	}

	u16, ok := Inject[uint16](c)
	if !ok || u16 != 16 {
		t.Errorf("expected uint16(16), got %v", u16)
	}

	u32, ok := Inject[uint32](c)
	if !ok || u32 != 32 {
		t.Errorf("expected uint32(32), got %v", u32)
	}

	u64, ok := Inject[uint64](c)
	if !ok || u64 != 64 {
		t.Errorf("expected uint64(64), got %v", u64)
	}
}

func TestProvide_ByteAndRune(t *testing.T) {
	c := &Container{}
	c.Add(Provide(byte('A')))
	c.Add(Provide(rune('中')))

	b, ok := Inject[byte](c)
	if !ok || b != 'A' {
		t.Errorf("expected byte('A'), got %v", b)
	}

	r, ok := Inject[rune](c)
	if !ok || r != '中' {
		t.Errorf("expected rune('中'), got %v", r)
	}
}

func TestProvide_ArrayType(t *testing.T) {
	c := &Container{}
	arr := [3]int{1, 2, 3}
	c.Add(Provide(arr))

	got, ok := Inject[[3]int](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got != arr {
		t.Errorf("expected %v, got %v", arr, got)
	}
}

func TestProvide_Channel(t *testing.T) {
	c := &Container{}
	ch := make(chan int, 10)
	c.Add(Provide(ch))

	got, ok := Inject[chan int](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got != ch {
		t.Error("expected same channel")
	}
}

func TestProvide_Function(t *testing.T) {
	c := &Container{}
	fn := func() string { return "hello" }
	c.Add(Provide(fn))

	got, ok := Inject[func() string](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got() != "hello" {
		t.Errorf("expected 'hello', got %s", got())
	}
}

func TestProvide_FunctionWithArgs(t *testing.T) {
	c := &Container{}
	fn := func(x int) int { return x * 2 }
	c.Add(Provide(fn))

	got, ok := Inject[func(int) int](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got(5) != 10 {
		t.Errorf("expected 10, got %d", got(5))
	}
}

func TestProvide_Interface(t *testing.T) {
	c := &Container{}
	c.Add(Provide(any("interface value")))

	got, ok := Inject[any](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got != "interface value" {
		t.Errorf("expected 'interface value', got %v", got)
	}
}

func TestProvide_ByteSlice(t *testing.T) {
	c := &Container{}
	bs := []byte{0x01, 0x02, 0x03}
	c.Add(Provide(bs))

	got, ok := Inject[[]byte](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if len(got) != 3 || got[0] != 0x01 {
		t.Errorf("expected [0x01, 0x02, 0x03], got %v", got)
	}
}

func TestProvide_NestedStruct(t *testing.T) {
	type Inner struct {
		Value string
	}
	type Outer struct {
		Inner Inner
		Count int
	}

	c := &Container{}
	outer := Outer{
		Inner: Inner{Value: "nested"},
		Count: 42,
	}
	c.Add(Provide(outer))

	got, ok := Inject[Outer](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got.Inner.Value != "nested" || got.Count != 42 {
		t.Errorf("expected nested struct, got %v", got)
	}
}

func TestProvide_EmptyStruct(t *testing.T) {
	type Empty struct{}

	c := &Container{}
	c.Add(Provide(Empty{}))

	got, ok := Inject[Empty](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	_ = got
}

func TestProvide_StructWithSlice(t *testing.T) {
	type Data struct {
		Items []string
	}

	c := &Container{}
	c.Add(Provide(Data{Items: []string{"a", "b", "c"}}))

	got, ok := Inject[Data](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if len(got.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(got.Items))
	}
}

func TestProvide_StructWithMap(t *testing.T) {
	type Data struct {
		Config map[string]int
	}

	c := &Container{}
	c.Add(Provide(Data{Config: map[string]int{"a": 1, "b": 2}}))

	got, ok := Inject[Data](c)
	if !ok {
		t.Fatal("expected Inject to return true")
	}
	if got.Config["a"] != 1 || got.Config["b"] != 2 {
		t.Errorf("expected map values, got %v", got.Config)
	}
}

func TestInjectTo_InterfaceType(t *testing.T) {
	c := &Container{}
	c.Add(Provide(any("test value")))

	var v any
	err := InjectTo(&v, c)
	if err != nil {
		t.Fatal("expected InjectTo to succeed")
	}
	if v != "test value" {
		t.Errorf("expected 'test value', got %v", v)
	}
}

func TestInjectTo_Success(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "test"}))

	var db Database
	err := InjectTo(&db, c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.DSN != "test" {
		t.Errorf("expected 'test', got %s", db.DSN)
	}
}

func TestMustInjectTo_Success(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Config{AppName: "test"}))

	var cfg Config
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	MustInjectTo(&cfg, c)

	if cfg.AppName != "test" {
		t.Errorf("expected 'test', got %s", cfg.AppName)
	}
}

func TestContainer_AddReturnsFalseForDuplicate(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "test"}))

	success := c.Add(Provide(Database{DSN: "test2"}))
	if success {
		t.Fatal("expected Add to return false for duplicate")
	}
}

func TestProvider_Inject_WrongType(t *testing.T) {
	p := Provide(Database{DSN: "test"})

	var cfg Config
	ok := p.Inject(&cfg)
	if ok {
		t.Fatal("expected Inject to return false for wrong type")
	}
}

func TestProvider_Inject_NilPointer(t *testing.T) {
	p := Provide(Database{DSN: "test"})

	ok := p.Inject(nil)
	if ok {
		t.Fatal("expected Inject to return false for nil pointer")
	}
}

func TestProvider_ID_IsUnique(t *testing.T) {
	p1 := Provide(Database{DSN: "test1"})
	p2 := Provide(Database{DSN: "test2"})
	p3 := Provide(Config{AppName: "test"})

	id1 := p1.ID()
	id2 := p2.ID()
	id3 := p3.ID()

	if id1 != id2 {
		t.Error("expected same ID for same type")
	}
	if id1 == id3 {
		t.Error("expected different ID for different type")
	}
}

func TestMultiContainerInjection(t *testing.T) {
	c1 := &Container{}
	c2 := &Container{}
	c3 := &Container{}

	c1.Add(Provide(Database{DSN: "db1"}))
	c2.Add(Provide(Config{AppName: "app2"}))
	c3.Add(Provide(Service{Name: "svc3"}))

	db, ok := Inject[Database](c1, c2, c3)
	if !ok || db.DSN != "db1" {
		t.Errorf("expected db1, got %v", db)
	}

	cfg, ok := Inject[Config](c1, c2, c3)
	if !ok || cfg.AppName != "app2" {
		t.Errorf("expected app2, got %v", cfg)
	}

	svc, ok := Inject[Service](c1, c2, c3)
	if !ok || svc.Name != "svc3" {
		t.Errorf("expected svc3, got %v", svc)
	}
}

func TestMultiContainerInjection_InjectTo(t *testing.T) {
	c1 := &Container{}
	c2 := &Container{}

	c1.Add(Provide(Database{DSN: "db1"}))
	c2.Add(Provide(Config{AppName: "app2"}))

	var db Database
	err := InjectTo(&db, c1, c2)
	if err != nil || db.DSN != "db1" {
		t.Errorf("expected db1, got %v", db)
	}

	var cfg Config
	err = InjectTo(&cfg, c1, c2)
	if err != nil || cfg.AppName != "app2" {
		t.Errorf("expected app2, got %v", cfg)
	}
}

func TestMultiContainerInjection_NotFound(t *testing.T) {
	c1 := &Container{}
	c2 := &Container{}

	c1.Add(Provide(Database{DSN: "db1"}))

	_, ok := Inject[Config](c1, c2)
	if ok {
		t.Error("expected Inject to return false for non-existent provider")
	}

	var cfg Config
	err := InjectTo(&cfg, c1, c2)
	if err == nil {
		t.Error("expected InjectTo to fail for non-existent provider")
	}
}

func ExampleMultiContainer() {
	c1 := &Container{}
	c2 := &Container{}

	c1.Add(Provide(Database{DSN: "mysql://localhost"}))
	c2.Add(Provide(Config{AppName: "multi-app"}))

	db, _ := Inject[Database](c1, c2)
	cfg, _ := Inject[Config](c1, c2)

	fmt.Printf("DB: %s, App: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost, App: multi-app
}

func TestInvoke(t *testing.T) {
	c := &Container{}

	type Config struct {
		String string
		Float  float64
		Int    int
	}

	c.Add(Lazy(func() (v Config, err error) {
		if err = InjectTo(&v.Float, c); err != nil {
			return
		}
		if err = InjectTo(&v.String, c); err != nil {
			return
		}
		if err = InjectTo(&v.Int, c); err != nil {
			return
		}
		return
	}))

	c.Add(Lazy(func() (v float64, err error) {
		r, err := ShouldInject[int](c)
		if err != nil {
			return
		}
		return float64(r), nil
	}))
	c.MustAdd(Lazy(func() (v string, err error) {
		return "test", nil
	}))

	c.Add(Lazy(func() (v int, err error) {
		return 5, nil
	}))

	v, err := ShouldInject[Config](c)
	t.Log(v, err)
}

func TestCircularDependency(t *testing.T) {
	c := &Container{}

	c.Add(Lazy(func() (v int, err error) {
		_, err = ShouldInject[string](c)
		if err != nil {
			return 0, err
		}
		return 1, nil
	}))

	c.Add(Lazy(func() (v string, err error) {
		_, err = ShouldInject[int](c)
		if err != nil {
			return "", err
		}
		return "ok", nil
	}))

	_, err := ShouldInject[int](c)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	t.Log("circular detected:", err)
}

type testLevel2 struct{ Value int }
type testLevel3 struct {
	L2    testLevel2
	Value string
}
type testConfig struct{ Name string }
type testService struct {
	Config testConfig
	Port   int
}
type testHandler struct {
	Service testService
	Path    string
}

func TestComplexDependencyScenarios(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) (any, error)
		validate func(any) error
		wantErr  bool
	}{
		{
			name: "MixedProvideAndLazy",
			setup: func() *Container {
				c := &Container{}
				c.Add(Provide("static-string"))
				c.Add(Lazy(func() (v int, err error) {
					s, err := ShouldInject[string](c)
					if err != nil {
						return 0, err
					}
					if s != "static-string" {
						return 0, fmt.Errorf("unexpected string: %s", s)
					}
					return 42, nil
				}))
				c.Add(Lazy(func() (v float64, err error) {
					i, err := ShouldInject[int](c)
					if err != nil {
						return 0, err
					}
					return float64(i) * 2, nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[float64](c)
			},
			validate: func(got any) error {
				v := got.(float64)
				if v != 84.0 {
					return fmt.Errorf("expected 84.0, got %f", v)
				}
				return nil
			},
		},
		{
			name: "LazyDependencyOrderIndependent",
			setup: func() *Container {
				c := &Container{}
				c.Add(Lazy(func() (v string, err error) {
					i, err := ShouldInject[int](c)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("int-%d", i), nil
				}))
				c.Add(Lazy(func() (v int, err error) {
					f, err := ShouldInject[float64](c)
					if err != nil {
						return 0, err
					}
					return int(f), nil
				}))
				c.Add(Lazy(func() (v float64, err error) {
					return 3.14, nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[string](c)
			},
			validate: func(got any) error {
				v := got.(string)
				if v != "int-3" {
					return fmt.Errorf("expected int-3, got %s", v)
				}
				return nil
			},
		},
		{
			name: "DeepDependencyChain",
			setup: func() *Container {
				c := &Container{}
				c.Add(Lazy(func() (v int, err error) { return 1, nil }))
				c.Add(Lazy(func() (v int, err error) {
					prev, err := ShouldInject[int](c)
					if err != nil {
						return 0, err
					}
					return prev + 1, nil
				}))
				c.Add(Lazy(func() (v testLevel2, err error) {
					i, err := ShouldInject[int](c)
					if err != nil {
						return testLevel2{}, err
					}
					return testLevel2{Value: i * 10}, nil
				}))
				c.Add(Lazy(func() (v testLevel3, err error) {
					l2, err := ShouldInject[testLevel2](c)
					if err != nil {
						return testLevel3{}, err
					}
					return testLevel3{L2: l2, Value: fmt.Sprintf("level3-%d", l2.Value)}, nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[testLevel3](c)
			},
			validate: func(got any) error {
				v := got.(testLevel3)
				if v.Value != "level3-10" {
					return fmt.Errorf("expected level3-10, got %s", v.Value)
				}
				return nil
			},
		},
		{
			name: "CrossDependencies",
			setup: func() *Container {
				c := &Container{}
				c.Add(Lazy(func() (v testConfig, err error) {
					return testConfig{Name: "my-app"}, nil
				}))
				c.Add(Lazy(func() (v int, err error) {
					cfg, err := ShouldInject[testConfig](c)
					if err != nil {
						return 0, err
					}
					if cfg.Name == "my-app" {
						return 8080, nil
					}
					return 9090, nil
				}))
				c.Add(Lazy(func() (v testService, err error) {
					cfg, err := ShouldInject[testConfig](c)
					if err != nil {
						return testService{}, err
					}
					port, err := ShouldInject[int](c)
					if err != nil {
						return testService{}, err
					}
					return testService{Config: cfg, Port: port}, nil
				}))
				c.Add(Lazy(func() (v testHandler, err error) {
					svc, err := ShouldInject[testService](c)
					if err != nil {
						return testHandler{}, err
					}
					return testHandler{Service: svc, Path: "/api"}, nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[testHandler](c)
			},
			validate: func(got any) error {
				v := got.(testHandler)
				if v.Service.Config.Name != "my-app" {
					return fmt.Errorf("expected my-app, got %s", v.Service.Config.Name)
				}
				if v.Service.Port != 8080 {
					return fmt.Errorf("expected 8080, got %d", v.Service.Port)
				}
				if v.Path != "/api" {
					return fmt.Errorf("expected /api, got %s", v.Path)
				}
				return nil
			},
		},
		{
			name: "MultiContainerMixedProviders",
			setup: func() *Container {
				c1, c2 := &Container{}, &Container{}
				c1.Add(Provide("from-c1"))
				c1.Add(Lazy(func() (v int, err error) {
					s, err := ShouldInject[string](c2)
					if err != nil {
						return 0, err
					}
					return len(s), nil
				}))
				c2.Add(Provide("hello"))
				c2.Add(Lazy(func() (v float64, err error) {
					i, err := ShouldInject[int](c1)
					if err != nil {
						return 0, err
					}
					return float64(i) * 1.5, nil
				}))
				return c2
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[float64](c)
			},
			validate: func(got any) error {
				v := got.(float64)
				if v != 7.5 {
					return fmt.Errorf("expected 7.5, got %f", v)
				}
				return nil
			},
		},
		{
			name: "ReverseOrderDependency",
			setup: func() *Container {
				c := &Container{}
				c.Add(Lazy(func() (v string, err error) {
					i, err := ShouldInject[int](c)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("got-%d", i), nil
				}))
				c.Add(Lazy(func() (v bool, err error) {
					s, err := ShouldInject[string](c)
					if err != nil {
						return false, err
					}
					return s == "got-42", nil
				}))
				c.Add(Lazy(func() (v int, err error) { return 42, nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[bool](c)
			},
			validate: func(got any) error {
				v := got.(bool)
				if !v {
					return fmt.Errorf("expected true")
				}
				return nil
			},
		},
		{
			name: "LazyWithErrorPropagation",
			setup: func() *Container {
				c := &Container{}
				c.Add(Lazy(func() (v int, err error) {
					return 0, fmt.Errorf("intentional error")
				}))
				c.Add(Lazy(func() (v string, err error) {
					_, err = ShouldInject[int](c)
					if err != nil {
						return "", fmt.Errorf("wrapped: %w", err)
					}
					return "ok", nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[string](c)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			got, err := tt.injectFn(c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && tt.validate != nil {
				if err := tt.validate(got); err != nil {
					t.Error(err)
				}
			}
		})
	}
}

type StringAlias string
type IntAlias int
type MyInt int

func TestTypeAliasesAndEquals(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) (any, error)
		validate func(any) error
		wantErr  bool
	}{
		{
			name: "TypeAlias",
			setup: func() *Container {
				c := &Container{}
				c.Add(Provide(StringAlias("alias-value")))
				c.Add(Lazy(func() (v IntAlias, err error) {
					s, err := ShouldInject[StringAlias](c)
					if err != nil {
						return 0, err
					}
					return IntAlias(len(s)), nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[IntAlias](c)
			},
			validate: func(got any) error {
				v := got.(IntAlias)
				if v != 11 {
					return fmt.Errorf("expected 11, got %d", v)
				}
				return nil
			},
		},
		{
			name: "TypeEquals_SameType",
			setup: func() *Container {
				c := &Container{}
				type MyInt = int
				c.Add(Provide(MyInt(100)))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[int](c)
			},
			validate: func(got any) error {
				v := got.(int)
				if v != 100 {
					return fmt.Errorf("expected 100, got %d", v)
				}
				return nil
			},
		},
		{
			name: "TypeAlias_WithDependency",
			setup: func() *Container {
				c := &Container{}
				type Name string
				type Age int
				type Person struct {
					Name Name
					Age  Age
				}
				c.Add(Provide(Name("Alice")))
				c.Add(Lazy(func() (v Age, err error) {
					return Age(30), nil
				}))
				c.Add(Lazy(func() (v Person, err error) {
					name, err := ShouldInject[Name](c)
					if err != nil {
						return Person{}, err
					}
					age, err := ShouldInject[Age](c)
					if err != nil {
						return Person{}, err
					}
					return Person{Name: name, Age: age}, nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[struct {
					Name string
					Age  int
				}](c)
			},
			wantErr: true,
		},
		{
			name: "SliceTypeAlias",
			setup: func() *Container {
				c := &Container{}
				type IntSlice []int
				c.Add(Provide(IntSlice{1, 2, 3}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[[]int](c)
			},
			wantErr: true,
		},
		{
			name: "StructTypeAlias",
			setup: func() *Container {
				c := &Container{}
				type MyStruct struct {
					Value int
				}
				c.Add(Provide(MyStruct{Value: 42}))
				c.Add(Lazy(func() (v string, err error) {
					s, err := ShouldInject[MyStruct](c)
					if err != nil {
						return "", err
					}
					return fmt.Sprintf("val-%d", s.Value), nil
				}))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[string](c)
			},
			validate: func(got any) error {
				v := got.(string)
				if v != "val-42" {
					return fmt.Errorf("expected val-42, got %s", v)
				}
				return nil
			},
		},
		{
			name: "PointerAlias",
			setup: func() *Container {
				c := &Container{}
				val := MyInt(999)
				c.Add(Provide(&val))
				return c
			},
			injectFn: func(c *Container) (any, error) {
				return ShouldInject[*MyInt](c)
			},
			validate: func(got any) error {
				v := got.(*MyInt)
				if *v != 999 {
					return fmt.Errorf("expected 999, got %d", *v)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			got, err := tt.injectFn(c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error: %v", err)
			}
			if !tt.wantErr && tt.validate != nil {
				if err := tt.validate(got); err != nil {
					t.Error(err)
				}
			}
		})
	}
}
