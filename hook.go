package godi

import (
	"context"
	"sync"
)

func (c *Container) Hook(name string, build func(v any, provided int) func(ctx context.Context)) func(func([]func(ctx context.Context))) {
	lock := sync.Mutex{}
	called := make(map[any]int)
	fns := make([]func(context.Context), 0)
	c.hooks.Store(name, func(id, v any) {
		lock.Lock()
		count := called[id]
		called[id]++
		lock.Unlock()
		if fn := build(v, count); fn != nil {
			lock.Lock()
			fns = append(fns, fn)
			lock.Unlock()
		}
	})
	return func(f func([]func(ctx context.Context))) {
		lock.Lock()
		hooks := make([]func(context.Context), 0, len(fns))
		for _, fn := range fns {
			hooks = append(hooks, fn)
		}
		lock.Unlock()
		f(hooks)
	}
}

func (c *Container) HookOnce(name string, build func(v any) func(ctx context.Context)) func(func([]func(ctx context.Context))) {
	return c.Hook(name, func(v any, provided int) func(ctx context.Context) {
		if provided > 0 {
			return nil
		}
		return build(v)
	})
}
