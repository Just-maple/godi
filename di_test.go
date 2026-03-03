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

type CntTypeA struct{ ID int }
type CntTypeB struct{ ID int }
type CntTypeC struct{ ID int }
type CntTypeD struct{ ID int }

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
			addFn: func(c *Container) error { return c.Add(Provide(Database{DSN: "mysql://localhost"})) },
		},
		{
			name: "duplicate error",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				return c
			},
			addFn:   func(c *Container) error { return c.Add(Provide(Database{DSN: "mysql://remote"})) },
			wantErr: true,
		},
		{
			name: "container frozen",
			setup: func() *Container {
				child := &Container{}
				child.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
				parent := &Container{}
				parent.MustAdd(child)
				return child
			},
			addFn:   func(c *Container) error { return c.Add(Provide(Config{AppName: "test"})) },
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

func TestInject_ErrorCases(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) error
		wantErr  bool
	}{
		{name: "not found", setup: func() *Container { return &Container{} }, injectFn: func(c *Container) error {
			var db Database
			return InjectTo[Database](c, &db)
		}, wantErr: true},
		{name: "wrong type", setup: func() *Container {
			c := &Container{}
			c.MustAdd(Provide(Database{DSN: "test"}))
			return c
		}, injectFn: func(c *Container) error {
			var cfg Config
			return InjectTo[Config](c, &cfg)
		}, wantErr: true},
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

func TestContainer_MustAdd(t *testing.T) {
	tests := []struct {
		name      string
		setup     func() *Container
		addFn     func(*Container)
		wantPanic bool
	}{
		{name: "success", setup: func() *Container { return &Container{} }, addFn: func(c *Container) {
			c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		}},
		{name: "panic on duplicate", setup: func() *Container {
			c := &Container{}
			c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
			return c
		}, addFn: func(c *Container) {
			c.MustAdd(Provide(Database{DSN: "mysql://remote"}))
		}, wantPanic: true},
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

func TestProvide_AllTypes(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Provide("test"), Provide(42), Provide(int8(8)), Provide(int64(64)),
		Provide(uint(100)), Provide(float32(3.14)), Provide(3.14159), Provide(true),
		Provide([]string{"a", "b", "c"}), Provide([]int{1, 2, 3}),
		Provide(map[string]int{"a": 1}), Provide(&struct{ Name string }{Name: "Alice"}),
		Provide(make(chan int)), Provide(func() string { return "hello" }),
	)
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
		{"float64", func() error {
			v, err := Inject[float64](c)
			if err != nil || v != 3.14159 {
				return fmt.Errorf("got %v", v)
			}
			return nil
		}},
		{"[]string", func() error {
			v, err := Inject[[]string](c)
			if err != nil || len(v) != 3 {
				return fmt.Errorf("got %v", v)
			}
			return nil
		}},
		{"map", func() error {
			v, err := Inject[map[string]int](c)
			if err != nil || v["a"] != 1 {
				return fmt.Errorf("got %v", v)
			}
			return nil
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.fn(); err != nil {
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
						c.MustAdd(Build(func(c *Container) (L1, error) { v, _ := Inject[L0](c); return L1(v) + 1, nil }))
					case 2:
						c.MustAdd(Build(func(c *Container) (L2, error) { v, _ := Inject[L1](c); return L2(v) + 1, nil }))
					case 3:
						c.MustAdd(Build(func(c *Container) (L3, error) { v, _ := Inject[L2](c); return L3(v) + 1, nil }))
					case 4:
						c.MustAdd(Build(func(c *Container) (L4, error) { v, _ := Inject[L3](c); return L4(v) + 1, nil }))
					}
				}
				if _, err := Inject[L4](c); err != nil {
					t.Fatalf("iter %d: %v", iter, err)
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
	} else {
		t.Log(err)
	}
}

func TestBuildWithError(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Build(func(c *Container) (int, error) { return 0, fmt.Errorf("intentional error") }),
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
			return func(ctx context.Context) { startupCalled++ }
		}
		if _, ok := v.(Config); ok {
			return func(ctx context.Context) { startupCalled++ }
		}
		return nil
	})
	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		if _, ok := v.(Database); ok {
			return func(ctx context.Context) { shutdownCalled++ }
		}
		if _, ok := v.(Config); ok {
			return func(ctx context.Context) { shutdownCalled++ }
		}
		return nil
	})
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "test-app"}),
	)
	_, _ = Inject[Database](c)
	_, _ = Inject[Config](c)
	startup.Iterate(context.Background(), false)
	if startupCalled != 2 {
		t.Errorf("expected startupCalled=2, got %d", startupCalled)
	}
	shutdown.Iterate(context.Background(), true)
	if shutdownCalled != 2 {
		t.Errorf("expected shutdownCalled=2, got %d", shutdownCalled)
	}
}

