package godi

import (
	"fmt"
	"sync"
)

type Provider interface {
	inject(*Container, any) error
	ID() any
}

type prov[T any] func(*Container, any) error

func (p prov[T]) inject(c *Container, ptr any) error { return p(c, ptr) }
func (p prov[T]) ID() any                            { return (*T)(nil) }

func Provide[T any](v T) Provider {
	return prov[T](func(_ *Container, dst any) error { *dst.(*T) = v; return nil })
}

type Container struct{ providers, injecting sync.Map }

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
		if err := c.Add(p); err != nil {
			panic(err)
		}
	}
	return c
}

func InjectTo[T any](v *T, cs ...*Container) (err error) {
	id := (*T)(nil)
	for _, c := range cs {
		if _, on := c.injecting.LoadOrStore(id, true); on {
			return fmt.Errorf("circular dependency for %T", v)
		}
		if p, ok := c.providers.Load(id); ok {
			err = p.(Provider).inject(c, v)
			c.injecting.Delete(id)
			return
		}
		c.injecting.Delete(id)
	}
	return fmt.Errorf("provider %T not found", v)
}

func MustInjectTo[T any](v *T, c ...*Container) {
	if e := InjectTo(v, c...); e != nil {
		panic(e)
	}
}

func Inject[T any](c ...*Container) (v T, _ error) { return v, InjectTo(&v, c...) }

func MustInject[T any](c ...*Container) T {
	v, e := Inject[T](c...)
	if e != nil {
		panic(e)
	}
	return v
}

func Lazy[T any](f func(*Container) (T, error)) Provider {
	l := new(struct {
		once  sync.Once
		value T
		err   error
	})
	return prov[T](func(c *Container, ptr any) error {
		if l.once.Do(func() { l.value, l.err = f(c) }); l.err != nil {
			return fmt.Errorf("create %T error: %w", l.value, l.err)
		}
		*ptr.(*T) = l.value
		return nil
	})
}

func Chain[R any, T any](f func(r R) (T, error)) Provider {
	return Lazy(func(c *Container) (zero T, _ error) {
		r, err := Inject[R](c)
		if err != nil {
			return
		}
		return f(r)
	})
}
