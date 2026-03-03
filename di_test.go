package godi

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// Common test types
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

// =============================================================================
// Core Functionality Tests - Combined complex scenarios
// =============================================================================

func TestCore_ProvideBuildInject(t *testing.T) {
	t.Run("BasicProvideAndInject", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		db, err := Inject[Database](c)
		if err != nil || db.DSN != "mysql://localhost" {
			t.Errorf("expected mysql://localhost, got %v, err %v", db.DSN, err)
		}
	})

	t.Run("BuildWithDependencies", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide(Database{DSN: "mysql://localhost:3306/test"}),
			Provide(Config{AppName: "test-app"}),
			Build(func(c *Container) (Service, error) {
				db, _ := Inject[Database](c)
				cfg, _ := Inject[Config](c)
				return Service{Name: "svc", DB: db, Cfg: cfg}, nil
			}),
		)
		svc, err := Inject[Service](c)
		if err != nil {
			t.Fatal(err)
		}
		if svc.DB.DSN != "mysql://localhost:3306/test" || svc.Cfg.AppName != "test-app" {
			t.Errorf("unexpected service: %+v", svc)
		}
	})

	t.Run("BuildWithSingleDependency", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Provide("input"),
			Build(func(s string) (int, error) { return len(s), nil }),
		)
		v, err := Inject[int](c)
		if err != nil || v != 5 {
			t.Errorf("expected 5, got %v, err %v", v, err)
		}
	})

	t.Run("AllSupportedTypes", func(t *testing.T) {
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
	})
}

func TestCore_ErrorsAndPanics(t *testing.T) {
	t.Run("DuplicateProviderError", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		err := c.Add(Provide(Database{DSN: "mysql://remote"}))
		if err == nil {
			t.Error("expected duplicate error")
		}
	})

	t.Run("MustAddPanicsOnDuplicate", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		c.MustAdd(Provide(Database{DSN: "mysql://remote"}))
	})

	t.Run("InjectNotFound", func(t *testing.T) {
		c := &Container{}
		var db Database
		err := InjectTo[Database](c, &db)
		if err == nil {
			t.Error("expected error for missing type")
		}
	})

	t.Run("MustInjectPanics", func(t *testing.T) {
		c := &Container{}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		MustInject[Database](c)
	})

	t.Run("BuildErrorPropagation", func(t *testing.T) {
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
		_, err := Inject[string](c)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "wrapped") {
			t.Errorf("expected wrapped error, got: %v", err)
		}
	})

	t.Run("CircularDependencyDetection", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Build(func(c *Container) (int, error) {
				_, err := Inject[string](c)
				return 0, err
			}),
			Build(func(c *Container) (string, error) {
				_, err := Inject[int](c)
				return "", err
			}),
		)
		_, err := Inject[int](c)
		if err == nil {
			t.Fatal("expected circular dependency error")
		}
		if !strings.Contains(err.Error(), "circular dependency") {
			t.Errorf("expected 'circular dependency' error, got: %v", err)
		}
	})

	t.Run("BuildPanicRecovery", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(
			Build(func(c *Container) (string, error) {
				i, err := Inject[int](c)
				return fmt.Sprintf("%v", i), err
			}),
			Build(func(c *Container) (int, error) {
				i, _ := strconv.Atoi(MustInject[string](c))
				return i, nil
			}),
		)
		_, err := Inject[string](c)
		if err == nil {
			t.Error("expected panic recovery error")
		}
	})

	t.Run("MustInjectAsPanics", func(t *testing.T) {
		c := &Container{}
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		var db Database
		MustInjectAs(c, &db)
	})

	t.Run("InjectMultipleWithError", func(t *testing.T) {
		c := &Container{}
		c.MustAdd(Provide(Database{DSN: "test"}))
		var db Database
		var cfg Config
		err := c.Inject(&db, &cfg)
		if err == nil {
			t.Error("expected error when injecting missing type")
		}
	})
}

func TestCore_ContainerOperations(t *testing.T) {
	t.Run("MultiInject", func(t *testing.T) {
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
		if db.DSN != "mysql://localhost" || cfg.AppName != "test-app" {
			t.Errorf("unexpected values: db=%v, cfg=%v", db, cfg)
		}
	})

	t.Run("InjectAs", func(t *testing.T) {
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
	})

	t.Run("FrozenContainer", func(t *testing.T) {
		child := &Container{}
		child.MustAdd(Provide(Database{DSN: "mysql://localhost"}))
		parent := &Container{}
		parent.MustAdd(child)
		err := child.Add(Provide(Config{AppName: "test"}))
		if err == nil {
			t.Error("expected frozen error")
		}
		if !strings.Contains(err.Error(), "frozen") {
			t.Errorf("expected 'frozen' error, got: %v", err)
		}
	})
}

// =============================================================================
// Hook System Tests
// =============================================================================