func TestHook_OnceOnlyTriggered(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	callCount := 0
	hook := c.HookOnce("test", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			if _, ok := v.(Database); ok {
				callCount++
			}
		}
	})
	_, _ = Inject[Database](c)
	_, _ = Inject[Database](c)
	_, _ = Inject[Database](c)
	hook.Iterate(context.Background(), false)
	if callCount != 1 {
		t.Errorf("expected HookOnce called 1 time, got %d", callCount)
	}
}

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
	if !parentHookCalled {
		t.Error("expected parent hook to be called")
	}
}

func TestNestedContainer_CircularDependencyDetection(t *testing.T) {
	type ServiceA struct{}
	type ServiceB struct{}
	child := &Container{}
	child.MustAdd(Build(func(c *Container) (ServiceA, error) {
		_, err := Inject[ServiceB](c)
		return ServiceA{}, err
	}))
	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(c *Container) (ServiceB, error) {
		_, err := Inject[ServiceA](c)
		return ServiceB{}, err
	}))
	_, err := Inject[ServiceA](parent)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
}

func TestNestedContainer_TypeDetection(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		verifyFn func(*Container) error
	}{
		{
			name: "Basic_TwoLevel",
			setup: func() *Container {
				child := &Container{}
				child.MustAdd(Provide(CntTypeA{ID: 42}))
				parent := &Container{}
				parent.MustAdd(child)
				return parent
			},
			verifyFn: func(c *Container) error {
				v, err := Inject[CntTypeA](c)
				if err != nil {
					return err
				}
				if v.ID != 42 {
					return fmt.Errorf("expected ID=42, got %d", v.ID)
				}
				return nil
			},
		},
		{
			name: "ThreeLevel_Hierarchy",
			setup: func() *Container {
				l1 := &Container{}
				l1.MustAdd(Provide(CntTypeA{ID: 1}))
				l2 := &Container{}
				l2.MustAdd(l1)
				l3 := &Container{}
				l3.MustAdd(l2)
				return l3
			},
			verifyFn: func(c *Container) error {
				v, err := Inject[CntTypeA](c)
				if err != nil {
					return err
				}
				if v.ID != 1 {
					return fmt.Errorf("expected ID=1, got %d", v.ID)
				}
				return nil
			},
		},
		{
			name: "MultipleChildren_Isolation",
			setup: func() *Container {
				c1 := &Container{}
				c1.MustAdd(Provide(CntTypeA{ID: 1}))
				c2 := &Container{}
				c2.MustAdd(Provide(CntTypeB{ID: 2}))
				parent := &Container{}
				parent.MustAdd(c1, c2)
				return parent
			},
			verifyFn: func(c *Container) error {
				if v, err := Inject[CntTypeA](c); err != nil || v.ID != 1 {
					return fmt.Errorf("CntTypeA: %w", err)
				}
				if v, err := Inject[CntTypeB](c); err != nil || v.ID != 2 {
					return fmt.Errorf("CntTypeB: %w", err)
				}
				return nil
			},
		},
		{
			name: "DuplicatePrevention",
			setup: func() *Container {
				child := &Container{}
				child.MustAdd(Provide(CntTypeA{ID: 1}))
				parent := &Container{}
				parent.MustAdd(child)
				err := parent.Add(Provide(CntTypeA{ID: 2}))
				if err == nil {
					t.Error("expected duplicate type error")
				}
				return parent
			},
			verifyFn: func(c *Container) error {
				v, err := Inject[CntTypeA](c)
				if err != nil {
					return err
				}
				if v.ID != 1 {
					return fmt.Errorf("expected ID=1 from child, got %d", v.ID)
				}
				return nil
			},
		},
		{
			name: "MixedProvideAndBuild",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(CntTypeA{ID: 1}),
					Build(func(c *Container) (CntTypeB, error) {
						a, _ := Inject[CntTypeA](c)
						return CntTypeB{ID: a.ID * 10}, nil
					}),
				)
				return c
			},
			verifyFn: func(c *Container) error {
				if v, err := Inject[CntTypeB](c); err != nil || v.ID != 10 {
					return fmt.Errorf("CntTypeB: %w", err)
				}
				return nil
			},
		},
		{
			name: "ConcurrentAccess_ThreadSafe",
			setup: func() *Container {
				child := &Container{}
				child.MustAdd(Provide(CntTypeA{ID: 42}))
				parent := &Container{}
				parent.MustAdd(child)
				return parent
			},
			verifyFn: func(c *Container) error {
				var wg sync.WaitGroup
				errors := make(chan error, 100)
				for i := 0; i < 100; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						v, err := Inject[CntTypeA](c)
						if err != nil {
							errors <- err
							return
						}
						if v.ID != 42 {
							errors <- fmt.Errorf("expected ID=42, got %d", v.ID)
							return
						}
						errors <- nil
					}()
				}
				wg.Wait()
				close(errors)
				for err := range errors {
					if err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			if tt.verifyFn != nil {
				if err := tt.verifyFn(c); err != nil {
					t.Errorf("verify failed: %v", err)
				}
			}
		})
	}
}

