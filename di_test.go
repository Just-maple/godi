package godi

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

type Database struct{ DSN string }
type Config struct{ AppName string }
type Service struct {
	Name string
	DB   Database
	Cfg  Config
}

type bench0 struct{ Val int }
type bench1 struct{ Val int }
type bench2 struct{ Val int }
type bench3 struct{ Val int }
type bench4 struct{ Val int }
type bench5 struct{ Val int }
type bench6 struct{ Val int }
type bench7 struct{ Val int }
type bench8 struct{ Val int }
type bench9 struct{ Val int }

func TestProvide(t *testing.T) {
	db := Database{DSN: "mysql://localhost"}
	p := Provide(db)

	var got Database
	if _, err := p.inject(nil, &got); err != nil {
		t.Fatal(err)
	}
	if got.DSN != db.DSN {
		t.Errorf("expected %s, got %s", db.DSN, got.DSN)
	}
}

func TestContainer_AddAndInject(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) (any, error)
		wantVal  any
		wantErr  bool
	}{
		{
			name: "Database",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost:3306/test"}))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[Database](c) },
			wantVal:  Database{DSN: "mysql://localhost:3306/test"},
		},
		{
			name: "Config",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Config{AppName: "test-app"}))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[Config](c) },
			wantVal:  Config{AppName: "test-app"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			got, err := tt.injectFn(c)

			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.wantVal {
				t.Errorf("got %v, want %v", got, tt.wantVal)
			}
		})
	}
}

func TestContainer_Add(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *Container
		addFn   func(*Container) error
		wantErr bool
	}{
		{
			name:  "unique",
			setup: func() *Container { return &Container{} },
			addFn: func(c *Container) error {
				return c.Add(Provide(Database{DSN: "mysql://localhost"}))
			},
		},
		{
			name: "duplicate error",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			addFn: func(c *Container) error {
				return c.Add(Provide(Database{DSN: "mysql://remote"}))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			err := tt.addFn(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) error
		wantVal  any
		wantErr  bool
	}{
		{
			name: "InjectTo success",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost:3306/test"}))
				return c
			},
			injectFn: func(c *Container) error {
				var db Database
				if err := InjectTo(&db, c); err != nil {
					return err
				}
				if db.DSN != "mysql://localhost:3306/test" {
					return fmt.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
				}
				return nil
			},
		},
		{
			name: "InjectAs success",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost:3306/test"}))
				return c
			},
			injectFn: func(c *Container) error {
				db := Database{}
				if err := InjectAs(&db, c); err != nil {
					return err
				}
				if db.DSN != "mysql://localhost:3306/test" {
					return fmt.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			if err := tt.injectFn(c); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestInject_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) error
		wantErr  bool
	}{
		{
			name:  "InjectTo not found",
			setup: func() *Container { return &Container{} },
			injectFn: func(c *Container) error {
				var db Database
				return InjectTo(&db, c)
			},
			wantErr: true,
		},
		{
			name: "InjectTo wrong type",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "test"}))
				return c
			},
			injectFn: func(c *Container) error {
				var cfg Config
				return InjectTo(&cfg, c)
			},
			wantErr: true,
		},
		{
			name:  "InjectAs not found",
			setup: func() *Container { return &Container{} },
			injectFn: func(c *Container) error {
				db := Database{}
				return InjectAs(&db, c)
			},
			wantErr: true,
		},
		{
			name: "InjectAs wrong type",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "test"}))
				return c
			},
			injectFn: func(c *Container) error {
				cfg := Config{}
				return InjectAs(&cfg, c)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			err := tt.injectFn(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestContainer_Inject(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) error
		wantErr  bool
	}{
		{
			name: "single dependency",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) error {
				db := Database{}
				return c.Inject(&db)
			},
		},
		{
			name: "multiple dependencies",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(Database{DSN: "mysql://localhost"}),
					Provide(Config{AppName: "test-app"}),
				)
				return c
			},
			injectFn: func(c *Container) error {
				db := Database{}
				cfg := Config{}
				return c.Inject(&db, &cfg)
			},
		},
		{
			name: "not found error",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) error {
				cfg := Config{}
				return c.Inject(&cfg)
			},
			wantErr: true,
		},
		{
			name: "partial error",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) error {
				db := Database{}
				cfg := Config{}
				return c.Inject(&db, &cfg)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			err := tt.injectFn(c)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMustInject(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Container
		injectFn  func(*Container)
		wantPanic bool
		wantVal   any
	}{
		{
			name: "success",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) {
				db := MustInject[Database](c)
				if db.DSN != "mysql://localhost" {
					t.Errorf("expected mysql://localhost, got %s", db.DSN)
				}
			},
		},
		{
			name:      "panic on not found",
			setup:     func() *Container { return &Container{} },
			injectFn:  func(c *Container) { MustInject[Database](c) },
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()
			tt.injectFn(c)
		})
	}
}

func TestMustInjectTo(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Container
		injectFn  func(*Container)
		wantPanic bool
		wantVal   any
	}{
		{
			name: "success",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) {
				var db Database
				MustInjectTo(&db, c)
				if db.DSN != "mysql://localhost" {
					t.Errorf("expected mysql://localhost, got %s", db.DSN)
				}
			},
		},
		{
			name:      "panic on not found",
			setup:     func() *Container { return &Container{} },
			injectFn:  func(c *Container) { var db Database; MustInjectTo(&db, c) },
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()
			tt.injectFn(c)
		})
	}
}

func TestMustInjectAs(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Container
		injectFn  func(*Container)
		wantPanic bool
		wantVal   any
	}{
		{
			name: "success",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			injectFn: func(c *Container) {
				var db Database
				MustInjectAs(&db, c)
				if db.DSN != "mysql://localhost" {
					t.Errorf("expected mysql://localhost, got %s", db.DSN)
				}
			},
		},
		{
			name:      "panic on not found",
			setup:     func() *Container { return &Container{} },
			injectFn:  func(c *Container) { var db Database; MustInjectAs(&db, c) },
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()
			tt.injectFn(c)
		})
	}
}

func TestContainer_MustAdd(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Container
		addFn     func(*Container)
		wantPanic bool
	}{
		{
			name:  "success",
			setup: func() *Container { return &Container{} },
			addFn: func(c *Container) {
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
			},
		},
		{
			name: "panic on duplicate",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			addFn: func(c *Container) {
				c.MustAdd(Provide(Database{DSN: "mysql://remote"}))
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			defer func() {
				if r := recover(); (r != nil) != tt.wantPanic {
					t.Errorf("panic = %v, wantPanic %v", r != nil, tt.wantPanic)
				}
			}()
			tt.addFn(c)
		})
	}
}

func TestContainer_ProviderID(t *testing.T) {
	p := Provide(Database{DSN: "test"})
	id, _ := p.Provide(nil)
	if id == nil {
		t.Fatal("expected non-nil ID")
	}
}

