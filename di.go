package di

import (
    "fmt"
    "sync"
)

func Provide[T any](t T) Provider {
    return provider[T](func(ptr any) (ok bool) {
        v, ok := ptr.(*T)
        if ok {
            *v = t
        }
        return
    })
}

type provider[T any] func(ptr any) bool

func (p provider[T]) Inject(ptr any) bool { return p(ptr) }

func (p provider[T]) ID() any { return new(T) }

type Provider interface {
    Inject(pt any) (ok bool)
    ID() any
}

type Container struct {
    providers []Provider
    sync.Mutex
}

func (containers *Container) Add(p Provider) (err error) {
    containers.Lock()
    defer containers.Unlock()
    for _, exist := range containers.providers {
        if ok := exist.Inject(p.ID()); ok {
            err = fmt.Errorf("provider %T already exists", p.ID())
        }
    }
    containers.providers = append(containers.providers, p)
    return
}

func InjectTo[T any](containers *Container, v *T) (err error) {
    containers.Lock()
    defer containers.Unlock()
    for _, exist := range containers.providers {
        if ok := exist.Inject(v); ok {
            return nil
        }
    }
    return fmt.Errorf("provider %T doesn't exist", v)

}

func Inject[T any](containers *Container) (v T, err error) {
    err = InjectTo(containers, &v)
    return
}