func TestHooks_Lifecycle(t *testing.T) {
	t.Run("HookAndHookOnce", func(t *testing.T) {
		c := &Container{}
		startupCalled := 0
		shutdownCalled := 0

		startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
			if provided > 0 {
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
			return func(ctx context.Context) {
				if _, ok := v.(Database); ok {
					shutdownCalled++
				}
				if _, ok := v.(Config); ok {
					shutdownCalled++
				}
			}
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
	})

	t.Run("HookOnceOnlyTriggered", func(t *testing.T) {
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
	})

	t.Run("NestedContainerHooks", func(t *testing.T) {
		type DB struct{ DSN string }
		child := &Container{}
		child.MustAdd(Provide(DB{DSN: "mysql://localhost"}))
		childHookCalled := false
		childStartup := child.Hook("startup", func(v any, provided int) func(context.Context) {
			if provided > 0 {
				return nil
			}
			return func(ctx context.Context) {
				if _, ok := v.(DB); ok {
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
				if _, ok := v.(DB); ok {
					parentHookCalled = true
				}
			}
		})
		_, _ = Inject[DB](parent)
		childStartup.Iterate(context.Background(), false)
		parentStartup.Iterate(context.Background(), false)
		if !childHookCalled {
			t.Error("expected child hook to be called")
		}
		if !parentHookCalled {
			t.Error("expected parent hook to be called")
		}
	})
}

// =============================================================================
// Nested Container Tests
// =============================================================================

func TestNestedContainer_Basic(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Container
		verifyFn func(*Container) error
	}{
		{
			name: "TwoLevel",
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
			name: "ThreeLevel",
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
			name: "MultipleChildren",
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
			name: "ConcurrentAccess",
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

func TestNestedContainer_CircularDependency(t *testing.T) {
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

func TestNestedContainer_CrossContainer_ChainDependency(t *testing.T) {
	type A struct{ BVal string }
	type B struct{ CVal string }
	type C struct{ DVal string }
	type D struct{ EVal string }
	type E struct{ FVal string }
	type F struct{ GVal string }
	type G struct{ Value string }

	grand := &Container{}
	grand.MustAdd(Provide(G{Value: "G-base"}))
	grand.MustAdd(Build(func(d D) (C, error) {
		return C{DVal: "C-depends-on-" + d.EVal}, nil
	}))

	child := &Container{}
	child.MustAdd(grand)
	child.MustAdd(Build(func(cc C) (B, error) {
		return B{CVal: "B-depends-on-" + cc.DVal}, nil
	}))
	child.MustAdd(Build(func(f F) (E, error) {
		return E{FVal: "E-depends-on-" + f.GVal}, nil
	}))

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

	a, err := Inject[A](parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "A-depends-on-B-depends-on-C-depends-on-D-depends-on-E-depends-on-F-depends-on-G-base"
	if a.BVal != expected {
		t.Errorf("expected %q, got %q", expected, a.BVal)
	}
}

func TestNestedContainer_CrossContainer_CircularDependency(t *testing.T) {
	type A struct{ GVal string }
	type B struct{ CVal string }
	type C struct{ DVal string }
	type D struct{ EVal string }
	type E struct{ FVal string }
	type F struct{ GVal string }
	type G struct{ AVal string }

	grand := &Container{}
	grand.MustAdd(Build(func(c *Container) (G, error) {
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

	_, err := Inject[A](parent)
	if err == nil {
		t.Fatal("expected circular dependency error")
	}
	if !strings.Contains(err.Error(), "circular dependency") {
		t.Errorf("expected 'circular dependency' error, got: %v", err)
	}
}

// =============================================================================
// Runtime Container Add Tests
// =============================================================================

func TestNestedContainer_RuntimeAdd(t *testing.T) {
	tests := []struct {
		name         string
		initialFloat float64
		wantStr      string
		wantIntErr   bool
	}{
		{
			name:         "NegativeValue_AddsNegativeContainer",
			initialFloat: -0.1,
			wantStr:      "-100",
			wantIntErr:   true,
		},
		{
			name:         "PositiveValue_AddsPositiveContainer",
			initialFloat: 0.5,
			wantStr:      "100",
			wantIntErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Container{}
			positive := (&Container{}).MustAdd(Provide(100))
			negative := (&Container{}).MustAdd(Provide(-100))

			c.MustAdd(
				Provide(tt.initialFloat),
				Build(func(c *Container) (str string, err error) {
					f, err := Inject[float64](c)
					if err != nil {
						return
					}
					if f > 0 {
						c.MustAdd(positive)
					} else {
						c.MustAdd(negative)
					}
					i, err := Inject[int](c)
					if err != nil {
						return
					}
					return strconv.Itoa(i), nil
				}),
			)

			str, err := Inject[string](c)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if str != tt.wantStr {
				t.Errorf("got string %s, want %s", str, tt.wantStr)
			}

			if v, e := Inject[int](c); (e != nil) != tt.wantIntErr {
				t.Errorf("Inject[int] error = %v, wantErr %v", e, tt.wantIntErr)
			} else if e == nil {
				t.Logf("Injected int from nested container: %d", v)
			}
		})
	}
}

func TestNestedContainer_RuntimeAdd_NoFrozenError(t *testing.T) {
	t.Run("RuntimeAdd_DoesNotTriggerFrozen", func(t *testing.T) {
		c := &Container{}
		nested := (&Container{}).MustAdd(Provide("nested-value"))

		c.MustAdd(
			Provide(42),
			Build(func(c *Container) (string, error) {
				c.MustAdd(nested)
				i, _ := Inject[int](c)
				return fmt.Sprintf("val-%d", i), nil
			}),
		)

		if _, err := Inject[string](c); err != nil {
			t.Fatalf("Build function should not trigger frozen error: %v", err)
		}

		if v, err := Inject[string](c); err != nil {
			t.Errorf("Re-injection failed: %v", err)
		} else if v != "val-42" {
			t.Errorf("got %s, want val-42", v)
		}
	})

	t.Run("DirectAddToFrozenContainer_Error", func(t *testing.T) {
		child := &Container{}
		child.MustAdd(Provide("child-value"))

		parent := &Container{}
		parent.MustAdd(child)

		err := child.Add(Provide("new-value"))
		if err == nil {
			t.Fatal("expected frozen error, got nil")
		}
		if !strings.Contains(err.Error(), "frozen") {
			t.Errorf("expected 'frozen' error, got: %v", err)
		}
	})

	t.Run("RuntimeAdd_MultipleContainers", func(t *testing.T) {
		type V1 string
		type V2 string
		type V3 string

		c := &Container{}
		container1 := (&Container{}).MustAdd(Provide(V1("value1")))
		container2 := (&Container{}).MustAdd(Provide(V2("value2")))
		container3 := (&Container{}).MustAdd(Provide(V3("value3")))

		c.MustAdd(
			Build(func(c *Container) (string, error) {
				c.MustAdd(container1)
				c.MustAdd(container2)
				c.MustAdd(container3)
				v1, _ := Inject[V1](c)
				return string(v1), nil
			}),
		)

		v, err := Inject[string](c)
		if err != nil {
			t.Fatalf("Runtime add multiple containers failed: %v", err)
		}
		if v != "value1" {
			t.Errorf("got %s, want value1", v)
		}
	})

	t.Run("RuntimeAdd_NestedContainer_WithHooks", func(t *testing.T) {
		type RuntimeDB struct{ DSN string }

		c := &Container{}
		nested := (&Container{}).MustAdd(Provide(RuntimeDB{DSN: "mysql://runtime"}))

		hookCalled := 0
		hook := c.HookOnce("test", func(v any) func(context.Context) {
			return func(ctx context.Context) {
				if _, ok := v.(RuntimeDB); ok {
					hookCalled++
				}
			}
		})

		c.MustAdd(
			Build(func(c *Container) (string, error) {
				c.MustAdd(nested)
				db, _ := Inject[RuntimeDB](c)
				return db.DSN, nil
			}),
		)

		result, err := Inject[string](c)
		if err != nil {
			t.Fatalf("Runtime add with hook failed: %v", err)
		}
		if result != "mysql://runtime" {
			t.Errorf("got DSN %s, want mysql://runtime", result)
		}

		hook.Iterate(context.Background(), false)
		if hookCalled != 1 {
			t.Errorf("hook called %d times, want 1", hookCalled)
		}
	})

	t.Run("RuntimeAdd_ChainedDependencies", func(t *testing.T) {
		type A struct{ Val string }
		type B struct{ Val string }
		type C struct{ Val string }

		c := &Container{}
		containerA := (&Container{}).MustAdd(Provide(A{Val: "A"}))
		containerB := (&Container{}).MustAdd(
			Provide(B{Val: "B"}),
			Build(func(c *Container) (string, error) {
				a, _ := Inject[A](c)
				b, _ := Inject[B](c)
				return a.Val + "-" + b.Val, nil
			}),
		)

		c.MustAdd(
			Build(func(c *Container) (C, error) {
				c.MustAdd(containerA)
				c.MustAdd(containerB)
				s, _ := Inject[string](c)
				return C{Val: "C-" + s}, nil
			}),
		)

		result, err := Inject[C](c)
		if err != nil {
			t.Fatal(err)
		}
		expected := "C-A-B"
		if result.Val != expected {
			t.Errorf("got %s, want %s", result.Val, expected)
		}
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

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
			verifyFn:    func(c *Container) error { _, err := Inject[CntTypeA](c); return err },
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
				for _, fn := range []func() error{
					func() error { _, e := Inject[bench0](c); return e },
					func() error { _, e := Inject[bench1](c); return e },
					func() error { _, e := Inject[bench2](c); return e },
					func() error { _, e := Inject[bench3](c); return e },
					func() error { _, e := Inject[bench4](c); return e },
				} {
					if err := fn(); err != nil {
						return err
					}
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
			verifyFn:    func(c *Container) error { _, err := Inject[CntTypeA](c); return err },
			goroutines:  50,
			wantSuccess: 1,
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

// =============================================================================
// Build Dependency Order Tests
// =============================================================================

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

// =============================================================================
// Benchmark Tests
// =============================================================================

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

func BenchmarkInjectTo(b *testing.B) {
	c := &Container{}
	c.MustAdd(Provide(Database{DSN: "test"}))
	b.ReportAllocs()
	b.ResetTimer()
	var db Database
	for i := 0; i < b.N; i++ {
		_ = InjectTo(c, &db)
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