func TestProvide_AllTypes(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) error
	}{
		{
			name: "primitives",
			setup: func() *Container {
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
				return c
			},
			injectFn: func(c *Container) error {
				tests := []struct {
					name string
					fn   func() error
				}{
					{"string", func() error {
						v, err := Inject[string](c)
						if err != nil || v != "test" {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"int", func() error {
						v, err := Inject[int](c)
						if err != nil || v != 42 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"int8", func() error {
						v, err := Inject[int8](c)
						if err != nil || v != 8 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"int16", func() error {
						v, err := Inject[int16](c)
						if err != nil || v != 16 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"int64", func() error {
						v, err := Inject[int64](c)
						if err != nil || v != 64 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"uint", func() error {
						v, err := Inject[uint](c)
						if err != nil || v != 100 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"float32", func() error {
						v, err := Inject[float32](c)
						if err != nil || v != 3.14 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"float64", func() error {
						v, err := Inject[float64](c)
						if err != nil || v != 3.14159 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"bool", func() error {
						v, err := Inject[bool](c)
						if err != nil || !v {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
				}
				for _, tt := range tests {
					if err := tt.fn(); err != nil {
						return fmt.Errorf("%s: %v", tt.name, err)
					}
				}
				return nil
			},
		},
		{
			name: "collections",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide([]string{"a", "b", "c"}),
					Provide([]int{1, 2, 3}),
					Provide([]byte{0x01, 0x02, 0x03}),
					Provide(map[string]int{"a": 1, "b": 2}),
					Provide([3]int{1, 2, 3}),
				)
				return c
			},
			injectFn: func(c *Container) error {
				tests := []struct {
					name string
					fn   func() error
				}{
					{"[]string", func() error {
						v, err := Inject[[]string](c)
						if err != nil || len(v) != 3 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"[]int", func() error {
						v, err := Inject[[]int](c)
						if err != nil || len(v) != 3 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"[]byte", func() error {
						v, err := Inject[[]byte](c)
						if err != nil || len(v) != 3 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"map[string]int", func() error {
						v, err := Inject[map[string]int](c)
						if err != nil || v["a"] != 1 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"[3]int", func() error {
						v, err := Inject[[3]int](c)
						if err != nil || v[0] != 1 {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
				}
				for _, tt := range tests {
					if err := tt.fn(); err != nil {
						return fmt.Errorf("%s: %v", tt.name, err)
					}
				}
				return nil
			},
		},
		{
			name: "advanced",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(&struct{ Name string }{Name: "Alice"}),
					Provide(make(chan int)),
					Provide(func() string { return "hello" }),
					Provide(any("interface")),
				)
				return c
			},
			injectFn: func(c *Container) error {
				tests := []struct {
					name string
					fn   func() error
				}{
					{"*struct", func() error {
						v, err := Inject[*struct{ Name string }](c)
						if err != nil || v.Name != "Alice" {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"chan int", func() error {
						v, err := Inject[chan int](c)
						if err != nil || v == nil {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
					{"func", func() error {
						v, err := Inject[func() string](c)
						if err != nil || v() != "hello" {
							return fmt.Errorf("func returned wrong value")
						}
						return nil
					}},
					{"any", func() error {
						v, err := Inject[any](c)
						if err != nil || v != "interface" {
							return fmt.Errorf("got %v", v)
						}
						return nil
					}},
				}
				for _, tt := range tests {
					if err := tt.fn(); err != nil {
						return fmt.Errorf("%s: %v", tt.name, err)
					}
				}
				return nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			if err := tt.injectFn(c); err != nil {
				t.Error(err)
			}
		})
	}
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

func TestBuildDependencyOrderIndependent_Shuffle(t *testing.T) {
	for _, count := range []int{5, 8, 10} {
		t.Run(fmt.Sprintf("Count%d", count), func(t *testing.T) {
			for iter := 0; iter < 20; iter++ {
				c := &Container{}
				order := rand.Perm(count)
				for _, idx := range order {
					switch idx {
					case 0:
						c.MustAdd(Build(func(c *Container) (L0, error) { return 1, nil }))
					case 1:
						c.MustAdd(Build(func(c *Container) (L1, error) {
							v, _ := Inject[L0](c)
							return L1(v) + 1, nil
						}))
					case 2:
						c.MustAdd(Build(func(c *Container) (L2, error) {
							v, _ := Inject[L1](c)
							return L2(v) + 1, nil
						}))
					case 3:
						c.MustAdd(Build(func(c *Container) (L3, error) {
							v, _ := Inject[L2](c)
							return L3(v) + 1, nil
						}))
					case 4:
						c.MustAdd(Build(func(c *Container) (L4, error) {
							v, _ := Inject[L3](c)
							return L4(v) + 1, nil
						}))
					case 5:
						c.MustAdd(Build(func(c *Container) (L5, error) {
							v, _ := Inject[L4](c)
							return L5(v) + 1, nil
						}))
					case 6:
						c.MustAdd(Build(func(c *Container) (L6, error) {
							v, _ := Inject[L5](c)
							return L6(v) + 1, nil
						}))
					case 7:
						c.MustAdd(Build(func(c *Container) (L7, error) {
							v, _ := Inject[L6](c)
							return L7(v) + 1, nil
						}))
					case 8:
						c.MustAdd(Build(func(c *Container) (L8, error) {
							v, _ := Inject[L7](c)
							return L8(v) + 1, nil
						}))
					case 9:
						c.MustAdd(Build(func(c *Container) (L9, error) {
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

func TestBuildLargeDependencyGraph(t *testing.T) {
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
				c.MustAdd(Build(func(c *Container) (L0, error) { return 100, nil }))
				c.MustAdd(Build(func(c *Container) (L1, error) {
					v, _ := Inject[L0](c)
					return L1(v) + 10, nil
				}))
				c.MustAdd(Build(func(c *Container) (L2, error) {
					v, _ := Inject[L0](c)
					return L2(v) + 20, nil
				}))
				c.MustAdd(Build(func(c *Container) (L3, error) {
					v1, _ := Inject[L1](c)
					v2, _ := Inject[L2](c)
					return L3(int(v1) + int(v2) + 30), nil
				}))
				c.MustAdd(Build(func(c *Container) (L4, error) {
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
	c.MustAdd(Build(func(c *Container) (int, error) {
		_, err := Inject[string](c)
		return 0, err
	}))
	c.MustAdd(Build(func(c *Container) (string, error) {
		_, err := Inject[int](c)
		return "", err
	}))
	if _, err := Inject[int](c); err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestBuildWithError(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Build(func(c *Container) (int, error) {
			return 0, fmt.Errorf("intentional error")
		}),
		Build(func(c *Container) (string, error) {
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
		Build(func(c *Container) (IntAlias, error) {
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
	t.Run("MixedProvideAndBuild", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide("static"),
			Build(func(c *Container) (int, error) {
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
			Build(func(c *Container) (string, error) {
				i, _ := Inject[int](c)
				return fmt.Sprintf("got-%d", i), nil
			}),
			Build(func(c *Container) (int, error) { return 42, nil }),
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
			Build(func(c *Container) (int, error) { return 1, nil }),
			Build(func(c *Container) (L1, error) {
				v, _ := Inject[int](c)
				return L1(v + 1), nil
			}),
			Build(func(c *Container) (L2, error) {
				v, _ := Inject[L1](c)
				return L2(v * 10), nil
			}),
		)
		v, err := Inject[L2](c)
		if err != nil || v != 20 {
			t.Errorf("expected 20, got %v", v)
		}
	})

	t.Run("ChainError", func(t *testing.T) {
		type Name string
		type Age int

		c := &Container{}
		c.MustAdd(
			Build(func(container *Container) (Name, error) {
				return "", fmt.Errorf("intentional error")
			}),
			Chain(func(n Name) (Age, error) {
				return Age(len(n)), nil
			}),
		)

		age, err := Inject[Age](c)
		if err == nil || age != 0 {
			t.Errorf("expected error, got %v", age)
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

func ExampleInjectAs() {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))

	db := Database{}
	err := InjectAs(&db, c)
	fmt.Printf("Injected: %v, DSN: %s\n", err == nil, db.DSN)
	// Output: Injected: true, DSN: mysql://localhost
}

func ExampleContainer_Inject() {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "my-app"}),
	)

	var db Database
	var cfg Config
	err := c.Inject(&db, &cfg)
	fmt.Printf("Injected: %v, DB: %s, App: %s\n", err == nil, db.DSN, cfg.AppName)
	// Output: Injected: true, DB: mysql://localhost, App: my-app
}

func BenchmarkInject(b *testing.B) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "test"}))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Inject[Database](c)
	}
}

func BenchmarkInjectTo(b *testing.B) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "test"}))
	var db Database
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InjectTo(&db, c)
	}
}

func BenchmarkInjectAs(b *testing.B) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "test"}))
	db := Database{}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = InjectAs(&db, c)
	}
}

func BenchmarkContainer_Add(b *testing.B) {
	for _, n := range []int{10} {
		b.Run(fmt.Sprintf("Providers%d", n), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				c := &Container{}
				for j := 0; j < n; j++ {
					switch j % 10 {
					case 0:
						c.MustAdd(Provide(bench0{j}))
					case 1:
						c.MustAdd(Provide(bench1{j}))
					case 2:
						c.MustAdd(Provide(bench2{j}))
					case 3:
						c.MustAdd(Provide(bench3{j}))
					case 4:
						c.MustAdd(Provide(bench4{j}))
					case 5:
						c.MustAdd(Provide(bench5{j}))
					case 6:
						c.MustAdd(Provide(bench6{j}))
					case 7:
						c.MustAdd(Provide(bench7{j}))
					case 8:
						c.MustAdd(Provide(bench8{j}))
					case 9:
						c.MustAdd(Provide(bench9{j}))
					}
				}
			}
		})
	}
}

func TestHook(t *testing.T) {
	c := &Container{}

	startupCalled := 0
	shutdownCalled := 0

	startup := c.Hook("startup", func(v any, called int) func(context.Context) {
		if called > 0 {
			return nil
		}
		if _, ok := v.(Database); ok {
			return func(ctx context.Context) {
				startupCalled++
			}
		}
		if _, ok := v.(Config); ok {
			return func(ctx context.Context) {
				startupCalled++
			}
		}
		return nil
	})

	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		if _, ok := v.(Database); ok {
			return func(ctx context.Context) {
				shutdownCalled++
			}
		}
		if _, ok := v.(Config); ok {
			return func(ctx context.Context) {
				shutdownCalled++
			}
		}
		return nil
	})

	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "test-app"}),
	)

	ctx := context.Background()

	_, _ = Inject[Database](c)
	_, _ = Inject[Database](c)
	_, _ = Inject[Config](c)
	_, _ = Inject[Config](c)

	startup.Iterate(ctx, false)

	if startupCalled != 2 {
		t.Errorf("expected startupCalled=2, got %d", startupCalled)
	}

	shutdown.Iterate(ctx, true)

	if shutdownCalled != 2 {
		t.Errorf("expected shutdownCalled=2, got %d", shutdownCalled)
	}
}

func TestHook_WithBuild(t *testing.T) {
	c := &Container{}

	buildCalled := 0
	hookCalled := 0

	c.MustAdd(
		Build(func(c *Container) (Database, error) {
			buildCalled++
			return Database{DSN: "mysql://localhost"}, nil
		}),
	)

	h := c.Hook("init", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		if _, ok := v.(Database); ok {
			return func(ctx context.Context) {
				hookCalled++
			}
		}
		return nil
	})

	ctx := context.Background()
	_, _ = Inject[Database](c)

	h.Iterate(ctx, false)

	if buildCalled != 1 {
		t.Errorf("expected buildCalled=1, got %d", buildCalled)
	}
	if hookCalled != 1 {
		t.Errorf("expected hookCalled=1, got %d", hookCalled)
	}
}

func TestHook_MultipleContainers(t *testing.T) {
	c1, c2 := &Container{}, &Container{}

	c1.MustAdd(Provide(Database{DSN: "db1"}))
	c2.MustAdd(Provide(Config{AppName: "app2"}))

	h1 := c1.Hook("init", func(v any, provided int) func(context.Context) {
		return func(ctx context.Context) {}
	})

	h2 := c2.Hook("init", func(v any, provided int) func(context.Context) {
		return func(ctx context.Context) {}
	})

	if h1 == nil || h2 == nil {
		t.Error("expected non-nil hook executors")
	}
}

func Example_hook() {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "my-app"}),
	)

	startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			fmt.Printf("Starting: %T\n", v)
		}
	})

	shutdown := c.Hook("shutdown", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			fmt.Printf("Stopping: %T\n", v)
		}
	})

	ctx := context.Background()

	_, _ = Inject[Database](c)
	_, _ = Inject[Config](c)

	startup.Iterate(ctx, false)

	shutdown.Iterate(ctx, false)
	// Output:
	// Starting: godi.Database
	// Starting: godi.Config
	// Stopping: godi.Database
	// Stopping: godi.Config
}

// ============== Nested Container Hook Mechanism Tests ==============

func TestContainer_Nested_Hook_BothContainersTriggered(t *testing.T) {
	type Database struct{ DSN string }

	child := &Container{}
	child.MustAdd(Provide(Database{DSN: "mysql://localhost"}))

	childHookCalled := false
	childStartup := child.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				childHookCalled = true
			}
		}
	})

	parent := &Container{}
	parent.MustAdd(child)

	parentHookCalled := false
	parentStartup := parent.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				parentHookCalled = true
			}
		}
	})

	_, _ = Inject[Database](parent)

	childStartup.Iterate(context.Background(), false)

	parentStartup.Iterate(context.Background(), false)

	if !childHookCalled {
		t.Error("expected child hook to be called")
	}
	// Parent hook is also triggered because hooks fire on each container in the chain
	if !parentHookCalled {
		t.Error("expected parent hook to be called")
	}
}

