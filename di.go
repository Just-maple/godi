package godi

import (
	"fmt"
	"sync"
)

type Provider interface {
	inject(*Container, any) error
	ID() any
	Is(any) bool
}

type provider[T any] func(*Container, any) error

func (p provider[T]) Is(a any) bool                      { _, ok := a.(*T); return ok }
func (p provider[T]) ID() any                            { return (*T)(nil) }
func (p provider[T]) inject(c *Container, ptr any) error { return p(c, ptr) }

func Provide[T any](v T) Provider {
	return provider[T](func(c *Container, p any) error { return inject(c, p.(*T), v) })
}

func inject[T any](c *Container, p *T, v T) error {
	*p = v
	if c != nil {
		c.hooks.Range(func(_, h any) bool { h.(*hook).check((*T)(nil), v); return true })
	}
	return nil
}

func Build[T any](f func(*Container) (T, error)) Provider {
	l := new(struct {
		value T
		once  sync.Once
		err   error
	})
	return provider[T](func(c *Container, ptr any) error {
		if l.once.Do(func() { l.value, l.err = f(c) }); l.err != nil {
			return fmt.Errorf("create %T error: %w", l.value, l.err)
		}
		return inject(c, ptr.(*T), l.value)
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

func (c *Container) Add(ps ...Provider) error {
	for _, p := range ps {
		id := p.ID()
		if _, exist := c.providers.LoadOrStore(id, p); exist {
			return fmt.Errorf("provider %T already exists", id)
		}
	}
	return nil
}

func (c *Container) MustAdd(ps ...Provider) *Container {
	for _, p := range ps {
		must(c.Add(p))
	}
	return c
}

func (c *Container) inject(provider Provider, id, v any) (err error) {
	if _, on := c.injecting.LoadOrStore(id, true); on {
		return fmt.Errorf("circular dependency for %T", v)
	}
	defer c.injecting.Delete(id)
	return provider.inject(c, v)
}

func (c *Container) Inject(ps ...any) error {
	for _, p := range ps {
		if e := InjectAs(p, c); e != nil {
			return e
		}
	}
	return nil
}

func InjectAs(v any, cs ...*Container) (err error) {
	for _, c := range cs {
		if c.providers.Range(func(id, p interface{}) bool {
			if pv := p.(Provider); pv.Is(v) {
				err = c.inject(pv, id, v)
				v = nil
			}
			return v != nil
		}); v == nil {
			return
		}
	}
	return fmt.Errorf("provider %T not found", v)
}

func InjectTo[T any](v *T, cs ...*Container) (err error) {
	id := (*T)(nil)
	for _, c := range cs {
		if p, ok := c.providers.Load(id); ok {
			return c.inject(p.(Provider), id, v)
		}
	}
	return fmt.Errorf("provider %T not found", v)
}

func MustInjectAs(v any, c ...*Container)          { must(InjectAs(v, c...)) }
func MustInjectTo[T any](v *T, c ...*Container)    { must(InjectTo(v, c...)) }
func Inject[T any](c ...*Container) (v T, _ error) { return v, InjectTo(&v, c...) }
func MustInject[T any](c ...*Container) (v T)      { MustInjectTo[T](&v, c...); return }
