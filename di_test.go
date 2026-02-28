package godi

import (
	"fmt"
	"math/rand"
	"testing"
)

type Database struct{ DSN string }
type Config struct{ AppName string }
type Service struct {
	Name string
	DB   Database
	Cfg  Config
}

func TestProvide(t *testing.T) {
	db := Database{DSN: "mysql://localhost"}
	p := Provide(db)

	var got Database
	if err := p.Inject(nil, &got); err != nil {
		t.Fatal(err)
	}
	if got.DSN != db.DSN {
		t.Errorf("expected %s, got %s", db.DSN, got.DSN)
	}
	if _, ok := p.ID().(*Database); !ok {
		t.Errorf("expected *Database ID, got %T", p.ID())
	}
}

func TestContainer_AddAndInject(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost:3306/test"}),
		Provide(Config{AppName: "test-app"}),
	)

	gotDB, err := Inject[Database](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotDB.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", gotDB.DSN)
	}

	gotCfg, err := Inject[Config](c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotCfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", gotCfg.AppName)
	}
}

func TestContainer_Add(t *testing.T) {
	t.Run("unique", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	})
	t.Run("duplicate error", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		if err := c.Add(Provide(Database{DSN: "mysql://remote"})); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestInject_ServiceWithDependencies(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost:3306/test"}),
		Provide(Config{AppName: "test-app"}),
	)

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)

	svc := Service{Name: "my-service", DB: db, Cfg: cfg}
	if svc.DB.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", svc.DB.DSN)
	}
	if svc.Cfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", svc.Cfg.AppName)
	}
}

func TestInjectTo(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost:3306/test"}))

	var db Database
	if err := InjectTo(&db, c); err != nil {
		t.Fatal(err)
	}
	if db.DSN != "mysql://localhost:3306/test" {
		t.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
	}
}

func TestInjectTo_ErrorCases(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		var db Database
		if err := InjectTo(&db, &Container{}); err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("wrong type", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "test"}))
		var cfg Config
		if err := InjectTo(&cfg, c); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestMustInject(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("unexpected panic")
			}
		}()
		db := MustInject[Database](c)
		if db.DSN != "mysql://localhost" {
			t.Errorf("expected mysql://localhost, got %s", db.DSN)
		}
	})
	t.Run("panic on not found", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		MustInject[Database](&Container{})
	})
}

func TestMustInjectTo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		var db Database
		MustInjectTo(&db, c)
		if db.DSN != "mysql://localhost" {
			t.Errorf("expected mysql://localhost, got %s", db.DSN)
		}
	})
	t.Run("panic on not found", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		var db Database
		MustInjectTo(&db, &Container{})
	})
}

func TestContainer_MustAdd(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		c := &Container{}
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("unexpected panic")
			}
		}()
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	})
	t.Run("panic on duplicate", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic")
			}
		}()
		c.MustAdd(Provide(Database{DSN: "mysql://remote"}))
	})
}

func TestContainer_ProviderID(t *testing.T) {
	c := &Container{}
	p := Provide(Database{DSN: "test"})
	c.MustAdd(p)

	id := p.ID()
	if id == nil {
		t.Fatal("expected non-nil ID")
	}
}