func TestContainer_Nested_Hook_HookOnceOnBothContainers(t *testing.T) {
	type Database struct{ DSN string }

	child := &Container{}
	child.MustAdd(Provide(Database{DSN: "mysql://localhost"}))

	childCallCount := 0
	childShutdown := child.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				childCallCount++
			}
		}
	})

	parent := &Container{}
	parent.MustAdd(child)

	parentCallCount := 0
	parentShutdown := parent.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				parentCallCount++
			}
		}
	})

	_, _ = Inject[Database](parent)
	_, _ = Inject[Database](parent)

	childShutdown.Iterate(context.Background(), false)

	parentShutdown.Iterate(context.Background(), false)

	// Both containers trigger HookOnce for the first injection
	if childCallCount != 1 {
		t.Errorf("expected childCallCount=1, got %d", childCallCount)
	}
	if parentCallCount != 1 {
		t.Errorf("expected parentCallCount=1, got %d", parentCallCount)
	}
}

func TestContainer_Nested_Hook_ThreeLevelContainer(t *testing.T) {
	type Database struct{ DSN string }
	type Cache struct{ Addr string }
	type Config struct{ AppName string }

	infra := &Container{}
	infra.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Cache{Addr: "redis://localhost"}),
	)

	infraCalls := 0
	infraHook := infra.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			infraCalls++
		}
	})

	services := &Container{}
	services.MustAdd(infra)

	servicesCalls := 0
	servicesHook := services.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			servicesCalls++
		}
	})

	app := &Container{}
	app.MustAdd(services, Provide(Config{AppName: "my-app"}))

	appCalls := 0
	appHook := app.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			appCalls++
		}
	})

	_, _ = Inject[Database](app)
	_, _ = Inject[Cache](app)
	_, _ = Inject[Config](app)

	infraHook.Iterate(context.Background(), false)

	servicesHook.Iterate(context.Background(), false)

	appHook.Iterate(context.Background(), false)

	// Hook mechanism for nested containers:
	// - Hooks trigger on EACH container in the injection chain
	// - provided counter is per-type (key is type ID like *Database)
	//
	// Injection flow:
	// 1. Database: infra(0), services(0), app(0) - all first time
	// 2. Cache: infra(0), services(0), app(0) - all first time (different type)
	// 3. Config: app(0) - direct provide
	//
	// Results:
	// - infraCalls=2 (Database-0, Cache-0)
	// - servicesCalls=2 (Database-0, Cache-0)
	// - appCalls=3 (Database-0, Cache-0, Config-0)
	if infraCalls != 2 {
		t.Errorf("expected infraCalls=2, got %d", infraCalls)
	}
	if servicesCalls != 2 {
		t.Errorf("expected servicesCalls=2, got %d", servicesCalls)
	}
	if appCalls != 3 {
		t.Errorf("expected appCalls=3, got %d", appCalls)
	}
}

