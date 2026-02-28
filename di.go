package godi

import (
	"fmt"
	"sync"
)

type Provider interface {
	Inject(any) bool
	ID() any
}

type prov[T any] func(any) bool

func (p prov[T]) Inject(ptr any) bool { return p(ptr) }
func (p prov[T]) ID() any             { return (*T)(nil) }

func Provide[T any](v T) Provider {
	return prov[T](func(dst any) bool {
		if p, ok := dst.(*T); ok {
			*p = v
			return true
		}
		return false
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
			if p.Inject(v) {
				c.injecting.Delete(id)
				return nil
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

func Inject[T any](c ...*Container) (v T, b bool) {
	return v, InjectTo(&v, c...) == nil
}

func ShouldInject[T any](c ...*Container) (v T, err error) {
	return v, InjectTo(&v, c...)
}

func MustInject[T any](c ...*Container) T {
	v, e := ShouldInject[T](c...)
	if e != nil {
		panic(e)
	}
	return v
}

func Lazy[T any](f func() (T, error)) Provider { return &lazy[T]{f: f} }

type lazy[T any] struct {
	f     func() (T, error)
	once  sync.Once
	value T
	err   error
}

func (l *lazy[T]) Inject(ptr any) bool {
	p, ok := ptr.(*T)
	if !ok {
		return false
	}
	if l.once.Do(func() { l.value, l.err = l.f() }); l.err == nil {
		*p = l.value
	}
	return l.err == nil
}

func (l *lazy[T]) ID() any { return (*T)(nil) }