func TestProvide_AllTypes(t *testing.T) {
	t.Run("primitives", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide("test"),
			Provide(42),
			Provide(int8(8)),
			Provide(int16(16)),
			Provide(int32(32)),
			Provide(int64(64)),
			Provide(uint(100)),
			Provide(uint8(8)),
			Provide(uint16(16)),
			Provide(uint32(32)),
			Provide(uint64(64)),
			Provide(float32(3.14)),
			Provide(3.14159),
			Provide(true),
		)
		if v, err := Inject[string](c); err != nil || v != "test" {
			t.Errorf("string: %v", err)
		}
		if v, err := Inject[int](c); err != nil || v != 42 {
			t.Errorf("int: %v", err)
		}
		if v, err := Inject[int8](c); err != nil || v != 8 {
			t.Errorf("int8: %v", err)
		}
		if v, err := Inject[int16](c); err != nil || v != 16 {
			t.Errorf("int16: %v", err)
		}
		if v, err := Inject[int64](c); err != nil || v != 64 {
			t.Errorf("int64: %v", err)
		}
		if v, err := Inject[uint](c); err != nil || v != 100 {
			t.Errorf("uint: %v", err)
		}
		if v, err := Inject[float32](c); err != nil || v != 3.14 {
			t.Errorf("float32: %v", err)
		}
		if v, err := Inject[float64](c); err != nil || v != 3.14159 {
			t.Errorf("float64: %v", err)
		}
		if v, err := Inject[bool](c); err != nil || !v {
			t.Errorf("bool: %v", err)
		}
	})

	t.Run("collections", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide([]string{"a", "b", "c"}),
			Provide([]int{1, 2, 3}),
			Provide([]byte{0x01, 0x02, 0x03}),
			Provide(map[string]int{"a": 1, "b": 2}),
			Provide([3]int{1, 2, 3}),
		)
		if v, err := Inject[[]string](c); err != nil || len(v) != 3 {
			t.Errorf("[]string: %v", err)
		}
		if v, err := Inject[[]int](c); err != nil || len(v) != 3 {
			t.Errorf("[]int: %v", err)
		}
		if v, err := Inject[[]byte](c); err != nil || len(v) != 3 {
			t.Errorf("[]byte: %v", err)
		}
		if v, err := Inject[map[string]int](c); err != nil || v["a"] != 1 {
			t.Errorf("map: %v", err)
		}
		if v, err := Inject[[3]int](c); err != nil || v[0] != 1 {
			t.Errorf("[3]int: %v", err)
		}
	})

	t.Run("advanced", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide(&struct{ Name string }{Name: "Alice"}),
			Provide(make(chan int)),
			Provide(func() string { return "hello" }),
			Provide(any("interface")),
		)
		if v, err := Inject[*struct{ Name string }](c); err != nil || v.Name != "Alice" {
			t.Errorf("*struct: %v", err)
		}
		if v, err := Inject[chan int](c); err != nil || v == nil {
			t.Errorf("chan: %v", err)
		}
		if v, err := Inject[func() string](c); err != nil || v() != "hello" {
			t.Errorf("func: %v", err)
		}
		if v, err := Inject[any](c); err != nil || v != "interface" {
			t.Errorf("any: %v", err)
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
	for _, count := range []int{5, 8, 10} {
		t.Run(fmt.Sprintf("Count%d", count), func(t *testing.T) {
			for iter := 0; iter < 20; iter++ {
				c := &Container{}
				order := rand.Perm(count)
				for _, idx := range order {
					switch idx {
					case 0:
						c.MustAdd(Lazy(func(c *Container) (L0, error) { return 1, nil }))
					case 1:
						c.MustAdd(Lazy(func(c *Container) (L1, error) {
							v, _ := Inject[L0](c)
							return L1(v) + 1, nil
						}))
					case 2:
						c.MustAdd(Lazy(func(c *Container) (L2, error) {
							v, _ := Inject[L1](c)
							return L2(v) + 1, nil
						}))
					case 3:
						c.MustAdd(Lazy(func(c *Container) (L3, error) {
							v, _ := Inject[L2](c)
							return L3(v) + 1, nil
						}))
					case 4:
						c.MustAdd(Lazy(func(c *Container) (L4, error) {
							v, _ := Inject[L3](c)
							return L4(v) + 1, nil
						}))
					case 5:
						c.MustAdd(Lazy(func(c *Container) (L5, error) {
							v, _ := Inject[L4](c)
							return L5(v) + 1, nil
						}))
					case 6:
						c.MustAdd(Lazy(func(c *Container) (L6, error) {
							v, _ := Inject[L5](c)
							return L6(v) + 1, nil
						}))
					case 7:
						c.MustAdd(Lazy(func(c *Container) (L7, error) {
							v, _ := Inject[L6](c)
							return L7(v) + 1, nil
						}))
					case 8:
						c.MustAdd(Lazy(func(c *Container) (L8, error) {
							v, _ := Inject[L7](c)
							return L8(v) + 1, nil
						}))
					case 9:
						c.MustAdd(Lazy(func(c *Container) (L9, error) {
							v, _ := Inject[L8](c)
							return L9(v) + 1, nil
						}))
					}
				}
				if _, err := Inject[L4](c); err != nil {
					t.Fatalf("iter %d: %v", iter, err)
				}
			}
		})
	}
}

