package godi

import (
	"context"
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
	if err := p.inject(nil, &got); err != nil {
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
	id := p.ID()
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

func TestMultiContainer(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() ([]*Container, *Container)
		injectFn func([]*Container) (any, error)
		wantVal  any
	}{
		{
			name: "Database from c1",
			setup: func() ([]*Container, *Container) {
				c1, c2, c3 := &Container{}, &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				c3.MustAdd(Provide(Service{Name: "svc3"}))
				return []*Container{c1, c2, c3}, c1
			},
			injectFn: func(cs []*Container) (any, error) { return Inject[Database](cs...) },
			wantVal:  Database{DSN: "db1"},
		},
		{
			name: "Config from c2",
			setup: func() ([]*Container, *Container) {
				c1, c2, c3 := &Container{}, &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				c3.MustAdd(Provide(Service{Name: "svc3"}))
				return []*Container{c1, c2, c3}, c2
			},
			injectFn: func(cs []*Container) (any, error) { return Inject[Config](cs...) },
			wantVal:  Config{AppName: "app2"},
		},
		{
			name: "Service from c3",
			setup: func() ([]*Container, *Container) {
				c1, c2, c3 := &Container{}, &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				c3.MustAdd(Provide(Service{Name: "svc3"}))
				return []*Container{c1, c2, c3}, c3
			},
			injectFn: func(cs []*Container) (any, error) { return Inject[Service](cs...) },
			wantVal:  Service{Name: "svc3"},
		},
		{
			name: "InjectTo Database",
			setup: func() ([]*Container, *Container) {
				c1, c2 := &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				return []*Container{c1, c2}, c1
			},
			injectFn: func(cs []*Container) (any, error) {
				var db Database
				err := InjectTo(&db, cs...)
				return db, err
			},
			wantVal: Database{DSN: "db1"},
		},
		{
			name: "InjectTo Config",
			setup: func() ([]*Container, *Container) {
				c1, c2 := &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				return []*Container{c1, c2}, c2
			},
			injectFn: func(cs []*Container) (any, error) {
				var cfg Config
				err := InjectTo(&cfg, cs...)
				return cfg, err
			},
			wantVal: Config{AppName: "app2"},
		},
		{
			name: "InjectAs Database",
			setup: func() ([]*Container, *Container) {
				c1, c2 := &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				return []*Container{c1, c2}, c1
			},
			injectFn: func(cs []*Container) (any, error) {
				db := Database{}
				err := InjectAs(&db, cs...)
				return db, err
			},
			wantVal: Database{DSN: "db1"},
		},
		{
			name: "InjectAs Config",
			setup: func() ([]*Container, *Container) {
				c1, c2 := &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				c2.MustAdd(Provide(Config{AppName: "app2"}))
				return []*Container{c1, c2}, c2
			},
			injectFn: func(cs []*Container) (any, error) {
				cfg := Config{}
				err := InjectAs(&cfg, cs...)
				return cfg, err
			},
			wantVal: Config{AppName: "app2"},
		},
		{
			name: "not found",
			setup: func() ([]*Container, *Container) {
				c1, c2 := &Container{}, &Container{}
				c1.MustAdd(Provide(Database{DSN: "db1"}))
				return []*Container{c1, c2}, nil
			},
			injectFn: func(cs []*Container) (any, error) {
				_, err := Inject[Config](cs...)
				return nil, err
			},
			wantVal: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs, _ := tt.setup()
			got, err := tt.injectFn(cs)
			if tt.wantVal == nil {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantVal {
				t.Errorf("got %v, want %v", got, tt.wantVal)
			}
		})
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

func Example_multiContainer() {
	c1, c2 := &Container{}, &Container{}
	c1.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	c2.MustAdd(Provide(Config{AppName: "multi-app"}))

	db, _ := Inject[Database](c1, c2)
	cfg, _ := Inject[Config](c1, c2)

	fmt.Printf("DB: %s, App: %s\n", db.DSN, cfg.AppName)
	// Output: DB: mysql://localhost, App: multi-app
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

	startup(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

	if startupCalled != 2 {
		t.Errorf("expected startupCalled=2, got %d", startupCalled)
	}

	shutdown(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

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

	h(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

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

	startup(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})

	shutdown(func(hooks []func(context.Context)) {
		for _, fn := range hooks {
			fn(ctx)
		}
	})
	// Output:
	// Starting: godi.Database
	// Starting: godi.Config
	// Stopping: godi.Database
	// Stopping: godi.Config
}
