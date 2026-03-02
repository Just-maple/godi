package godi

import (
	"context"
	"sync"
)

type HookFunc = func(ctx context.Context)

type hook struct {
	sync.Mutex
	build  func(any, int) HookFunc
	fns    []HookFunc
	called map[any]int
}

func (h *hook) check(id, v any) {
	h.Lock()
	defer h.Unlock()
	h.called[id]++
	if fn := h.build(v, h.called[id]-1); fn != nil {
		h.fns = append(h.fns, fn)
	}
}

func (h *hook) execute(exec func([]HookFunc)) {
	h.Lock()
	hooks := make([]HookFunc, 0, len(h.fns))
	for _, fn := range h.fns {
		hooks = append(hooks, fn)
	}
	h.Unlock()
	exec(hooks)
}

func (c *Container) Hook(name string, build func(v any, provided int) func(ctx context.Context)) func(func([]func(ctx context.Context))) {
	h := &hook{build: build, called: make(map[any]int)}
	c.hooks.Store(name, h)
	return h.execute
}

func (c *Container) HookOnce(name string, build func(v any, provided int) func(ctx context.Context)) func(func([]func(ctx context.Context))) {
	return c.Hook(name, func(v any, provided int) func(ctx context.Context) {
		if provided > 0 {
			return nil
		}
		return build(v, provided)
	})
}