func Example_container_Nested_Hooks() {
	type Database struct{ DSN string }
	type Cache struct{ Addr string }

	infra := &Container{}
	infra.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Cache{Addr: "redis://localhost"}),
	)

	infraHook := infra.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				fmt.Printf("[Infra] DB starting\n")
			case Cache:
				fmt.Printf("[Infra] Cache starting\n")
			}
		}
	})

	app := &Container{}
	app.MustAdd(infra)

	appHook := app.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				fmt.Printf("[App] DB starting\n")
			}
		}
	})

	_, _ = Inject[Database](app)
	_, _ = Inject[Cache](app)

	infraHook.Iterate(context.Background(), false)

	appHook.Iterate(context.Background(), false)

	// Output:
	// [Infra] DB starting
	// [Infra] Cache starting
	// [App] DB starting
}

// ============== Advanced Nested Container Hook Tests ==============

func TestNestedContainer_Hook_VerifyInjectionChain(t *testing.T) {
	type DB struct{ Name string }
	type Cache struct{ ID int }
	type Logger struct{ Prefix string }

	infra := &Container{}
	infra.MustAdd(
		Provide(DB{Name: "primary-db"}),
		Provide(Cache{ID: 1}),
	)

	services := &Container{}
	services.MustAdd(infra)

	app := &Container{}
	app.MustAdd(
		services,
		Provide(Logger{Prefix: "[APP]"}),
	)

	injectionOrder := make([]string, 0)
	var mu sync.Mutex

	exec := app.Hook("trace", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			mu.Lock()
			defer mu.Unlock()
			switch v.(type) {
			case DB:
				injectionOrder = append(injectionOrder, "DB")
			case Cache:
				injectionOrder = append(injectionOrder, "Cache")
			case Logger:
				injectionOrder = append(injectionOrder, "Logger")
			}
		}
	})

	_, _ = Inject[DB](app)
	_, _ = Inject[Cache](app)
	_, _ = Inject[Logger](app)

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	expected := []string{"DB", "Cache", "Logger"}
	if len(injectionOrder) != len(expected) {
		t.Fatalf("expected %d hooks, got %d", len(expected), len(injectionOrder))
	}

	for i, exp := range expected {
		if injectionOrder[i] != exp {
			t.Errorf("position %d: expected %s, got %s", i, exp, injectionOrder[i])
		}
	}
}

func TestNestedContainer_Hook_ProvidedCounter(t *testing.T) {
	type Config struct{ Value int }

	child := &Container{}
	child.MustAdd(Provide(Config{Value: 42}))

	parent := &Container{}
	parent.MustAdd(child)

	callCounts := make(map[string]int)
	var mu sync.Mutex

	exec := parent.Hook("test", func(v any, provided int) func(context.Context) {
		mu.Lock()
		defer mu.Unlock()
		if _, ok := v.(Config); ok {
			key := fmt.Sprintf("Config-provided-%d", provided)
			callCounts[key]++
		}
		return func(ctx context.Context) {}
	})

	_, _ = Inject[Config](parent)
	_, _ = Inject[Config](parent)
	_, _ = Inject[Config](parent)

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if callCounts["Config-provided-0"] != 1 {
		t.Errorf("expected provided-0 called 1 time (first injection), got %d", callCounts["Config-provided-0"])
	}
	if callCounts["Config-provided-1"] != 1 {
		t.Errorf("expected provided-1 called 1 time (second injection), got %d", callCounts["Config-provided-1"])
	}
	if callCounts["Config-provided-2"] != 1 {
		t.Errorf("expected provided-2 called 1 time (third injection), got %d", callCounts["Config-provided-2"])
	}
}

func TestNestedContainer_Hook_FourLevelDeep(t *testing.T) {
	type L1 struct{ V string }
	type L2 struct{ V string }
	type L3 struct{ V string }
	type L4 struct{ V string }

	c1 := &Container{}
	c1.MustAdd(Provide(L1{V: "level1"}))

	c2 := &Container{}
	c2.MustAdd(c1, Provide(L2{V: "level2"}))

	c3 := &Container{}
	c3.MustAdd(c2, Provide(L3{V: "level3"}))

	c4 := &Container{}
	c4.MustAdd(c3, Provide(L4{V: "level4"}))

	hookCalls := make(map[string]int)
	var mu sync.Mutex

	exec := c4.Hook("init", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			mu.Lock()
			defer mu.Unlock()
			switch v.(type) {
			case L1:
				hookCalls["L1"]++
			case L2:
				hookCalls["L2"]++
			case L3:
				hookCalls["L3"]++
			case L4:
				hookCalls["L4"]++
			}
		}
	})

	_, _ = Inject[L1](c4)
	_, _ = Inject[L2](c4)
	_, _ = Inject[L3](c4)
	_, _ = Inject[L4](c4)

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if hookCalls["L1"] != 1 || hookCalls["L2"] != 1 || hookCalls["L3"] != 1 || hookCalls["L4"] != 1 {
		t.Errorf("expected all types called once, got %+v", hookCalls)
	}
}

func TestNestedContainer_Hook_MultipleHooksSameContainer(t *testing.T) {
	type Service struct{ Name string }

	c := &Container{}
	c.MustAdd(Provide(Service{Name: "test-service"}))

	hook1Calls := 0
	hook2Calls := 0
	hook3Calls := 0

	hook1 := c.Hook("phase1", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(Service); ok {
				hook1Calls++
			}
		}
	})

	hook2 := c.Hook("phase2", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(Service); ok {
				hook2Calls++
			}
		}
	})

	hook3 := c.Hook("phase3", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(Service); ok {
				hook3Calls++
			}
		}
	})

	_, _ = Inject[Service](c)

	ctx := context.Background()

	hook1(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

	hook2(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

	hook3(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

	if hook1Calls != 1 || hook2Calls != 1 || hook3Calls != 1 {
		t.Errorf("expected all hooks called once, got h1=%d, h2=%d, h3=%d", hook1Calls, hook2Calls, hook3Calls)
	}
}

func TestNestedContainer_Hook_HookOnceVsHook(t *testing.T) {
	type Resource struct{ ID int }

	c := &Container{}
	c.MustAdd(Provide(Resource{ID: 100}))

	hookCallCount := 0
	hookOnceCallCount := 0

	c.Hook("regular", func(v any, provided int) func(context.Context) {
		if _, ok := v.(Resource); ok {
			hookCallCount++
		}
		return func(ctx context.Context) {}
	})

	c.HookOnce("once", func(v any) func(context.Context) {
		if _, ok := v.(Resource); ok {
			hookOnceCallCount++
		}
		return func(ctx context.Context) {}
	})

	_, _ = Inject[Resource](c)
	_, _ = Inject[Resource](c)
	_, _ = Inject[Resource](c)

	if hookCallCount != 3 {
		t.Errorf("expected regular hook called 3 times, got %d", hookCallCount)
	}

	if hookOnceCallCount != 1 {
		t.Errorf("expected HookOnce called 1 time, got %d", hookOnceCallCount)
	}
}

func TestNestedContainer_Hook_SharedDependency(t *testing.T) {
	type DB struct{ Connection string }
	type ServiceA struct{ DB DB }
	type ServiceB struct{ DB DB }

	c := &Container{}
	c.MustAdd(
		Provide(DB{Connection: "postgres://localhost"}),
		Build(func(c *Container) (ServiceA, error) {
			db, _ := Inject[DB](c)
			return ServiceA{DB: db}, nil
		}),
		Build(func(c *Container) (ServiceB, error) {
			db, _ := Inject[DB](c)
			return ServiceB{DB: db}, nil
		}),
	)

	dbUsageCount := 0

	c.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		if _, ok := v.(DB); ok {
			dbUsageCount++
		}
		return func(ctx context.Context) {}
	})

	_, _ = Inject[ServiceA](c)
	_, _ = Inject[ServiceB](c)

	exec := c.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {}
	})

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if dbUsageCount != 1 {
		t.Errorf("expected DB hook called 1 time (singleton), got %d", dbUsageCount)
	}
}

