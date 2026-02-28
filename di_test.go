package di

import (
	"fmt"
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
	t.Run("add unique provider", func(t *testing.T) {
		c := &Container{}
		err := c.Add(Provide(Database{DSN: "mysql://localhost"}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("add duplicate provider returns error", func(t *testing.T) {
		c := &Container{}
		_ = c.Add(Provide(Database{DSN: "mysql://localhost"}))
		err := c.Add(Provide(Database{DSN: "mysql://remote"}))
		if err == nil {
			t.Fatal("expected error for duplicate provider")
		}
	})
}

func TestInject(t *testing.T) {
	t.Run("inject single dependency", func(t *testing.T) {
		c := &Container{}
		_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))

		db, err := Inject[Database](c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db.DSN != "mysql://localhost:3306/test" {
			t.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
		}
	})

	t.Run("inject non-existent provider returns error", func(t *testing.T) {
		c := &Container{}
		db, err := Inject[Database](c)
		if err == nil {
			t.Fatal("expected error for non-existent provider")
		}
		var zero Database
		if db != zero {
			t.Errorf("expected zero value, got %v", db)
		}
	})
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
	err := InjectTo(c, &db)
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
