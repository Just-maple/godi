package godi

import (
	"errors"
	"fmt"
	"sync"
)

var errTypeMismatch = errors.New("type mismatch")

type Provider interface {
	Inject(*Container, any) error
	ID() any
}

type prov[T any] func(*Container, any) error

func (p prov[T]) Inject(c *Container, ptr any) error { return p(c, ptr) }
func (p prov[T]) ID() any                            { return (*T)(nil) }

func Provide[T any](v T) Provider {
	return prov[T](func(_ *Container, dst any) error {
		if p, ok := dst.(*T); ok {
			*p = v
			return nil
		}
		return errTypeMismatch
	})
}

type Container struct {
	providers sync.Map
	injecting sync.Map
}

func (c *Container) Add(ps ...Provider) error {
	for _, p := range ps {
		id := p.ID()
		if _, exists := c.providers.Load(id); exists {
			return fmt.Errorf("provider %T already exists", id)
		}
		c.providers.Store(id, p)
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
		if _, found := c.injecting.Load(id); !found {
			c.injecting.Store(id, true)
			c.providers.Range(func(key, value any) bool {
				err = value.(Provider).Inject(c, v)
				found = err == nil || !errors.Is(err, errTypeMismatch)
				return !found
			})
			if c.injecting.Delete(id); !found {
				continue
			}
			return err
		}
		return fmt.Errorf("circular dependency for %T", v)
	}
	return fmt.Errorf("provider %T not found", v)
}

func MustInjectTo[T any](v *T, c ...*Container) {
	if e := InjectTo(v, c...); e != nil {
		panic(e)
	}
}

func Inject[T any](c ...*Container) (v T, _ error) {
	return v, InjectTo(&v, c...)
}

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
		if p, ok := ptr.(*T); ok {
			if l.once.Do(func() { l.value, l.err = f(c) }); l.err != nil {
				return fmt.Errorf("create %T error: %w", l.value, l.err)
			}
			*p = l.value
			return nil
		}
		return errTypeMismatch
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