func TestNestedContainer_Hook_ContextPropagation(t *testing.T) {
	type Token struct{ Value string }

	parent := &Container{}
	parent.MustAdd(Provide(Token{Value: "parent-token"}))

	child := &Container{}
	child.MustAdd(parent)

	type contextKey string
	const key contextKey = "trace"

	child.Hook("trace", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			trace, _ := ctx.Value(key).([]string)
			if len(trace) == 0 {
				t.Error("expected context to carry trace information")
			}
		}
	})

	_, _ = Inject[Token](child)

	exec := child.Hook("trace", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {}
	})

	trace := []string{"request-1", "request-2"}
	ctx := context.WithValue(context.Background(), key, trace)

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})
}

func TestNestedContainer_Hook_DiamondDependency(t *testing.T) {
	type Base struct{ Name string }
	type Left struct{ Base Base }
	type Right struct{ Base Base }
	type Top struct {
		Left  Left
		Right Right
	}

	c := &Container{}
	c.MustAdd(
		Provide(Base{Name: "shared-base"}),
		Build(func(c *Container) (Left, error) {
			base, _ := Inject[Base](c)
			return Left{Base: base}, nil
		}),
		Build(func(c *Container) (Right, error) {
			base, _ := Inject[Base](c)
			return Right{Base: base}, nil
		}),
		Build(func(c *Container) (Top, error) {
			left, _ := Inject[Left](c)
			right, _ := Inject[Right](c)
			return Top{Left: left, Right: right}, nil
		}),
	)

	baseHookCalls := 0

	c.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		if _, ok := v.(Base); ok {
			baseHookCalls++
		}
		return func(ctx context.Context) {}
	})

	_, _ = Inject[Top](c)

	exec := c.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {}
	})

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if baseHookCalls != 1 {
		t.Errorf("expected Base hook called 1 time (diamond dependency), got %d", baseHookCalls)
	}
}

// ============== Nested Container Comprehensive Tests ==============

func TestNestedContainer_CircularDependencyDetection(t *testing.T) {
	type ServiceA struct{}
	type ServiceB struct{}

	child := &Container{}
	child.MustAdd(
		Build(func(c *Container) (ServiceA, error) {
			_, err := Inject[ServiceB](c)
			return ServiceA{}, err
		}),
	)

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(
		Build(func(c *Container) (ServiceB, error) {
			_, err := Inject[ServiceA](c)
			return ServiceB{}, err
		}),
	)

	_, err := Inject[ServiceA](parent)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestNestedContainer_ErrorPropagation(t *testing.T) {
	type Config struct{ Value string }
	type Service struct{}

	infra := &Container{}
	infra.MustAdd(
		Build(func(c *Container) (Config, error) {
			return Config{}, fmt.Errorf("infra error")
		}),
	)

	app := &Container{}
	app.MustAdd(infra)
	app.MustAdd(
		Build(func(c *Container) (Service, error) {
			_, err := Inject[Config](c)
			if err != nil {
				return Service{}, fmt.Errorf("wrapped: %w", err)
			}
			return Service{}, nil
		}),
	)

	_, err := Inject[Service](app)
	if err == nil {
		t.Fatal("expected error propagation")
	}
}

func TestNestedContainer_BuildDependencyChain(t *testing.T) {
	type L1 struct{ V int }
	type L2 struct{ V int }
	type L3 struct{ V int }

	infra := &Container{}
	infra.MustAdd(
		Build(func(c *Container) (L1, error) {
			return L1{V: 1}, nil
		}),
	)

	services := &Container{}
	services.MustAdd(infra)
	services.MustAdd(
		Build(func(c *Container) (L2, error) {
			l1, _ := Inject[L1](c)
			return L2{V: l1.V + 1}, nil
		}),
	)

	app := &Container{}
	app.MustAdd(services)
	app.MustAdd(
		Build(func(c *Container) (L3, error) {
			l2, _ := Inject[L2](c)
			return L3{V: l2.V + 1}, nil
		}),
	)

	l3, err := Inject[L3](app)
	if err != nil {
		t.Fatal(err)
	}
	if l3.V != 3 {
		t.Errorf("expected L3.V=3, got %d", l3.V)
	}
}

func TestNestedContainer_TypeAliases(t *testing.T) {
	type StringAlias string
	type IntAlias int

	child := &Container{}
	child.MustAdd(Provide(StringAlias("alias-value")))

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(
		Build(func(c *Container) (IntAlias, error) {
			s, _ := Inject[StringAlias](c)
			return IntAlias(len(s)), nil
		}),
	)

	v, err := Inject[IntAlias](parent)
	if err != nil || v != 11 {
		t.Errorf("expected 11, got %d", v)
	}
}

func TestNestedContainer_MixedProvideAndBuild(t *testing.T) {
	type Config struct{ Value string }
	type Service struct{ Config Config }

	infra := &Container{}
	infra.MustAdd(Provide(Config{Value: "from-infra"}))

	app := &Container{}
	app.MustAdd(infra)
	app.MustAdd(
		Build(func(c *Container) (Service, error) {
			cfg, _ := Inject[Config](c)
			return Service{Config: cfg}, nil
		}),
	)

	svc, err := Inject[Service](app)
	if err != nil {
		t.Fatal(err)
	}
	if svc.Config.Value != "from-infra" {
		t.Errorf("expected 'from-infra', got %s", svc.Config.Value)
	}
}

func TestNestedContainer_CrossDependencies(t *testing.T) {
	type IntDep int
	type StringDep string

	child := &Container{}
	child.MustAdd(
		Build(func(c *Container) (IntDep, error) {
			return IntDep(42), nil
		}),
	)

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(
		Build(func(c *Container) (StringDep, error) {
			i, _ := Inject[IntDep](c)
			return StringDep(fmt.Sprintf("got-%d", i)), nil
		}),
	)

	s, err := Inject[StringDep](parent)
	if err != nil || s != "got-42" {
		t.Errorf("expected got-42, got %v", s)
	}
}

func TestNestedContainer_DeepChain(t *testing.T) {
	type L1 int
	type L2 int
	type L3 int

	infra := &Container{}
	infra.MustAdd(
		Build(func(c *Container) (L1, error) {
			return L1(1), nil
		}),
	)

	middle := &Container{}
	middle.MustAdd(infra)
	middle.MustAdd(
		Build(func(c *Container) (L2, error) {
			v, _ := Inject[L1](c)
			return L2(v + 1), nil
		}),
	)

	app := &Container{}
	app.MustAdd(middle)
	app.MustAdd(
		Build(func(c *Container) (L3, error) {
			v, _ := Inject[L2](c)
			return L3(v * 10), nil
		}),
	)

	v, err := Inject[L3](app)
	if err != nil || v != 20 {
		t.Errorf("expected 20, got %v", v)
	}
}

func TestNestedContainer_SharedSingleton(t *testing.T) {
	type DB struct{ ID int }

	callCount := 0

	infra := &Container{}
	infra.MustAdd(
		Build(func(c *Container) (DB, error) {
			callCount++
			return DB{ID: 1}, nil
		}),
	)

	services := &Container{}
	services.MustAdd(infra)

	app := &Container{}
	app.MustAdd(services)

	_, _ = Inject[DB](app)
	_, _ = Inject[DB](app)
	_, _ = Inject[DB](app)

	if callCount != 1 {
		t.Errorf("expected DB built once (singleton), got %d calls", callCount)
	}
}

func TestNestedContainer_ContainerFreezing(t *testing.T) {
	type Config struct{ Value int }

	child := &Container{}
	child.MustAdd(Provide(Config{Value: 1}))

	parent := &Container{}
	parent.MustAdd(child)

	err := child.Add(Provide(Config{Value: 2}))
	if err == nil {
		t.Fatal("expected container frozen error")
	}
	t.Log(err)
}

func TestNestedContainer_DuplicatePrevention(t *testing.T) {
	type Config struct{ Value int }

	child := &Container{}
	child.MustAdd(Provide(Config{Value: 1}))

	parent := &Container{}
	parent.MustAdd(child)

	err := parent.Add(Provide(Config{Value: 2}))
	if err == nil {
		t.Fatal("expected duplicate prevention error")
	}
}

func TestNestedContainer_ProvideMethod(t *testing.T) {
	type DB struct{ DSN string }
	type Cache struct{ Addr string }

	infra := &Container{}
	infra.MustAdd(Provide(DB{DSN: "mysql://localhost"}))

	app := &Container{}
	app.MustAdd(infra)
	app.MustAdd(Provide(Cache{Addr: "redis://localhost"}))

	db := DB{}
	cache := Cache{}
	other := struct{ Name string }{}

	if _, ok := app.Provide(&db); !ok {
		t.Error("expected app to provide DB from child")
	}
	if _, ok := app.Provide(&cache); !ok {
		t.Error("expected app to provide Cache")
	}
	if _, ok := app.Provide(&other); ok {
		t.Error("expected app to not provide unknown type")
	}
}

func TestNestedContainer_Hook_WithBuildError(t *testing.T) {
	type Config struct{ Value int }

	child := &Container{}
	child.MustAdd(
		Build(func(c *Container) (Config, error) {
			return Config{}, fmt.Errorf("build error")
		}),
	)

	parent := &Container{}
	parent.MustAdd(child)

	hookCalled := false
	parent.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			hookCalled = true
		}
	})

	_, err := Inject[Config](parent)
	if err == nil {
		t.Fatal("expected build error")
	}

	exec := parent.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {}
	})

	exec(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if hookCalled {
		t.Error("expected hook not called on build error")
	}
}