func TestBuild_TableDriven(t *testing.T) {
	type (
		Name     string
		Greeting string
		S        string
		I        int
		R        string
		A        int
		B        int
		Named    struct{}
		Alias    = struct{}
	)

	tests := []struct {
		name     string
		setup    func() *Container
		injectFn func(*Container) (any, error)
		want     any
		wantErr  bool
	}{
		{
			name: "single dependency auto-inject",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide("input"), Build(func(s string) (int, error) { return len(s), nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int](c) },
			want:     5,
			wantErr:  false,
		},
		{
			name: "struct{} literal no dependency",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Build(func(struct{}) (int, error) { return 42, nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int](c) },
			want:     42,
			wantErr:  false,
		},
		{
			name: "struct{} type alias",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Build(func(Alias) (int32, error) { return 100, nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int32](c) },
			want:     int32(100),
			wantErr:  false,
		},
		{
			name: "var struct field inject",
			setup: func() *Container {
				c := &Container{}
				var s struct{}
				c.MustAdd(Provide(s), Build(func(struct{}) (int16, error) { return 7, nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int16](c) },
			want:     int16(7),
			wantErr:  false,
		},
		{
			name: "named struct requires Provide",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(Named{}), Build(func(Named) (int64, error) { return 99, nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int64](c) },
			want:     int64(99),
			wantErr:  false,
		},
		{
			name: "Container multi-dep",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(A(10)), Provide(B(20)),
					Build(func(c *Container) (int, error) {
						a, _ := Inject[A](c)
						b, _ := Inject[B](c)
						return int(a) + int(b), nil
					}),
				)
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int](c) },
			want:     30,
			wantErr:  false,
		},
		{
			name: "Container complex",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(Name("world")),
					Build(func(c *Container) (Greeting, error) {
						n, _ := Inject[Name](c)
						return Greeting("hello " + string(n)), nil
					}),
				)
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[Greeting](c) },
			want:     Greeting("hello world"),
			wantErr:  false,
		},
		{
			name: "dependency chain",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(
					Provide(S("hello")),
					Build(func(s S) (I, error) { return I(len(s)), nil }),
					Build(func(n I) (R, error) { return R(fmt.Sprintf("len:%d", n)), nil }),
				)
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[R](c) },
			want:     R("len:5"),
			wantErr:  false,
		},
		{
			name: "build error",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide("input"), Build(func(s string) (int, error) { return 0, fmt.Errorf("error") }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int](c) },
			want:     0,
			wantErr:  true,
		},
		{
			name: "dependency not found",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Build(func(s string) (int, error) { return len(s), nil }))
				return c
			},
			injectFn: func(c *Container) (any, error) { return Inject[int](c) },
			want:     0,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			got, err := tt.injectFn(c)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("got = %v, want = %v", got, tt.want)
			}
		})
	}
}

