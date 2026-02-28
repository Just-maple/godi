// Package infrastructure provides external service connections
// These are concrete implementations of interfaces defined in pkg/interfaces
package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/Just-maple/godi/examples/09-web-app/pkg/interfaces"
)

// DBConnection implements interfaces.Database
// This is a concrete implementation that can be swapped
type DBConnection struct {
	dsn       string
	connected time.Time
}

// NewDBConnection creates a new database connection
// Returns interfaces.Database (abstraction) for dependency inversion
func NewDBConnection(dsn string) interfaces.Database {
	fmt.Printf("[Infrastructure] Database connection established: %s\n", dsn)
	return &DBConnection{
		dsn:       dsn,
		connected: time.Now(),
	}
}

// Query implements interfaces.Database
func (c *DBConnection) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, nil
}

// Execute implements interfaces.Database
func (c *DBConnection) Execute(query string, args ...interface{}) (int64, error) {
	return 0, nil
}

// Close implements interfaces.Database
func (c *DBConnection) Close() error {
	fmt.Printf("[Infrastructure] Database connection closed: %s\n", c.dsn)
	return nil
}

// CacheClient implements interfaces.Cache
// This is a concrete implementation that can be swapped
type CacheClient struct {
	addr string
}

// NewCacheClient creates a new cache client
// Returns interfaces.Cache (abstraction) for dependency inversion
func NewCacheClient(addr string) interfaces.Cache {
	fmt.Printf("[Infrastructure] Cache client connected: %s\n", addr)
	return &CacheClient{addr: addr}
}

// Get implements interfaces.Cache
func (c *CacheClient) Get(key string) (interface{}, error) {
	return nil, nil
}

// Set implements interfaces.Cache
func (c *CacheClient) Set(key string, value interface{}, ttl int) error {
	return nil
}

// Delete implements interfaces.Cache
func (c *CacheClient) Delete(key string) error {
	return nil
}

// Initialize implements interfaces.Service
func (c *CacheClient) Initialize(ctx context.Context) error {
	return nil
}

// Close closes the cache connection
func (c *CacheClient) Close() error {
	fmt.Printf("[Infrastructure] Cache client disconnected: %s\n", c.addr)
	return nil
}
