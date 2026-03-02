package main

import (
	"context"
	"fmt"

	"github.com/Just-maple/godi"
)

type Database struct {
	DSN string
}

type Cache struct {
	Addr string
}

type Logger struct {
	Level string
}

type Config struct {
	AppName string
}

type ServiceA struct {
	Name string
	DB   Database
}

type ServiceB struct {
	Name  string
	Cache Cache
	Log   Logger
}

func main() {
	fmt.Println("=== Multi-Level Nested Container Hook Mechanism Test ===\n")

	infra := &godi.Container{}
	infra.MustAdd(
		godi.Provide(Database{DSN: "mysql://localhost:3306/app"}),
		godi.Provide(Cache{Addr: "redis://localhost:6379"}),
	)

	infraInitCalls := make(map[string]int)
	infraCleanupCalls := make(map[string]int)

	infraInit := infra.Hook("init", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				infraInitCalls["Database"]++
				fmt.Printf("[Infra-Init] Database initialized (call #%d)\n", infraInitCalls["Database"])
			case Cache:
				infraInitCalls["Cache"]++
				fmt.Printf("[Infra-Init] Cache initialized (call #%d)\n", infraInitCalls["Cache"])
			}
		}
	})

	infraCleanup := infra.HookOnce("cleanup", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				infraCleanupCalls["Database"]++
				fmt.Printf("[Infra-Cleanup] Database cleaned up (call #%d)\n", infraCleanupCalls["Database"])
			case Cache:
				infraCleanupCalls["Cache"]++
				fmt.Printf("[Infra-Cleanup] Cache cleaned up (call #%d)\n", infraCleanupCalls["Cache"])
			}
		}
	})

	services := &godi.Container{}
	services.MustAdd(
		infra,
		godi.Provide(Logger{Level: "info"}),
	)

	servicesInitCalls := make(map[string]int)
	servicesCleanupCalls := make(map[string]int)

	servicesInit := services.Hook("init", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				servicesInitCalls["Database"]++
				fmt.Printf("[Services-Init] Database registered (call #%d)\n", servicesInitCalls["Database"])
			case Cache:
				servicesInitCalls["Cache"]++
				fmt.Printf("[Services-Init] Cache registered (call #%d)\n", servicesInitCalls["Cache"])
			case Logger:
				servicesInitCalls["Logger"]++
				fmt.Printf("[Services-Init] Logger registered (call #%d)\n", servicesInitCalls["Logger"])
			}
		}
	})

	servicesCleanup := services.HookOnce("cleanup", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				servicesCleanupCalls["Database"]++
				fmt.Printf("[Services-Cleanup] Database cleanup (call #%d)\n", servicesCleanupCalls["Database"])
			case Cache:
				servicesCleanupCalls["Cache"]++
				fmt.Printf("[Services-Cleanup] Cache cleanup (call #%d)\n", servicesCleanupCalls["Cache"])
			case Logger:
				servicesCleanupCalls["Logger"]++
				fmt.Printf("[Services-Cleanup] Logger cleanup (call #%d)\n", servicesCleanupCalls["Logger"])
			}
		}
	})

	app := &godi.Container{}
	app.MustAdd(
		services,
		godi.Provide(Config{AppName: "nested-hook-demo"}),
		godi.Build(func(c *godi.Container) (ServiceA, error) {
			db, _ := godi.Inject[Database](c)
			return ServiceA{Name: "service-a", DB: db}, nil
		}),
		godi.Build(func(c *godi.Container) (ServiceB, error) {
			cache, _ := godi.Inject[Cache](c)
			log, _ := godi.Inject[Logger](c)
			return ServiceB{Name: "service-b", Cache: cache, Log: log}, nil
		}),
	)

	appInitCalls := make(map[string]int)
	appCleanupCalls := make(map[string]int)

	appInit := app.Hook("init", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				appInitCalls["Database"]++
				fmt.Printf("[App-Init] Database provided (call #%d)\n", appInitCalls["Database"])
			case Cache:
				appInitCalls["Cache"]++
				fmt.Printf("[App-Init] Cache provided (call #%d)\n", appInitCalls["Cache"])
			case Logger:
				appInitCalls["Logger"]++
				fmt.Printf("[App-Init] Logger provided (call #%d)\n", appInitCalls["Logger"])
			case Config:
				appInitCalls["Config"]++
				fmt.Printf("[App-Init] Config provided (call #%d)\n", appInitCalls["Config"])
			case ServiceA:
				appInitCalls["ServiceA"]++
				fmt.Printf("[App-Init] ServiceA provided (call #%d)\n", appInitCalls["ServiceA"])
			case ServiceB:
				appInitCalls["ServiceB"]++
				fmt.Printf("[App-Init] ServiceB provided (call #%d)\n", appInitCalls["ServiceB"])
			}
		}
	})

	appCleanup := app.HookOnce("cleanup", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			switch v.(type) {
			case Database:
				appCleanupCalls["Database"]++
				fmt.Printf("[App-Cleanup] Database cleanup (call #%d)\n", appCleanupCalls["Database"])
			case Cache:
				appCleanupCalls["Cache"]++
				fmt.Printf("[App-Cleanup] Cache cleanup (call #%d)\n", appCleanupCalls["Cache"])
			case Logger:
				appCleanupCalls["Logger"]++
				fmt.Printf("[App-Cleanup] Logger cleanup (call #%d)\n", appCleanupCalls["Logger"])
			case Config:
				appCleanupCalls["Config"]++
				fmt.Printf("[App-Cleanup] Config cleanup (call #%d)\n", appCleanupCalls["Config"])
			case ServiceA:
				appCleanupCalls["ServiceA"]++
				fmt.Printf("[App-Cleanup] ServiceA cleanup (call #%d)\n", appCleanupCalls["ServiceA"])
			case ServiceB:
				appCleanupCalls["ServiceB"]++
				fmt.Printf("[App-Cleanup] ServiceB cleanup (call #%d)\n", appCleanupCalls["ServiceB"])
			}
		}
	})

	ctx := context.Background()

	fmt.Println("--- Injecting Dependencies ---")
	db, _ := godi.Inject[Database](app)
	fmt.Printf("Injected: Database{DSN: %s}\n\n", db.DSN)

	cache, _ := godi.Inject[Cache](app)
	fmt.Printf("Injected: Cache{Addr: %s}\n\n", cache.Addr)

	log, _ := godi.Inject[Logger](app)
	fmt.Printf("Injected: Logger{Level: %s}\n\n", log.Level)

	cfg, _ := godi.Inject[Config](app)
	fmt.Printf("Injected: Config{AppName: %s}\n\n", cfg.AppName)

	fmt.Println("--- Injecting Built Services ---")
	svcA, _ := godi.Inject[ServiceA](app)
	fmt.Printf("Injected: ServiceA{Name: %s, DB: %+v}\n\n", svcA.Name, svcA.DB)

	svcB, _ := godi.Inject[ServiceB](app)
	fmt.Printf("Injected: ServiceB{Name: %s, Cache: %+v, Log: %+v}\n\n", svcB.Name, svcB.Cache, svcB.Log)

	fmt.Println("--- Re-injecting (should not trigger HookOnce) ---")
	_, _ = godi.Inject[Database](app)
	_, _ = godi.Inject[Cache](app)
	fmt.Println()

	fmt.Println("=== Executing Init Hooks ===")
	infraInit.Iterate(ctx, false)
	servicesInit.Iterate(ctx, false)
	appInit.Iterate(ctx, false)

	fmt.Println("\n=== Executing Cleanup Hooks (HookOnce - only first injection) ===")
	infraCleanup.Iterate(ctx, true)
	servicesCleanup.Iterate(ctx, true)
	appCleanup.Iterate(ctx, true)

	fmt.Println("\n=== Verification Summary ===")
	fmt.Printf("Infra Init Calls: Database=%d, Cache=%d (expected: 1, 1)\n",
		infraInitCalls["Database"], infraInitCalls["Cache"])
	fmt.Printf("Services Init Calls: Database=%d, Cache=%d, Logger=%d (expected: 1, 1, 1)\n",
		servicesInitCalls["Database"], servicesInitCalls["Cache"], servicesInitCalls["Logger"])
	fmt.Printf("App Init Calls: Database=%d, Cache=%d, Logger=%d, Config=%d, ServiceA=%d, ServiceB=%d (expected: 1, 1, 1, 1, 1, 1)\n",
		appInitCalls["Database"], appInitCalls["Cache"], appInitCalls["Logger"],
		appInitCalls["Config"], appInitCalls["ServiceA"], appInitCalls["ServiceB"])

	fmt.Printf("\nInfra Cleanup Calls: Database=%d, Cache=%d (expected: 1, 1)\n",
		infraCleanupCalls["Database"], infraCleanupCalls["Cache"])
	fmt.Printf("Services Cleanup Calls: Database=%d, Cache=%d, Logger=%d (expected: 1, 1, 1)\n",
		servicesCleanupCalls["Database"], servicesCleanupCalls["Cache"], servicesCleanupCalls["Logger"])
	fmt.Printf("App Cleanup Calls: Database=%d, Cache=%d, Logger=%d, Config=%d, ServiceA=%d, ServiceB=%d (expected: 1, 1, 1, 1, 1, 1)\n",
		appCleanupCalls["Database"], appCleanupCalls["Cache"], appCleanupCalls["Logger"],
		appCleanupCalls["Config"], appCleanupCalls["ServiceA"], appCleanupCalls["ServiceB"])
}