func TestLazyLargeDependencyGraph(t *testing.T) {
	for _, tt := range []struct {
		name string
	}{
		{"LinearChain"},
		{"BinaryTree"},
		{"Diamond"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for run := 0; run < 10; run++ {
				c := &Container{}
				c.MustAdd(Lazy(func(c *Container) (L0, error) { return 100, nil }))
				c.MustAdd(Lazy(func(c *Container) (L1, error) {
					v, _ := Inject[L0](c)
					return L1(v) + 10, nil
				}))
				c.MustAdd(Lazy(func(c *Container) (L2, error) {
					v, _ := Inject[L0](c)
					return L2(v) + 20, nil
				}))
				c.MustAdd(Lazy(func(c *Container) (L3, error) {
					v1, _ := Inject[L1](c)
					v2, _ := Inject[L2](c)
					return L3(int(v1) + int(v2) + 30), nil
				}))
				c.MustAdd(Lazy(func(c *Container) (L4, error) {
					v, _ := Inject[L3](c)
					return L4(v) + 40, nil
				}))
				if _, err := Inject[L4](c); err != nil {
					t.Fatalf("run %d: %v", run, err)
				}
			}
		})
	}
}

func TestCircularDependency(t *testing.T) {
	c := &Container{}
	c.MustAdd(Lazy(func(c *Container) (int, error) {
		_, err := Inject[string](c)
		return 0, err
	}))
	c.MustAdd(Lazy(func(c *Container) (string, error) {
		_, err := Inject[int](c)
		return "", err
	}))
	if _, err := Inject[int](c); err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestMultiContainer(t *testing.T) {
	c1, c2, c3 := &Container{}, &Container{}, &Container{}
	c1.MustAdd(Provide(Database{DSN: "db1"}))
	c2.MustAdd(Provide(Config{AppName: "app2"}))
	c3.MustAdd(Provide(Service{Name: "svc3"}))

	db, err := Inject[Database](c1, c2, c3)
	if err != nil || db.DSN != "db1" {
		t.Errorf("expected db1, got %v", db)
	}

	cfg, err := Inject[Config](c1, c2, c3)
	if err != nil || cfg.AppName != "app2" {
		t.Errorf("expected app2, got %v", cfg)
	}

	svc, err := Inject[Service](c1, c2, c3)
	if err != nil || svc.Name != "svc3" {
		t.Errorf("expected svc3, got %v", svc)
	}
}

func TestMultiContainer_InjectTo(t *testing.T) {
	c1, c2 := &Container{}, &Container{}
	c1.MustAdd(Provide(Database{DSN: "db1"}))
	c2.MustAdd(Provide(Config{AppName: "app2"}))

	var db Database
	if err := InjectTo(&db, c1, c2); err != nil || db.DSN != "db1" {
		t.Errorf("expected db1, got %v", db)
	}

	var cfg Config
	if err := InjectTo(&cfg, c1, c2); err != nil || cfg.AppName != "app2" {
		t.Errorf("expected app2, got %v", cfg)
	}
}

func TestMultiContainer_NotFound(t *testing.T) {
	c1, c2 := &Container{}, &Container{}
	c1.Add(Provide(Database{DSN: "db1"}))

	if _, err := Inject[Config](c1, c2); err == nil {
		t.Error("expected error")
	}

	var cfg Config
	if err := InjectTo(&cfg, c1, c2); err == nil {
		t.Error("expected error")
	}
}

func TestLazyWithError(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Lazy(func(c *Container) (int, error) {
			return 0, fmt.Errorf("intentional error")
		}),
		Lazy(func(c *Container) (string, error) {
			_, err := Inject[int](c)
			if err != nil {
				return "", fmt.Errorf("wrapped: %w", err)
			}
			return "ok", nil
		}),
	)
	if _, err := Inject[string](c); err == nil {
		t.Fatal("expected error")
	}
}