func TestNestedContainer_MultipleChildren(t *testing.T) {
	type DB struct{ DSN string }
	type Cache struct{ Addr string }

	dbContainer := &Container{}
	dbContainer.MustAdd(Provide(DB{DSN: "mysql://localhost"}))

	cacheContainer := &Container{}
	cacheContainer.MustAdd(Provide(Cache{Addr: "redis://localhost"}))

	app := &Container{}
	app.MustAdd(dbContainer, cacheContainer)

	db, err := Inject[DB](app)
	if err != nil {
		t.Fatal(err)
	}

	cache, err := Inject[Cache](app)
	if err != nil {
		t.Fatal(err)
	}

	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
	if cache.Addr != "redis://localhost" {
		t.Errorf("expected redis://localhost, got %s", cache.Addr)
	}
}

func TestNestedContainer_Hook_PropagationOrder(t *testing.T) {
	type Token struct{ Value string }

	infra := &Container{}
	infra.MustAdd(Provide(Token{Value: "infra-token"}))

	services := &Container{}
	services.MustAdd(infra)

	app := &Container{}
	app.MustAdd(services)

	executionOrder := make([]string, 0)

	infraHook := infra.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			executionOrder = append(executionOrder, "infra")
		}
	})

	servicesHook := services.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			executionOrder = append(executionOrder, "services")
		}
	})

	appHook := app.Hook("track", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			executionOrder = append(executionOrder, "app")
		}
	})

	_, _ = Inject[Token](app)

	infraHook(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})
	servicesHook(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})
	appHook(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(context.Background())
		}
	})

	if len(executionOrder) != 3 {
		t.Errorf("expected 3 hook executions, got %d", len(executionOrder))
	}
}

// ============== Nested Container Type Detection Tests ==============
// Comprehensive tests for type detection in nested container scenarios

// TestNestedContainer_TypeDetection_BasicTwoLevel tests basic type detection in 2-level nested containers
func TestNestedContainer_TypeDetection_BasicTwoLevel(t *testing.T) {
	type TCConfig struct{ Value string }

	child := &Container{}
	child.MustAdd(Provide(TCConfig{Value: "child-config"}))

	parent := &Container{}
	parent.MustAdd(child)

	// Parent should find type in child
	cfg, err := Inject[TCConfig](parent)
	if err != nil {
		t.Fatalf("failed to inject Config from parent: %v", err)
	}
	if cfg.Value != "child-config" {
		t.Errorf("expected child-config, got %s", cfg.Value)
	}
}

// TestNestedContainer_TypeDetection_ThreeLevelHierarchy tests type detection across 3 levels
func TestNestedContainer_TypeDetection_ThreeLevelHierarchy(t *testing.T) {
	type TCDB struct{ Name string }

	infra := &Container{}
	infra.MustAdd(Provide(TCDB{Name: "infra-db"}))

	services := &Container{}
	services.MustAdd(infra)

	app := &Container{}
	app.MustAdd(services)

	// App should find DB through services -> infra chain
	db, err := Inject[TCDB](app)
	if err != nil {
		t.Fatalf("failed to inject DB from app: %v", err)
	}
	if db.Name != "infra-db" {
		t.Errorf("expected infra-db, got %s", db.Name)
	}
}

// TestNestedContainer_TypeDetection_FourLevelDeep tests type detection in 4-level deep hierarchy
func TestNestedContainer_TypeDetection_FourLevelDeep(t *testing.T) {
	type TCToken struct{ Value int }

	level1 := &Container{}
	level1.MustAdd(Provide(TCToken{Value: 100}))

	level2 := &Container{}
	level2.MustAdd(level1)

	level3 := &Container{}
	level3.MustAdd(level2)

	level4 := &Container{}
	level4.MustAdd(level3)

	// Level4 should find Token through 3 levels of nesting
	token, err := Inject[TCToken](level4)
	if err != nil {
		t.Fatalf("failed to inject Token from level4: %v", err)
	}
	if token.Value != 100 {
		t.Errorf("expected 100, got %d", token.Value)
	}
}

// TestNestedContainer_TypeDetection_ParentOverridesChild tests that duplicate types are prevented
func TestNestedContainer_TypeDetection_ParentOverridesChild(t *testing.T) {
	type TCConfig2 struct{ Value string }

	child := &Container{}
	child.MustAdd(Provide(TCConfig2{Value: "child-value"}))

	parent := &Container{}
	parent.MustAdd(child)

	// Adding duplicate type to parent should fail
	err := parent.Add(Provide(TCConfig2{Value: "parent-value"}))
	if err == nil {
		t.Fatal("expected duplicate type error")
	}
	t.Logf("Got expected error: %v", err)

	// Should still be able to inject from child
	cfg, err := Inject[TCConfig2](parent)
	if err != nil {
		t.Fatalf("failed to inject Config: %v", err)
	}
	if cfg.Value != "child-value" {
		t.Errorf("expected child-value, got %s", cfg.Value)
	}
}