func TestInjectAs(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	db := Database{}
	err := InjectAs(c, &db)
	if err != nil {
		t.Fatal(err)
	}
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestMustInjectAs(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	db := Database{}
	MustInjectAs(c, &db)
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestMustInjectTo(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	var db Database
	MustInjectTo(c, &db)
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestMustInject(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	db := MustInject[Database](c)
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestMustInject_Panic(t *testing.T) {
	c := &Container{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustInject[Database](c)
}

func TestMustInjectAs_Panic(t *testing.T) {
	c := &Container{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	db := Database{}
	MustInjectAs(c, &db)
}

func TestMustInjectTo_Panic(t *testing.T) {
	c := &Container{}
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	var db Database
	MustInjectTo(c, &db)
}

func TestContainer_Inject(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	var db Database
	err := c.Inject(&db)
	if err != nil {
		t.Fatal(err)
	}
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestContainer_Inject_Multiple(t *testing.T) {
	c := &Container{}
	c.MustAdd(
		Provide(Database{DSN: "mysql://localhost"}),
		Provide(Config{AppName: "test-app"}),
	)
	var db Database
	var cfg Config
	err := c.Inject(&db, &cfg)
	if err != nil {
		t.Fatal(err)
	}
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
	if cfg.AppName != "test-app" {
		t.Errorf("expected test-app, got %s", cfg.AppName)
	}
}

func TestContainer_Inject_Error(t *testing.T) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
	var cfg Config
	err := c.Inject(&cfg)
	if err == nil {
		t.Fatal("expected error for missing type")
	}
}

func TestContainer_Concurrent(t *testing.T) {
	tests := []struct {
		name        string
		setup       func() *Container
		operateFn   func(*Container, int, *sync.WaitGroup, chan<- error)
		verifyFn    func(*Container) error
		goroutines  int
		wantSuccess int
	}{
		{
			name:  "Add_SameType_DuplicatePrevention",
			setup: func() *Container { return &Container{} },
			operateFn: func(c *Container, id int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errCh <- c.Add(Provide(CntTypeA{ID: id}))
				}()
			},
			verifyFn: func(c *Container) error {
				_, err := Inject[CntTypeA](c)
				return err
			},
			goroutines:  100,
			wantSuccess: 1,
		},
		{
			name:  "Add_DifferentTypes_AllRegistered",
			setup: func() *Container { return &Container{} },
			operateFn: func(c *Container, _ int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					errCh <- c.Add(
						Provide(bench0{Val: rand.Int()}),
						Provide(bench1{Val: rand.Int()}),
						Provide(bench2{Val: rand.Int()}),
						Provide(bench3{Val: rand.Int()}),
						Provide(bench4{Val: rand.Int()}),
					)
				}()
			},
			verifyFn: func(c *Container) error {
				if _, err := Inject[bench0](c); err != nil {
					return fmt.Errorf("bench0: %w", err)
				}
				if _, err := Inject[bench1](c); err != nil {
					return fmt.Errorf("bench1: %w", err)
				}
				if _, err := Inject[bench2](c); err != nil {
					return fmt.Errorf("bench2: %w", err)
				}
				if _, err := Inject[bench3](c); err != nil {
					return fmt.Errorf("bench3: %w", err)
				}
				if _, err := Inject[bench4](c); err != nil {
					return fmt.Errorf("bench4: %w", err)
				}
				return nil
			},
			goroutines:  50,
			wantSuccess: 1,
		},
		{
			name:  "Nested_SingleLevel_DuplicatePrevention",
			setup: func() *Container { return &Container{} },
			operateFn: func(c *Container, id int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeA{ID: id}))
					errCh <- c.Add(child)
				}()
			},
			verifyFn: func(c *Container) error {
				_, err := Inject[CntTypeA](c)
				return err
			},
			goroutines:  50,
			wantSuccess: 1,
		},
		{
			name:  "Nested_MultipleChildren_ThreeTypes",
			setup: func() *Container { return &Container{} },
			operateFn: func(c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					child := &Container{}
					switch idx % 3 {
					case 0:
						child.MustAdd(Provide(CntTypeA{ID: idx}))
					case 1:
						child.MustAdd(Provide(CntTypeB{ID: idx}))
					case 2:
						child.MustAdd(Provide(CntTypeC{ID: idx}))
					}
					errCh <- c.Add(child)
				}()
			},
			verifyFn: func(c *Container) error {
				if _, err := Inject[CntTypeA](c); err != nil {
					return fmt.Errorf("CntTypeA: %w", err)
				}
				if _, err := Inject[CntTypeB](c); err != nil {
					return fmt.Errorf("CntTypeB: %w", err)
				}
				if _, err := Inject[CntTypeC](c); err != nil {
					return fmt.Errorf("CntTypeC: %w", err)
				}
				return nil
			},
			goroutines:  30,
			wantSuccess: 3,
		},
		{
			name:  "Mixed_DirectAndNested",
			setup: func() *Container { return &Container{} },
			operateFn: func(c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					switch idx % 2 {
					case 0:
						errCh <- c.Add(Provide(CntTypeA{ID: idx}))
					case 1:
						child := &Container{}
						child.MustAdd(Provide(CntTypeB{ID: idx}))
						errCh <- c.Add(child)
					}
				}()
			},
			verifyFn: func(c *Container) error {
				if _, err := Inject[CntTypeA](c); err != nil {
					return err
				}
				if _, err := Inject[CntTypeB](c); err != nil {
					return err
				}
				return nil
			},
			goroutines:  40,
			wantSuccess: 2,
		},
		{
			name: "Inject_ThreadSafe",
			setup: func() *Container {
				c := &Container{}
				c.MustAdd(Provide(CntTypeA{ID: 42}))
				return c
			},
			operateFn: func(c *Container, _ int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(1)
				go func() {
					defer wg.Done()
					v, err := Inject[CntTypeA](c)
					if err != nil {
						errCh <- err
						return
					}
					if v.ID != 42 {
						errCh <- fmt.Errorf("expected ID=42, got %d", v.ID)
						return
					}
					errCh <- nil
				}()
			},
			verifyFn:    nil,
			goroutines:  100,
			wantSuccess: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.setup()
			var wg sync.WaitGroup
			errors := make(chan error, tt.goroutines)
			for i := 0; i < tt.goroutines; i++ {
				tt.operateFn(c, i, &wg, errors)
			}
			wg.Wait()
			close(errors)
			successCount := 0
			for err := range errors {
				if err == nil {
					successCount++
				}
			}
			if tt.wantSuccess > 0 && successCount != tt.wantSuccess {
				t.Errorf("expected %d successes, got %d", tt.wantSuccess, successCount)
			}
			if tt.verifyFn != nil {
				if err := tt.verifyFn(c); err != nil {
					t.Errorf("verify failed: %v", err)
				}
			}
		})
	}
}

