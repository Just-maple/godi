package main

import (
	"context"
	"fmt"

	"github.com/Just-maple/godi"
)

type Config struct {
	AppName string
}

type Database struct {
	DSN string
}

type Service struct {
	Name string
}

func main() {
	c := &godi.Container{}

	c.MustAdd(
		godi.Provide(Config{AppName: "my-app"}),
		godi.Provide(Database{DSN: "mysql://localhost"}),
		godi.Provide(Service{Name: "user-service"}),
	)

	startup := c.Hook("startup", func(v any, provided int) func(context.Context) {
		if provided > 0 {
			return nil
		}
		return func(ctx context.Context) {
			fmt.Printf("Starting: %T\n", v)
		}
	})

	shutdown := c.HookOnce("shutdown", func(v any) func(context.Context) {
		return func(ctx context.Context) {
			fmt.Printf("Stopping: %T\n", v)
		}
	})

	ctx := context.Background()

	cfg := godi.MustInject[Config](c)
	db := godi.MustInject[Database](c)
	svc := godi.MustInject[Service](c)

	startup.Iterate(ctx, false)

	fmt.Printf("Running: %s, %s, %s\n", cfg.AppName, db.DSN, svc.Name)

	shutdown.Iterate(ctx, true)
}