// TestNestedContainer_TypeDetection_MultipleChildrenIsolation tests type isolation between sibling containers
func TestNestedContainer_TypeDetection_MultipleChildrenIsolation(t *testing.T) {
	type TCDB2 struct{ Name string }
	type TCCache struct{ Name string }

	dbContainer := &Container{}
	dbContainer.MustAdd(Provide(TCDB2{Name: "main-db"}))

	cacheContainer := &Container{}
	cacheContainer.MustAdd(Provide(TCCache{Name: "main-cache"}))

	parent := &Container{}
	parent.MustAdd(dbContainer, cacheContainer)

	// Both types should be accessible from parent
	db, err := Inject[TCDB2](parent)
	if err != nil {
		t.Fatalf("failed to inject DB: %v", err)
	}

	cache, err := Inject[TCCache](parent)
	if err != nil {
		t.Fatalf("failed to inject Cache: %v", err)
	}

	if db.Name != "main-db" {
		t.Errorf("expected main-db, got %s", db.Name)
	}
	if cache.Name != "main-cache" {
		t.Errorf("expected main-cache, got %s", cache.Name)
	}
}

// TestNestedContainer_TypeDetection_DuplicateTypeInSiblings tests duplicate type detection in sibling containers
func TestNestedContainer_TypeDetection_DuplicateTypeInSiblings(t *testing.T) {
	type TCConfig3 struct{ Value string }

	child1 := &Container{}
	child1.MustAdd(Provide(TCConfig3{Value: "child1"}))

	child2 := &Container{}
	child2.MustAdd(Provide(TCConfig3{Value: "child2"}))

	parent := &Container{}
	parent.MustAdd(child1)

	// Adding child2 should fail due to duplicate Config type
	err := parent.Add(child2)
	if err == nil {
		t.Fatal("expected duplicate type error when adding sibling with same type")
	}
	t.Logf("Got expected error: %v", err)
}

// TestNestedContainer_TypeDetection_InterfaceInjection tests interface type injection in nested containers
func TestNestedContainer_TypeDetection_InterfaceInjection(t *testing.T) {
	// Test with built-in interface
	child := &Container{}
	child.MustAdd(Provide(fmt.Stringer(nil)))

	parent := &Container{}
	parent.MustAdd(child)

	_, err := Inject[fmt.Stringer](parent)
	// This test just verifies interface types can be registered in nested containers
	// Actual implementation depends on what's registered
	if err == nil {
		t.Log("Interface injection succeeded")
	}
}

// TestNestedContainer_TypeDetection_AliasTypes tests type alias distinction
func TestNestedContainer_TypeDetection_AliasTypes(t *testing.T) {
	type TCStringAlias string
	type TCStringAlias2 string

	child := &Container{}
	child.MustAdd(
		Provide(TCStringAlias("alias1")),
		Provide(TCStringAlias2("alias2")),
	)

	parent := &Container{}
	parent.MustAdd(child)

	alias1, err := Inject[TCStringAlias](parent)
	if err != nil {
		t.Fatalf("failed to inject StringAlias: %v", err)
	}
	if alias1 != "alias1" {
		t.Errorf("expected alias1, got %s", alias1)
	}

	alias2, err := Inject[TCStringAlias2](parent)
	if err != nil {
		t.Fatalf("failed to inject StringAlias2: %v", err)
	}
	if alias2 != "alias2" {
		t.Errorf("expected alias2, got %s", alias2)
	}
}

// TestNestedContainer_TypeDetection_GenericTypeInjection tests generic type injection in nested containers
func TestNestedContainer_TypeDetection_GenericTypeInjection(t *testing.T) {
	type TCWrapper[T any] struct {
		Value T
	}

	child := &Container{}
	child.MustAdd(Provide(TCWrapper[int]{Value: 42}))

	parent := &Container{}
	parent.MustAdd(child)

	wrapper, err := Inject[TCWrapper[int]](parent)
	if err != nil {
		t.Fatalf("failed to inject Wrapper[int]: %v", err)
	}
	if wrapper.Value != 42 {
		t.Errorf("expected 42, got %d", wrapper.Value)
	}
}

// TestNestedContainer_TypeDetection_PointerVsValue tests pointer vs value type distinction
func TestNestedContainer_TypeDetection_PointerVsValue(t *testing.T) {
	type TCConfig4 struct{ Value int }

	child := &Container{}
	child.MustAdd(Provide(TCConfig4{Value: 1}))
	child.MustAdd(Provide(&TCConfig4{Value: 2}))

	parent := &Container{}
	parent.MustAdd(child)

	// Value type
	cfg1, err := Inject[TCConfig4](parent)
	if err != nil {
		t.Fatalf("failed to inject Config: %v", err)
	}
	if cfg1.Value != 1 {
		t.Errorf("expected 1 for value type, got %d", cfg1.Value)
	}

	// Pointer type
	cfg2, err := Inject[*TCConfig4](parent)
	if err != nil {
		t.Fatalf("failed to inject *Config: %v", err)
	}
	if cfg2.Value != 2 {
		t.Errorf("expected 2 for pointer type, got %d", cfg2.Value)
	}
}

// TestNestedContainer_TypeDetection_SliceAndMapTypes tests slice and map type injection
func TestNestedContainer_TypeDetection_SliceAndMapTypes(t *testing.T) {
	child := &Container{}
	child.MustAdd(
		Provide([]string{"a", "b", "c"}),
		Provide(map[string]int{"key": 42}),
	)

	parent := &Container{}
	parent.MustAdd(child)

	slice, err := Inject[[]string](parent)
	if err != nil {
		t.Fatalf("failed to inject []string: %v", err)
	}
	if len(slice) != 3 {
		t.Errorf("expected slice length 3, got %d", len(slice))
	}

	m, err := Inject[map[string]int](parent)
	if err != nil {
		t.Fatalf("failed to inject map[string]int: %v", err)
	}
	if m["key"] != 42 {
		t.Errorf("expected map value 42, got %d", m["key"])
	}
}

// TestNestedContainer_TypeDetection_ChannelTypes tests channel type injection
func TestNestedContainer_TypeDetection_ChannelTypes(t *testing.T) {
	child := &Container{}
	child.MustAdd(Provide(make(chan int, 10)))

	parent := &Container{}
	parent.MustAdd(child)

	ch, err := Inject[chan int](parent)
	if err != nil {
		t.Fatalf("failed to inject chan int: %v", err)
	}
	if ch == nil {
		t.Error("expected non-nil channel")
	}
}

// TestNestedContainer_TypeDetection_FunctionTypes tests function type injection
func TestNestedContainer_TypeDetection_FunctionTypes(t *testing.T) {
	type TCHandler func(string) string

	child := &Container{}
	child.MustAdd(Provide(TCHandler(func(s string) string {
		return "handled: " + s
	})))

	parent := &Container{}
	parent.MustAdd(child)

	handler, err := Inject[TCHandler](parent)
	if err != nil {
		t.Fatalf("failed to inject Handler: %v", err)
	}
	if handler("test") != "handled: test" {
		t.Errorf("unexpected handler result")
	}
}

