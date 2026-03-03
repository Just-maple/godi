// Package middleware provides request processing pipeline
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// LoggingMiddleware logs request processing
type LoggingMiddleware struct {
	debug bool
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(debug bool) *LoggingMiddleware {
	return &LoggingMiddleware{debug: debug}
}

// Process implements interfaces.Middleware
func (m *LoggingMiddleware) Process(handler interfaces.Handler) interfaces.Handler {
	return &loggingHandler{
		handler: handler,
		debug:   m.debug,
	}
}

type loggingHandler struct {
	handler interfaces.Handler
	debug   bool
}

func (h *loggingHandler) Handle(ctx context.Context) error {
	start := time.Now()
	if h.debug {
		fmt.Println("[DEBUG] Request started")
	}

	err := h.handler.Handle(ctx)

	duration := time.Since(start)
	if h.debug {
		fmt.Printf("[DEBUG] Request completed in %v\n", duration)
	}

	return err
}
