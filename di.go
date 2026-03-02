package godi

import (
	"fmt"
	"sync"
)

type Provider interface {
	inject(c *Container, ptr any) (v any, err error)
	Provide(any) (any, bool)
}

type provider[T any] func(*Container, *T) (T, error)

func (p provider[T]) Provide(a any) (any, bool)                 { _, ok := a.(*T); return (*T)(nil), ok }
func (p provider[T]) inject(c *Container, ptr any) (any, error) { return p(c, ptr.(*T)) }

func Provide[T any](v T) Provider {
	return provider[T](func(c *Container, p *T) (T, error) { *p = v; return v, nil })
}

func Build[T any](f func(*Container) (T, error)) Provider {
	l := new(struct {
		value T
		once  sync.Once
		err   error
	})
	return provider[T](func(c *Container, ptr *T) (zero T, _ error) {
		if l.once.Do(func() { l.value, l.err = f(c) }); l.err != nil {
			return zero, fmt.Errorf("build %T error: %w", l.value, l.err)
		}
		*ptr = l.value
		return l.value, nil
	})
}

func Chain[R any, T any](f func(r R) (T, error)) Provider {
	return Build(func(c *Container) (zero T, err error) {
		r, err := Inject[R](c)
		if err != nil {
			return zero, err
		}
		return f(r)
	})
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

type Container struct{ providers, injecting, hooks sync.Map }

func (c *Container) Provide(v any) (id any, ok bool) {
	if v == nil {
		return c, false
	} else if sub, is := v.(*Container); is {
		sub.providers.Range(func(k, _ any) bool { id, ok = c.Provide(k); return !ok })
	} else {
		c.providers.Range(func(_, p any) bool { id, ok = p.(Provider).Provide(v); return !ok })
	}
	return id, ok
}

func (c *Container) inject(_ *Container, v any) (value any, err error) {
	if c.providers.Range(func(_, p any) bool {
		pv := p.(Provider)
		if _id, ok := pv.Provide(v); ok {
			value, err = c.injectFrom(pv, _id, v)
		}
		return value == nil && err == nil
	}); value == nil && err == nil {
		err = fmt.Errorf("provider %T not found", v)
	}
	return
}

func (c *Container) Add(ps ...Provider) error {
	if parent, exist := c.injecting.Load(c); exist {
		return fmt.Errorf("container frozen: already provided to %T %p", parent, parent)
	}
	for _, p := range ps {
		id, _ := p.Provide(nil)
		if _id, provided := c.Provide(id); provided {
			return fmt.Errorf("provider %T already exists", _id)
		} else if sub, ok := id.(*Container); ok {
			sub.injecting.Store(sub, c)
		}
		c.providers.Store(id, p)
	}
	return nil
}

func (c *Container) MustAdd(ps ...Provider) *Container {
	for _, p := range ps {
		must(c.Add(p))
	}
	return c
}

func (c *Container) injectFrom(p Provider, id, ptr any) (v any, err error) {
	if _, on := c.injecting.LoadOrStore(id, ptr); on {
		return nil, fmt.Errorf("circular dependency for %T", ptr)
	} else if v, err = p.inject(c, ptr); err == nil {
		c.hooks.Range(func(_, h any) bool { h.(func(any, any))(id, v); return true })
	}
	c.injecting.Delete(id)
	return
}

func (c *Container) Inject(ps ...any) error {
	for _, p := range ps {
		if e := InjectAs(p, c); e != nil {
			return e
		}
	}
	return nil
}

func InjectTo[T any](v *T, c *Container) (err error) {
	id := (*T)(nil)
	if p, ok := c.providers.Load(id); ok {
		_, err = c.injectFrom(p.(Provider), id, v)
		return
	}
	_, err = c.inject(c, v)
	return
}

func InjectAs(v any, c *Container) (err error)  { _, err = c.inject(c, v); return }
func Inject[T any](c *Container) (v T, _ error) { return v, InjectTo(&v, c) }
func MustInjectAs(v any, c *Container)          { must(InjectAs(v, c)) }
func MustInjectTo[T any](v *T, c *Container)    { must(InjectTo(v, c)) }
func MustInject[T any](c *Container) (v T)      { MustInjectTo[T](&v, c); return }
