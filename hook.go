package godi

import (
	"context"
	"sync"
)

type Callbacks func(func([]func(ctx context.Context)))

func (callbacks Callbacks) Iterate(ctx context.Context, reverse bool) {
	callbacks(func(fns []func(ctx context.Context)) {
		for v, i, l := 0, 0, len(fns); i < l; i++ {
			if v = i; reverse {
				v = l - i - 1
			}
			fns[v](ctx)
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
		defer mu.Unlock()
		if fn := build(v, called[id]); fn != nil {
			fns = append(fns, fn)
		}
		called[id]++
	})
	return func(f func([]func(ctx context.Context))) {
		mu.Lock()
		cbs := append(make([]func(context.Context), 0, len(fns)), fns...)
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