func TestContainer_ConcurrentMustAdd(t *testing.T) {
	c := &Container{}
	const count = 50
	var wg sync.WaitGroup
	panicCount := 0
	var mu sync.Mutex
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					panicCount++
					mu.Unlock()
				}
			}()
			c.MustAdd(Provide(CntTypeA{ID: id}))
		}(i)
	}
	wg.Wait()
	if panicCount < 1 {
		t.Errorf("expected at least 1 panic, got %d", panicCount)
	}
}

func TestContainer_ConcurrentCrossNested(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() (*Container, *Container, *Container)
		operateFn  func(*Container, *Container, *Container, int, *sync.WaitGroup, chan<- error)
		verifyFn   func(*Container, *Container, *Container) error
		goroutines int
	}{
		{
			name: "ABC_Separate",
			setup: func() (*Container, *Container, *Container) {
				return &Container{}, &Container{}, &Container{}
			},
			operateFn: func(a, b, c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(3)
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeA{ID: idx}))
					errCh <- a.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeB{ID: idx}))
					errCh <- b.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeC{ID: idx}))
					errCh <- c.Add(child)
				}()
			},
			verifyFn: func(a, b, c *Container) error {
				if _, err := Inject[CntTypeA](a); err != nil {
					return fmt.Errorf("CntTypeA: %w", err)
				}
				if _, err := Inject[CntTypeB](b); err != nil {
					return fmt.Errorf("CntTypeB: %w", err)
				}
				if _, err := Inject[CntTypeC](c); err != nil {
					return fmt.Errorf("CntTypeC: %w", err)
				}
				return nil
			},
			goroutines: 30,
		},
		{
			name: "ABC_CrossAdd",
			setup: func() (*Container, *Container, *Container) {
				return &Container{}, &Container{}, &Container{}
			},
			operateFn: func(a, b, c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(3)
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeA{ID: idx}))
					errCh <- b.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeB{ID: idx}))
					errCh <- c.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeC{ID: idx}))
					errCh <- a.Add(child)
				}()
			},
			verifyFn: func(a, b, c *Container) error {
				if _, err := Inject[CntTypeA](b); err != nil {
					return fmt.Errorf("CntTypeA in B: %w", err)
				}
				if _, err := Inject[CntTypeB](c); err != nil {
					return fmt.Errorf("CntTypeB in C: %w", err)
				}
				if _, err := Inject[CntTypeC](a); err != nil {
					return fmt.Errorf("CntTypeC in A: %w", err)
				}
				return nil
			},
			goroutines: 20,
		},
		{
			name: "TwoLevel_Cross",
			setup: func() (*Container, *Container, *Container) {
				return &Container{}, &Container{}, &Container{}
			},
			operateFn: func(a, b, c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(3)
				go func() {
					defer wg.Done()
					l1 := &Container{}
					l1.MustAdd(Provide(CntTypeA{ID: idx}))
					l2 := &Container{}
					l2.MustAdd(l1)
					errCh <- b.Add(l2)
				}()
				go func() {
					defer wg.Done()
					l1 := &Container{}
					l1.MustAdd(Provide(CntTypeB{ID: idx}))
					l2 := &Container{}
					l2.MustAdd(l1)
					errCh <- c.Add(l2)
				}()
				go func() {
					defer wg.Done()
					l1 := &Container{}
					l1.MustAdd(Provide(CntTypeC{ID: idx}))
					l2 := &Container{}
					l2.MustAdd(l1)
					errCh <- a.Add(l2)
				}()
			},
			verifyFn: func(a, b, c *Container) error {
				if _, err := Inject[CntTypeA](b); err != nil {
					return fmt.Errorf("CntTypeA in B: %w", err)
				}
				if _, err := Inject[CntTypeB](c); err != nil {
					return fmt.Errorf("CntTypeB in C: %w", err)
				}
				if _, err := Inject[CntTypeC](a); err != nil {
					return fmt.Errorf("CntTypeC in A: %w", err)
				}
				return nil
			},
			goroutines: 15,
		},
		{
			name: "Diamond_FourContainers",
			setup: func() (*Container, *Container, *Container) {
				return &Container{}, &Container{}, &Container{}
			},
			operateFn: func(a, b, c *Container, idx int, wg *sync.WaitGroup, errCh chan<- error) {
				wg.Add(4)
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeA{ID: idx}))
					errCh <- a.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeB{ID: idx}))
					errCh <- b.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeC{ID: idx}))
					errCh <- c.Add(child)
				}()
				go func() {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeD{ID: idx}))
					errCh <- a.Add(child)
				}()
			},
			verifyFn: func(a, b, c *Container) error {
				if _, err := Inject[CntTypeA](a); err != nil {
					return fmt.Errorf("CntTypeA: %w", err)
				}
				if _, err := Inject[CntTypeB](b); err != nil {
					return fmt.Errorf("CntTypeB: %w", err)
				}
				if _, err := Inject[CntTypeC](c); err != nil {
					return fmt.Errorf("CntTypeC: %w", err)
				}
				if _, err := Inject[CntTypeD](a); err != nil {
					return fmt.Errorf("CntTypeD: %w", err)
				}
				return nil
			},
			goroutines: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b, c := tt.setup()
			var wg sync.WaitGroup
			errors := make(chan error, tt.goroutines*10)
			for i := 0; i < tt.goroutines; i++ {
				tt.operateFn(a, b, c, i, &wg, errors)
			}
			wg.Wait()
			close(errors)
			successCount := 0
			for err := range errors {
				if err == nil {
					successCount++
				}
			}
			t.Logf("successful adds: %d", successCount)
			if tt.verifyFn != nil {
				if err := tt.verifyFn(a, b, c); err != nil {
					t.Errorf("verify failed: %v", err)
				}
			}
		})
	}
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

