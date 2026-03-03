package godi

import (
	"errors"
	"fmt"
	"sync"
)

// must panics if the provided error is not nil.
// Used by Must* functions to convert errors to panics.
func must(err error) {
	if err != nil {
		panic(err)
	}
}

// typName returns the type name by extracting it from the formatted type string.
// For example, "*godi.Database" becomes "godi.Database".
func typName(ptr interface{}) string { return "[" + fmt.Sprintf("%T", ptr)[1:] + "]" }

// Provider is the interface that wraps the basic injection operations.
// All providers (Provide, Build) must implement this interface.
type Provider interface {
	// inject performs the actual dependency injection into the container.
	inject(c *Container, ptr any) (v any, err error)
	// Provide checks if a value matches the provider's type and returns type information.
	Provide(any) (any, bool)
}

// provider is the concrete implementation of Provider for generic types.
// It wraps a function that takes a container and pointer, returning the injected value.
type provider[T any] func(*Container, *T) (T, error)

// Provide returns type information for the provider.
// It checks if the given value matches the provider's type T.
func (p provider[T]) Provide(v any) (any, bool) { _, ok := v.(*T); return (*T)(nil), ok }

// inject executes the provider function to inject the value into the pointer.
func (p provider[T]) inject(c *Container, ptr any) (any, error) { return p(c, ptr.(*T)) }

// Provide creates a Provider that returns a pre-existing value.
// This is used for simple values that don't require construction logic.
// Example: Provide(Config{DSN: "mysql://localhost"})
func Provide[T any](v T) Provider {
	return provider[T](func(_ *Container, ptr *T) (T, error) { *ptr = v; return v, nil })
}

// Build creates a Provider that constructs a value lazily (on first use).
// The factory function f can depend on:
//   - No dependencies: func(struct{}) (T, error)
//   - Container access: func(*Container) (T, error)
//   - Single dependency: func(Dependency) (T, error)
//
// The built value is cached (singleton pattern) after first construction.
// Example: Build(func(c *Container) (*Database, error) { return NewDB() })
func Build[R, T any](f func(R) (T, error)) Provider {
	// l stores the lazy-initialized value with sync.Once for thread-safety
	l := new(struct {
		value T
		once  sync.Once
		err   error
	})
	return provider[T](func(c *Container, ptr *T) (zero T, err error) {
		// Recover from panics and convert to errors
		defer func() {
			if e := recover(); e != nil {
				err = fmt.Errorf("recovered from build %s panic: %v", typName(ptr), e)
			}
		}()
		var v R
		// Handle different dependency patterns
		switch pr := any(&v).(type) {
		case *struct{}:
			// No dependencies - empty struct
		case **Container:
			// Container access - inject container itself
			*pr = c
		default:
			// Single dependency - inject from container
			if e := InjectTo[R](c, &v); e != nil {
				return zero, e
			}
		}
		// Execute factory function once (singleton)
		if l.once.Do(func() { l.value, l.err = f(v) }); l.err != nil {
			return zero, fmt.Errorf("build %s error: %w", typName(ptr), l.err)
		}
		*ptr = l.value
		return l.value, nil
	})
}

// Container is the core dependency injection container.
// It manages providers, handles injection, and supports lifecycle hooks.
// Thread-safe for concurrent access.
type Container struct {
	once      sync.Once // Reserved for future initialization logic
	hooks     *sync.Map // Stores lifecycle hooks (Hook, HookOnce)
	providers sync.Map  // Stores all registered providers
}

// locked is a sentinel value used to mark frozen containers.
// When a container is added as a child, it becomes frozen and cannot accept new providers.
var locked = &Container{}

// Add registers one or more providers to the container.
// Returns an error if:
//   - The container is frozen (already added as a child to another container)
//   - A provider for the same type already exists
//
// When adding a child container, it becomes frozen to prevent modification.
func (c *Container) Add(pvs ...Provider) error {
	// Acquire lock using atomic compare-and-swap pattern
	// This prevents concurrent modifications and detects frozen state
	for acquired, val := new(Container), any(nil); val != acquired; {
		if val, _ = c.providers.LoadOrStore(locked, acquired); val == locked {
			return errors.New("container frozen: already provided as child")
		}
	}
	defer c.providers.Delete(locked)

	// Register each provider
	for _, p := range pvs {
		id, _ := p.Provide(nil)
		// Check for duplicate types
		if typ, provided := c.Provide(id); provided {
			return fmt.Errorf("provider %s already exists", typName(typ))
		} else if sub, ok := id.(*Container); ok {
			// Mark child container as frozen
			sub.providers.Store(locked, locked)
		}
		c.providers.Store(id, p)
	}
	return nil
}

