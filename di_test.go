package godi

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

	newVal := provider.new()
	if newVal == nil {
		t.Fatal("expected non-nil value from new()")
	}
	if _, ok := newVal.(*Database); !ok {
		t.Errorf("expected *Database from new(), got %T", newVal)
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

func TestInject(t *testing.T) {
	t.Run("inject single dependency", func(t *testing.T) {
		c := &Container{}
		_ = c.Add(Provide(Database{DSN: "mysql://localhost:3306/test"}))

		db, err := ShouldInject[Database](c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if db.DSN != "mysql://localhost:3306/test" {
			t.Errorf("expected mysql://localhost:3306/test, got %s", db.DSN)
		}
	})

	t.Run("inject non-existent provider returns error", func(t *testing.T) {
		c := &Container{}
		db, err := ShouldInject[Database](c)
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
	err := ShouldInjectTo(&db, c)
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

func TestShouldInjectTo_ErrorCase(t *testing.T) {
	c := &Container{}
	var db Database
	err := ShouldInjectTo(&db, c)
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
	ok := InjectTo(&db, c)
	if !ok {
		t.Fatal("expected InjectTo to return true")
	}
	if db.DSN != "mysql://localhost" {
		t.Errorf("expected mysql://localhost, got %s", db.DSN)
	}
}

func TestInjectTo_WrongType(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "mysql://localhost"}))

	var cfg Config
	ok := InjectTo(&cfg, c)
	if ok {
		t.Fatal("expected InjectTo to return false for wrong type")
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
	expectedMsg := "provider godi.Database not found"
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
	ok := InjectTo(&db, c)
	fmt.Printf("Injected: %v, DSN: %s\n", ok, db.DSN)
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
	ok := InjectTo(&v, c)
	if !ok {
		t.Fatal("expected InjectTo to return true")
	}
	if v != "test value" {
		t.Errorf("expected 'test value', got %v", v)
	}
}

func TestShouldInjectTo_Success(t *testing.T) {
	c := &Container{}
	c.Add(Provide(Database{DSN: "test"}))

	var db Database
	err := ShouldInjectTo(&db, c)
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
	ok := InjectTo(&db, c1, c2)
	if !ok || db.DSN != "db1" {
		t.Errorf("expected db1, got %v", db)
	}

	var cfg Config
	ok = InjectTo(&cfg, c1, c2)
	if !ok || cfg.AppName != "app2" {
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
	ok = InjectTo(&cfg, c1, c2)
	if ok {
		t.Error("expected InjectTo to return false for non-existent provider")
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
