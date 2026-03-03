package godi

import (
	"errors"
	"fmt"
	"sync"
)

type Provider interface {
	inject(c *Container, ptr any) (v any, err error)
	Provide(any) (any, bool)
}

type provider[T any] func(*Container, *T) (T, error)

func (p provider[T]) Provide(v any) (any, bool)                 { _, ok := v.(*T); return (*T)(nil), ok }
func (p provider[T]) inject(c *Container, ptr any) (any, error) { return p(c, ptr.(*T)) }

func Provide[T any](v T) Provider {
	return provider[T](func(c *Container, ptr *T) (T, error) { *ptr = v; return v, nil })
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

type Container struct {
	once      sync.Once
	hooks     *sync.Map
	providers sync.Map
}

var locked = &Container{}

func (c *Container) Add(pvs ...Provider) error {
	for acquired, val := new(Container), any(nil); val != acquired; {
		if val, _ = c.providers.LoadOrStore(locked, acquired); val == locked {
			return errors.New("container frozen: already provided as child")
		}
	}
	defer c.providers.Delete(locked)
	for _, p := range pvs {
		id, _ := p.Provide(nil)
		if typ, provided := c.Provide(id); provided {
			return fmt.Errorf("provider %T already exists", typ)
		} else if sub, ok := id.(*Container); ok {
			sub.providers.Store(locked, locked)
		}
		c.providers.Store(id, p)
	}
	return nil
}

func (c *Container) MustAdd(ps ...Provider) *Container { must(c.Add(ps...)); return c }

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

func (c *Container) inject(_ *Container, ptr any) (value any, err error) {
	if c.providers.Range(func(_, p any) bool {
		pv := p.(Provider)
		if id, ok := pv.Provide(ptr); ok {
			value, err = c.from(pv, id, ptr)
		}
		return value == nil && err == nil
	}); value == nil && err == nil {
		err = fmt.Errorf("provider %T not found", ptr)
	}
	return
}

func (c *Container) from(p Provider, id, ptr any) (v any, err error) {
	if stat, _ := c.providers.Load(id); stat == locked {
		return nil, fmt.Errorf("circular dependency for %T", ptr)
	}
	cp := &Container{hooks: c.hooks}
	c.providers.Range(func(k, v interface{}) bool {
		if k == id {
			v = locked
		}
		cp.providers.Store(k, v)
		return true
	})
	if v, err = p.inject(cp, ptr); err == nil && cp.hooks != nil {
		cp.hooks.Range(func(_, h any) bool { h.(func(any, any))(id, v); return true })
	}
	return
}

func (c *Container) Inject(ptrs ...any) error {
	for _, ptr := range ptrs {
		if _, e := c.inject(c, ptr); e != nil {
			return e
		}
	}
	return nil
}

func InjectTo[T any](c *Container, ptr *T) (err error) {
	id := (*T)(nil)
	if p, ok := c.providers.Load(id); ok {
		_, err = c.from(p.(Provider), id, ptr)
	} else {
		_, err = c.inject(c, ptr)
	}
	return
}

func MustInjectTo[T any](c *Container, v *T)         { must(InjectTo[T](c, v)) }
func Inject[T any](c *Container) (v T, _ error)      { return v, InjectTo[T](c, &v) }
func MustInject[T any](c *Container) (v T)           { MustInjectTo[T](c, &v); return }
func InjectAs(c *Container, ptrs ...any) (err error) { return c.Inject(ptrs...) }
func MustInjectAs(c *Container, ptrs ...any)         { must(InjectAs(c, ptrs...)) }
