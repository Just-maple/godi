package godi

import (
	"context"
	"sync"
)

type Callbacks func(func([]func(ctx context.Context)))

func (callbacks Callbacks) Iterate(ctx context.Context, reverse bool) {
	callbacks(func(fns []func(ctx context.Context)) {
		for i := 0; i < len(fns); i++ {
			fn := fns[i]
			if reverse {
				fn = fns[len(fns)-1-i]
			}
			fn(ctx)
		}
	})
}

func (c *Container) Hook(name string, build func(v any, provided int) func(ctx context.Context)) Callbacks {
	c.once.Do(func() { c.hooks = new(sync.Map) })
	mu := sync.Mutex{}
	called := make(map[any]int)
	fns := make([]func(context.Context), 0)
	c.hooks.Store(name, func(id, v any) {
		mu.Lock()
		count := called[id]
		called[id]++
		mu.Unlock()
		if fn := build(v, count); fn != nil {
			mu.Lock()
			fns = append(fns, fn)
			mu.Unlock()
		}
	})
	return func(f func([]func(ctx context.Context))) {
		mu.Lock()
		cbs := make([]func(context.Context), len(fns))
		copy(cbs, fns)
		mu.Unlock()
		f(cbs)
	}
}

func (c *Container) HookOnce(name string, build func(v any) func(ctx context.Context)) Callbacks {
	return c.Hook(name, func(v any, provided int) func(ctx context.Context) {
		if provided > 0 {
			return nil
		}
		return build(v)
	})
}