// MustAdd is like Add but panics on error instead of returning it.
// Useful for initialization code where errors are unexpected.
func (c *Container) MustAdd(ps ...Provider) *Container { must(c.Add(ps...)); return c }

// Provide checks if a type is provided by this container (including nested containers).
// Returns the type information and a boolean indicating if it was found.
// Supports checking both direct providers and child containers.
func (c *Container) Provide(v any) (id any, ok bool) {
	if v == nil {
		return c, false
	} else if sub, is := v.(*Container); is {
		// Recursively check child containers
		sub.providers.Range(func(k, _ any) bool { id, ok = c.Provide(k); return !ok })
	} else {
		// Check direct providers
		c.providers.Range(func(_, p any) bool { id, ok = p.(Provider).Provide(v); return !ok })
	}
	return id, ok
}

// inject is the internal injection logic that searches for providers.
// It traverses the container hierarchy to find the appropriate provider.
func (c *Container) inject(parent *Container, ptr any) (value any, err error) {
	ok := false
	// Search through all providers in this container
	if c.providers.Range(func(_, p any) bool {
		pv := p.(Provider)
		if value, ok = pv.Provide(ptr); ok {
			// Found provider - execute injection with circular dependency tracking
			value, err = c.from(pv, value, ptr, parent)
		}
		return !ok
	}); !ok {
		// No provider found
		err = fmt.Errorf("provider %s not found", typName(ptr))
	}
	return
}

// from executes the provider injection while tracking dependencies for circular detection.
// It creates a temporary container context to track the dependency chain.
func (c *Container) from(p Provider, id, ptr any, parent *Container) (v any, err error) {
	// Check if this type is already being injected (circular dependency detection)
	if stat, _ := c.providers.Load(id); stat == locked {
		return nil, fmt.Errorf("circular dependency for %s ", typName(id))
	}

	// Create temporary container context for this injection
	cp := &Container{hooks: c.hooks}
	copyMap := func(k, v interface{}) bool { cp.providers.Store(k, v); return true }

	// Copy providers from current and parent containers
	if c.providers.Range(copyMap); parent != nil {
		parent.providers.Range(copyMap)
	}

	// Mark current type as being injected
	cp.providers.Store(id, locked)
	cp.providers.Delete(locked)

	// Execute the actual injection
	if v, err = p.inject(cp, ptr); err == nil && cp.hooks != nil {
		// Trigger hooks after successful injection
		cp.hooks.Range(func(_, h any) bool { h.(func(any, any))(id, v); return true })
	}
	return
}

// Inject injects dependencies into multiple pointers.
// Returns the first error encountered, or nil if all injections succeed.
func (c *Container) Inject(ptrs ...any) error {
	for _, ptr := range ptrs {
		if _, e := c.inject(c, ptr); e != nil {
			return e
		}
	}
	return nil
}

// InjectTo injects a dependency into a specific pointer.
// First checks if the type exists directly, otherwise searches the container hierarchy.
func InjectTo[T any](c *Container, ptr *T) (err error) {
	// Try direct provider first
	if p, ok := c.providers.Load((*T)(nil)); ok {
		_, err = c.from(p.(Provider), (*T)(nil), ptr, nil)
	} else {
		// Search container hierarchy
		_, err = c.inject(c, ptr)
	}
	return
}

// MustInjectTo is like InjectTo but panics on error.
func MustInjectTo[T any](c *Container, v *T) { must(InjectTo[T](c, v)) }

// Inject retrieves a dependency of type T from the container.
// Returns the value and an error if not found.
func Inject[T any](c *Container) (v T, _ error) { return v, InjectTo[T](c, &v) }

// MustInject is like Inject but panics on error and returns only the value.
func MustInject[T any](c *Container) (v T) { MustInjectTo[T](c, &v); return }

// InjectAs injects dependencies using non-generic interface.
// Useful when generic syntax is not available or for dynamic types.
func InjectAs(c *Container, ptrs ...any) (err error) { return c.Inject(ptrs...) }

// MustInjectAs is like InjectAs but panics on error.
func MustInjectAs(c *Container, ptrs ...any) { must(InjectAs(c, ptrs...)) }