func BenchmarkInjectAs(b *testing.B) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "test"}))
	b.ReportAllocs()
	b.ResetTimer()
	var db Database
	for i := 0; i < b.N; i++ {
		_ = InjectAs(c, &db)
	}
}

func BenchmarkContainer_Add(b *testing.B) {
	b.Run("10Providers", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c := &Container{}
			c.MustAdd(
				Provide(bench0{i}), Provide(bench1{i}), Provide(bench2{i}),
				Provide(bench3{i}), Provide(bench4{i}), Provide(bench5{i}),
				Provide(bench6{i}), Provide(bench7{i}), Provide(bench8{i}),
				Provide(bench9{i}),
			)
		}
	})
}

func BenchmarkContainer_ConcurrentAdd(b *testing.B) {
	b.Run("SameType", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			c := &Container{}
			var wg sync.WaitGroup
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					_ = c.Add(Provide(CntTypeA{ID: id}))
				}(j)
			}
			wg.Wait()
		}
	})
}

func BenchmarkContainer_ConcurrentNestedAdd(b *testing.B) {
	b.Run("TwoLevels", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			parent := &Container{}
			var wg sync.WaitGroup
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					child := &Container{}
					child.MustAdd(Provide(CntTypeA{ID: id}))
					_ = parent.Add(child)
				}(j)
			}
			wg.Wait()
		}
	})
}

