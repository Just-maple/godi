package godi

import (
	"fmt"
	"sync"
)

func Provide[T any](v T) Provider {
	return provider[T](func(dst any) bool {
		if prt, _ := dst.(*T); prt != nil {
			*prt = v
			return true
		}
		return false
	})
}

type provider[T any] func(ptr any) bool

func (p provider[T]) Inject(ptr any) bool { return p(ptr) }
func (p provider[T]) ID() any             { return (*T)(nil) }
func (p provider[T]) new() any            { return new(T) }

type Provider interface {
	Inject(pt any) (ok bool)
	ID() any
	new() any
}

type Container struct {
	providers []Provider
	mu        sync.Mutex
}

func (containers *Container) ShouldAdd(p Provider) (err error) {
	if containers.Add(p) {
		return nil
	}
	return fmt.Errorf("provider %T already exists", p.ID())
}

func (containers *Container) MustAdd(p Provider) {
	if e := containers.ShouldAdd(p); e != nil {
		panic(e)
	}
}

func (containers *Container) Add(p Provider) (success bool) {
	containers.mu.Lock()
	defer containers.mu.Unlock()
	id := p.ID()
	for _, pv := range containers.providers {
		if pv.ID() == id {
			return false
		}
	}
	containers.providers = append(containers.providers, p)
	return true
}

func withLock[T any](mutex *sync.Mutex, fn func() T) T {
	mutex.Lock()
	defer mutex.Unlock()
	return fn()
}

func injectTo[T any](v *T, c *Container) (ok bool) {
	for i := 0; withLock(&c.mu, func() bool { return i < len(c.providers) }); i++ {
		p := withLock(&c.mu, func() Provider { return c.providers[i] })
		lazy, is := p.(*lazyProvider)
		if is && lazy.instance == (*firing)(nil) {
			continue
		}
		if ok = p.Inject(v); ok || (is && lazy.err != nil) {
			return
		}
	}
	return
}

func InjectTo[T any](v *T, containers ...*Container) (ok bool) {
	for _, container := range containers {
		if injectTo[T](v, container) {
			return true
		}
	}
	return
}

func MustInjectTo[T any](v *T, containers ...*Container) {
	if e := ShouldInjectTo(v, containers...); e != nil {
		panic(e)
	}
	return
}

func ShouldInjectTo[T any](v *T, containers ...*Container) (err error) {
	if !InjectTo(v, containers...) {
		err = fmt.Errorf("provider %T not found", v)
	}
	return
}

func Inject[T any](containers ...*Container) (v T, ok bool) {
	ok = InjectTo(&v, containers...)
	return
}

func ShouldInject[T any](containers ...*Container) (v T, err error) {
	v, ok := Inject[T](containers...)
	if !ok {
		err = fmt.Errorf("provider %T not found", v)
	}
	return
}

func MustInject[T any](containers ...*Container) (v T) {
	v, e := ShouldInject[T](containers...)
	if e != nil {
		panic(e)
	}
	return
}

func Lazy[T any](factory func() (v T, err error)) Provider {
	l := &lazyProvider{}
	l.Provider = provider[T](func(ptr any) bool {
		l.once.Do(func() {
			l.instance = (*firing)(nil)
			l.instance, l.err = factory()
		})
		prt, _ := ptr.(*T)
		if l.err != nil || prt == nil {
			return false
		}
		*prt = l.instance.(T)
		return true
	})
	return l
}

type firing struct{}

type lazyProvider struct {
	Provider
	once     sync.Once
	instance interface{}
	err      error
}