func TestTypeAliases(t *testing.T) {
	type StringAlias string
	type IntAlias int

	c := &Container{}
	c.MustAdd(
		Provide(StringAlias("alias-value")),
		Lazy(func(c *Container) (IntAlias, error) {
			s, _ := Inject[StringAlias](c)
			return IntAlias(len(s)), nil
		}),
	)

	v, err := Inject[IntAlias](c)
	if err != nil || v != 11 {
		t.Errorf("expected 11, got %d", v)
	}
}

func TestComplexScenarios(t *testing.T) {
	t.Run("MixedProvideAndLazy", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide("static"),
			Lazy(func(c *Container) (int, error) {
				s, _ := Inject[string](c)
				if s != "static" {
					return 0, fmt.Errorf("unexpected: %s", s)
				}
				return 42, nil
			}),
		)
		v, err := Inject[int](c)
		if err != nil || v != 42 {
			t.Errorf("expected 42, got %v", v)
		}
	})

	t.Run("CrossDependencies", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Lazy(func(c *Container) (string, error) {
				i, _ := Inject[int](c)
				return fmt.Sprintf("got-%d", i), nil
			}),
			Lazy(func(c *Container) (int, error) { return 42, nil }),
		)
		v, err := Inject[string](c)
		if err != nil || v != "got-42" {
			t.Errorf("expected got-42, got %v", v)
		}
	})

	t.Run("DeepChain", func(t *testing.T) {
		type L1 int
		type L2 int
		c := &Container{}
		c.MustAdd(
			Lazy(func(c *Container) (int, error) { return 1, nil }),
			Lazy(func(c *Container) (L1, error) {
				v, _ := Inject[int](c)
				return L1(v + 1), nil
			}),
			Lazy(func(c *Container) (L2, error) {
				v, _ := Inject[L1](c)
				return L2(v * 10), nil
			}),
		)
		v, err := Inject[L2](c)
		if err != nil || v != 20 {
			t.Errorf("expected 20, got %v", v)
		}
	})

	t.Run("ChainMultiContainer", func(t *testing.T) {
		type Name string
		type Age int

		c1, c2 := &Container{}, &Container{}
		c1.MustAdd(Provide(Name("Alice")))
		c2.MustAdd(
			Provide(Name("Bob")),
			Chain(func(n Name) (Age, error) {
				return Age(len(n)), nil
			}),
		)

		age, err := Inject[Age](c2)
		if err != nil || age != 3 {
			t.Errorf("expected 3, got %v", age)
		}
	})

	t.Run("ChainErrorPropagation", func(t *testing.T) {
		type Input string
		type Output string

		c := &Container{}
		c.MustAdd(
			Provide(Input("test")),
			Chain(func(s Input) (Output, error) {
				return Output(s + "-ok"), nil
			}),
		)

		out, err := Inject[Output](c)
		if err != nil || out != "test-ok" {
			t.Errorf("expected test-ok, got %v", out)
		}
	})
}

func ExampleProvide() {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "my-app"}),
	)

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)

	fmt.Printf("DB: %s, App: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost, App: my-app
}

func ExampleContainer_MustAdd() {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "test"}),
	)

	db, _ := Inject[Database](c)
	cfg, _ := Inject[Config](c)

	fmt.Printf("DB: %s, Config: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost, Config: test
}

func ExampleMustInject() {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))

	db := MustInject[Database](c)
	fmt.Printf("DB: %s\n", db.DSN)
	// Output: DB: mysql://localhost
}

func ExampleInject() {
	c := &Container{}
	c.MustAdd(Provide(Config{AppName: "my-app"}))

	cfg, err := Inject[Config](c)
	fmt.Printf("App: %s, Error: %v\n", cfg.AppName, err)
	// Output: App: my-app, Error: <nil>
}

func ExampleInjectTo() {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))

	var db Database
	err := InjectTo(&db, c)
	fmt.Printf("Injected: %v, DSN: %s\n", err == nil, db.DSN)
	// Output: Injected: true, DSN: mysql://localhost
}

func ExampleMultiContainer() {
	c1, c2 := &Container{}, &Container{}
	c1.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	c2.MustAdd(Provide(Config{AppName: "multi-app"}))

	db, _ := Inject[Database](c1, c2)
	cfg, _ := Inject[Config](c1, c2)

	fmt.Printf("DB: %s, App: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost, App: multi-app
}
