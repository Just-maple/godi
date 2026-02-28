package godi

import (
	"errors"
	"fmt"
	"sync"
)

var errTypeMismatch = errors.New("type mismatch")

type Provider interface {
	Inject(any) error
	ID() any
}

type prov[T any] func(any) error

func (p prov[T]) Inject(ptr any) error { return p(ptr) }
func (p prov[T]) ID() any              { return (*T)(nil) }

func Provide[T any](v T) Provider {
	return prov[T](func(dst any) error {
		if p, ok := dst.(*T); ok {
			*p = v
			return nil
		}
		return errTypeMismatch
	})
}

type Container struct {
	providers []Provider
	mu        sync.Mutex
	injecting sync.Map
}

func (c *Container) Add(p Provider) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, pv := range c.providers {
		if pv.ID() == p.ID() {
			return false
		}
	}
	c.providers = append(c.providers, p)
	return true
}

func (c *Container) ShouldAdd(p Provider) error {
	if c.Add(p) {
		return nil
	}
	return fmt.Errorf("provider %T already exists", p.ID())
}

func (c *Container) MustAdd(p Provider) {
	if e := c.ShouldAdd(p); e != nil {
		panic(e)
	}
}

func InjectTo[T any](v *T, cs ...*Container) error {
	id := (*T)(nil)
	for _, c := range cs {
		if _, on := c.injecting.Load(id); on {
			return fmt.Errorf("circular dependency for %T", v)
		}
		c.injecting.Store(id, true)
		c.mu.Lock()
		ps := c.providers
		c.mu.Unlock()
		for _, p := range ps {
			if err := p.Inject(v); err == nil {
				c.injecting.Delete(id)
				return nil
			} else if !errors.Is(err, errTypeMismatch) {
				c.injecting.Delete(id)
				return err
			}
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

func Inject[T any](c ...*Container) (T, error) {
	var v T
	return v, InjectTo(&v, c...)
}

func MustInject[T any](c ...*Container) T {
	v, e := Inject[T](c...)
	if e != nil {
		panic(e)
	}
	return v
}

func Lazy[T any](f func() (T, error)) Provider {
	l := &lazy[T]{f: f}
	return l
}

type lazy[T any] struct {
	f     func() (T, error)
	once  sync.Once
	value T
	err   error
}

func (l *lazy[T]) Inject(ptr any) error {
	p, ok := ptr.(*T)
	if !ok {
		return errTypeMismatch
	}
	l.once.Do(func() { l.value, l.err = l.f() })
	if l.err != nil {
		return fmt.Errorf("create %T error: %w", l.value, l.err)
	}
	*p = l.value
	return nil
}

func (l *lazy[T]) ID() any { return (*T)(nil) }