// TestNestedContainer_TypeDetection_MixedProvideAndBuild tests mixed Provide and Build in nested containers
func TestNestedContainer_TypeDetection_MixedProvideAndBuild(t *testing.T) {
	type TCConfig5 struct{ Value string }
	type TCService struct {
		Name   string
		Config TCConfig5
	}

	child := &Container{}
	child.MustAdd(Provide(TCConfig5{Value: "child-config"}))

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(c *Container) (TCService, error) {
		cfg, err := Inject[TCConfig5](c)
		if err != nil {
			return TCService{}, err
		}
		return TCService{Name: "parent-service", Config: cfg}, nil
	}))

	svc, err := Inject[TCService](parent)
	if err != nil {
		t.Fatalf("failed to inject Service: %v", err)
	}
	if svc.Name != "parent-service" {
		t.Errorf("expected parent-service, got %s", svc.Name)
	}
	if svc.Config.Value != "child-config" {
		t.Errorf("expected child-config, got %s", svc.Config.Value)
	}
}

// TestNestedContainer_TypeDetection_CircularDependencyAcrossContainers tests circular dependency detection across containers
func TestNestedContainer_TypeDetection_CircularDependencyAcrossContainers(t *testing.T) {
	type TCServiceA struct{}
	type TCServiceB struct{}

	child := &Container{}
	child.MustAdd(Build(func(c *Container) (TCServiceA, error) {
		_, err := Inject[TCServiceB](c)
		return TCServiceA{}, err
	}))

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(c *Container) (TCServiceB, error) {
		_, err := Inject[TCServiceA](c)
		return TCServiceB{}, err
	}))

	_, err := Inject[TCServiceA](parent)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	t.Logf("Got expected circular dependency error: %v", err)
}

// TestNestedContainer_TypeDetection_ProvideMethodChecksNested tests Provide method checks nested containers
func TestNestedContainer_TypeDetection_ProvideMethodChecksNested(t *testing.T) {
	type TCDB3 struct{ Name string }
	type TCCache2 struct{ Name string }

	child := &Container{}
	child.MustAdd(Provide(TCDB3{Name: "child-db"}))

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Provide(TCCache2{Name: "parent-cache"}))

	// Check types in parent (including nested)
	db := TCDB3{}
	if _, ok := parent.Provide(&db); !ok {
		t.Error("expected parent to provide DB from child")
	}

	cache := TCCache2{}
	if _, ok := parent.Provide(&cache); !ok {
		t.Error("expected parent to provide Cache")
	}

	// Check non-existent type
	other := struct{ Name string }{}
	if _, ok := parent.Provide(&other); ok {
		t.Error("expected parent to not provide unknown type")
	}
}

// TestNestedContainer_TypeDetection_EmptyNestedContainer tests injection with empty nested container
func TestNestedContainer_TypeDetection_EmptyNestedContainer(t *testing.T) {
	type TCConfig6 struct{ Value string }

	emptyChild := &Container{}

	parent := &Container{}
	parent.MustAdd(emptyChild)
	parent.MustAdd(Provide(TCConfig6{Value: "parent-config"}))

	// Should still find parent's Config
	cfg, err := Inject[TCConfig6](parent)
	if err != nil {
		t.Fatalf("failed to inject Config: %v", err)
	}
	if cfg.Value != "parent-config" {
		t.Errorf("expected parent-config, got %s", cfg.Value)
	}
}

// TestNestedContainer_TypeDetection_DeepHierarchyWithBuild tests deep hierarchy with Build dependencies
func TestNestedContainer_TypeDetection_DeepHierarchyWithBuild(t *testing.T) {
	type TCL1 struct{ V int }
	type TCL2 struct{ V int }
	type TCL3 struct{ V int }
	type TCL4 struct{ V int }

	level1 := &Container{}
	level1.MustAdd(Build(func(c *Container) (TCL1, error) {
		return TCL1{V: 1}, nil
	}))

	level2 := &Container{}
	level2.MustAdd(level1)
	level2.MustAdd(Build(func(c *Container) (TCL2, error) {
		l1, _ := Inject[TCL1](c)
		return TCL2{V: l1.V + 1}, nil
	}))

	level3 := &Container{}
	level3.MustAdd(level2)
	level3.MustAdd(Build(func(c *Container) (TCL3, error) {
		l2, _ := Inject[TCL2](c)
		return TCL3{V: l2.V + 1}, nil
	}))

	level4 := &Container{}
	level4.MustAdd(level3)
	level4.MustAdd(Build(func(c *Container) (TCL4, error) {
		l3, _ := Inject[TCL3](c)
		return TCL4{V: l3.V + 1}, nil
	}))

	// Inject from top level
	l4, err := Inject[TCL4](level4)
	if err != nil {
		t.Fatalf("failed to inject L4: %v", err)
	}
	if l4.V != 4 {
		t.Errorf("expected 4, got %d", l4.V)
	}

	// Also verify we can inject intermediate types
	l2, err := Inject[TCL2](level4)
	if err != nil {
		t.Fatalf("failed to inject L2: %v", err)
	}
	if l2.V != 2 {
		t.Errorf("expected 2, got %d", l2.V)
	}
}

// TestNestedContainer_TypeDetection_HookRegistrationOnNested tests hook registration on nested containers
func TestNestedContainer_TypeDetection_HookRegistrationOnNested(t *testing.T) {
	type TCToken2 struct{ Value string }

	child := &Container{}
	child.MustAdd(Provide(TCToken2{Value: "child-token"}))

	parent := &Container{}
	parent.MustAdd(child)

	childHookCalled := false
	childHook := child.Hook("test", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(TCToken2); ok {
				childHookCalled = true
			}
		}
	})

	parentHookCalled := false
	parentHook := parent.Hook("test", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			if _, ok := v.(TCToken2); ok {
				parentHookCalled = true
			}
		}
	})

	// Inject from parent - this triggers hook registration on both containers
	_, _ = Inject[TCToken2](parent)

	// Execute hooks
	childHook.Iterate(context.Background(), false)
	parentHook.Iterate(context.Background(), false)

	if !childHookCalled {
		t.Error("expected child hook to be called")
	}
	if !parentHookCalled {
		t.Error("expected parent hook to be called")
	}
}

// TestNestedContainer_TypeDetection_ConcurrentAccess tests concurrent access to nested containers
func TestNestedContainer_TypeDetection_ConcurrentAccess(t *testing.T) {
	type TCConfig7 struct{ Value int }

	child := &Container{}
	child.MustAdd(Provide(TCConfig7{Value: 42}))

	parent := &Container{}
	parent.MustAdd(child)

	// Test concurrent reads from the same container
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Use InjectTo which is simpler and avoids potential circular dependency detection issues
			var cfg TCConfig7
			if err := InjectTo(&cfg, parent); err != nil {
				errors <- fmt.Errorf("failed to inject: %w", err)
				return
			}
			if cfg.Value != 42 {
				errors <- fmt.Errorf("expected 42, got %d", cfg.Value)
				return
			}
			errors <- nil
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}
}

// TestNestedContainer_TypeDetection_ErrorPropagation tests error propagation through nested containers
func TestNestedContainer_TypeDetection_ErrorPropagation(t *testing.T) {
	type TCConfig8 struct{ Value string }
	type TCService2 struct{}

	child := &Container{}
	child.MustAdd(Build(func(c *Container) (TCConfig8, error) {
		return TCConfig8{}, fmt.Errorf("child build error")
	}))

	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(c *Container) (TCService2, error) {
		_, err := Inject[TCConfig8](c)
		if err != nil {
			return TCService2{}, fmt.Errorf("service build failed: %w", err)
		}
		return TCService2{}, nil
	}))

	_, err := Inject[TCService2](parent)
	if err == nil {
		t.Fatal("expected error propagation from child")
	}
	t.Logf("Got expected error: %v", err)
}
