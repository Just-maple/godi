// Package lifecycle provides application lifecycle management
package lifecycle

import (
	"context"
	"fmt"
	"sync"
)

// Hook represents a shutdown hook function
type Hook func(context.Context) error

// Lifecycle manages the application lifecycle
// Shutdown hooks are executed in reverse order (LIFO)
type Lifecycle struct {
	hooks      []Hook
	hooksMutex sync.Mutex
	name       string
}

// New creates a new lifecycle manager
func New(name string) *Lifecycle {
	return &Lifecycle{
		hooks: make([]Hook, 0),
		name:  name,
	}
}

// AddShutdownHook adds a shutdown hook to be called on cleanup
// Hooks are executed in reverse order of registration
func (l *Lifecycle) AddShutdownHook(hook Hook) {
	l.hooksMutex.Lock()
	defer l.hooksMutex.Unlock()
	l.hooks = append(l.hooks, hook)
}

// Shutdown executes all shutdown hooks in reverse order
func (l *Lifecycle) Shutdown(ctx context.Context) error {
	l.hooksMutex.Lock()
	hooks := make([]Hook, len(l.hooks))
	copy(hooks, l.hooks)
	l.hooksMutex.Unlock()

	fmt.Printf("\n[%s] Starting shutdown (%d hooks)\n", l.name, len(hooks))

	// Execute hooks in reverse order (LIFO)
	for i := len(hooks) - 1; i >= 0; i-- {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] Shutdown cancelled\n", l.name)
			return ctx.Err()
		default:
			if err := hooks[i](ctx); err != nil {
				fmt.Printf("[%s] Hook %d error: %v\n", l.name, i, err)
			}
		}
	}

	fmt.Printf("[%s] Shutdown complete\n", l.name)
	return nil
}
