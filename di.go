package godi

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func typName(ptr interface{}) string { return fmt.Sprintf("%T", ptr)[1:] }

type Provider interface {
	inject(c *Container, ptr any) (v any, err error)
	Provide(any) (any, bool)
}

type provider[T any] func(*Container, *T) (T, error)

func (p provider[T]) Provide(v any) (any, bool)                 { _, ok := v.(*T); return (*T)(nil), ok }
func (p provider[T]) inject(c *Container, ptr any) (any, error) { return p(c, ptr.(*T)) }

func Provide[T any](v T) Provider {
	return provider[T](func(_ *Container, ptr *T) (T, error) { *ptr = v; return v, nil })
}

func Build[R, T any](f func(R) (T, error)) Provider {
	l := new(struct {
		value T
		once  sync.Once
		err   error
	})
	return provider[T](func(c *Container, ptr *T) (zero T, err error) {
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("panic: %v", e)
			}
		}()
		var v R
		switch pr := any(&v).(type) {
		case *struct{}:
		case **Container:
			*pr = c
		default:
			if e := InjectTo[R](c, &v); e != nil {
				return zero, e
			}
		}
		if l.once.Do(func() { l.value, l.err = f(v) }); l.err != nil {
			return zero, fmt.Errorf("build %s error: %w", typName(ptr), l.err)
		}
		*ptr = l.value
		return l.value, nil
	})
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
			return fmt.Errorf("provider %s already exists", typName(typ))
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

func (c *Container) inject(parent *Container, ptr any) (value any, err error) {
	ok := false
	if c.providers.Range(func(_, p any) bool {
		pv := p.(Provider)
		if value, ok = pv.Provide(ptr); ok {
			value, err = c.from(pv, value, ptr, parent)
		}
		return !ok
	}); !ok {
		err = fmt.Errorf("provider %s not found", typName(ptr))
	}
	return
}

func (c *Container) from(p Provider, id, ptr any, parent *Container) (v any, err error) {
	type depends struct {
		Container
		depth int
	}
	stat, _ := c.providers.Load(id)
	if _, ok := stat.(*depends); ok {
		deps, depth := make([]string, 0), make(map[string]int)
		c.providers.Range(func(k, v any) bool {
			if vv, is := v.(*depends); is {
				name := typName(k)
				deps, depth[name] = append(deps, name), vv.depth
			}
			return true
		})
		sort.Slice(deps, func(i, j int) bool { return depth[deps[i]] > depth[deps[j]] })
		return nil, fmt.Errorf("circular dependency for %s <-> %s", typName(id), strings.Join(deps, " <- "))
	}
	cp, dep := &Container{hooks: c.hooks}, &depends{}
	copyMap := func(k, v interface{}) bool {
		if _, ok := v.(*depends); ok {
			dep.depth += 1
		}
		cp.providers.Store(k, v)
		return true
	}
	if c.providers.Range(copyMap); parent != nil {
		parent.providers.Range(copyMap)
	}
	cp.providers.Store(id, dep)
	cp.providers.Delete(locked)
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
	if p, ok := c.providers.Load((*T)(nil)); ok {
		_, err = c.from(p.(Provider), (*T)(nil), ptr, nil)
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