// TestNestedContainer_CrossContainer_ChainDependency tests dependency chain across 3-level nested containers
// Container hierarchy:
//
//	parent: A, D, F
//	child:  B, E
//	grand:  C, G
//
// Chain: A => B => C => D => E => F => G (should work)
func TestNestedContainer_CrossContainer_ChainDependency(t *testing.T) {
	type A struct{ BVal string }
	type B struct{ CVal string }
	type C struct{ DVal string }
	type D struct{ EVal string }
	type E struct{ FVal string }
	type F struct{ GVal string }
	type G struct{ Value string }

	// Grandchild container: provides C, G
	grand := &Container{}
	grand.MustAdd(Provide(G{Value: "G-base"}))
	grand.MustAdd(Build(func(d D) (C, error) {
		return C{DVal: "C-depends-on-" + d.EVal}, nil
	}))

	// Child container: provides B, E
	child := &Container{}
	child.MustAdd(grand)
	child.MustAdd(Build(func(cc C) (B, error) {
		return B{CVal: "B-depends-on-" + cc.DVal}, nil
	}))
	child.MustAdd(Build(func(f F) (E, error) {
		return E{FVal: "E-depends-on-" + f.GVal}, nil
	}))

	// Parent container: provides A, D, F
	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(bb B) (A, error) {
		return A{BVal: "A-depends-on-" + bb.CVal}, nil
	}))
	parent.MustAdd(Build(func(e E) (D, error) {
		return D{EVal: "D-depends-on-" + e.FVal}, nil
	}))
	parent.MustAdd(Build(func(g G) (F, error) {
		return F{GVal: "F-depends-on-" + g.Value}, nil
	}))

	// Inject A - should traverse: A(parent) => B(child) => C(grand) => D(parent) => E(child) => F(parent) => G(grand)
	a, err := Inject[A](parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the full chain was resolved
	expected := "A-depends-on-B-depends-on-C-depends-on-D-depends-on-E-depends-on-F-depends-on-G-base"
	if a.BVal != expected {
		t.Errorf("expected %q, got %q", expected, a.BVal)
	}
}

// TestNestedContainer_CrossContainer_CircularDependency tests circular dependency detection across 3-level nested containers
// Container hierarchy:
//
//	parent: A, D, F
//	child:  B, E
//	grand:  C, G
//
// Chain: A => B => C => D => E => F => G => A (circular - should fail)
func TestNestedContainer_CrossContainer_CircularDependency(t *testing.T) {
	type A struct{ GVal string }
	type B struct{ CVal string }
	type C struct{ DVal string }
	type D struct{ EVal string }
	type E struct{ FVal string }
	type F struct{ GVal string }
	type G struct{ AVal string }

	// Grandchild container: provides C, G (G depends on A - creates circular!)
	grand := &Container{}
	grand.MustAdd(Build(func(c *Container) (G, error) {
		// This creates the circular dependency: G => A
		a, err := Inject[A](c)
		if err != nil {
			return G{}, err
		}
		return G{AVal: "G-depends-on-" + a.GVal}, nil
	}))
	grand.MustAdd(Build(func(c *Container) (C, error) {
		d, err := Inject[D](c)
		if err != nil {
			return C{}, err
		}
		return C{DVal: "C-depends-on-" + d.EVal}, nil
	}))

	// Child container: provides B, E
	child := &Container{}
	child.MustAdd(grand)
	child.MustAdd(Build(func(c *Container) (B, error) {
		cc, err := Inject[C](c)
		if err != nil {
			return B{}, err
		}
		return B{CVal: "B-depends-on-" + cc.DVal}, nil
	}))
	child.MustAdd(Build(func(c *Container) (E, error) {
		f, err := Inject[F](c)
		if err != nil {
			return E{}, err
		}
		return E{FVal: "E-depends-on-" + f.GVal}, nil
	}))

	// Parent container: provides A, D, F
	parent := &Container{}
	parent.MustAdd(child)
	parent.MustAdd(Build(func(c *Container) (A, error) {
		bb, err := Inject[B](c)
		if err != nil {
			return A{}, err
		}
		return A{GVal: "A-depends-on-" + bb.CVal}, nil
	}))
	parent.MustAdd(Build(func(e E) (D, error) {
		return D{EVal: "D-depends-on-" + e.FVal}, nil
	}))
	parent.MustAdd(Build(func(c *Container) (F, error) {
		g, err := Inject[G](c)
		if err != nil {
			return F{}, err
		}
		return F{GVal: "F-depends-on-" + g.AVal}, nil
	}))

	// Should detect circular dependency
	_, err := Inject[A](parent)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	if !contains(err.Error(), "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %v", err)
	}
	t.Log(err)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
