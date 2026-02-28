// Package config provides application configuration
package config

// Config holds all application configuration
type Config struct {
	// AppName is the application name
	AppName string

	// DatabaseDSN is the database connection string
	DatabaseDSN string

	// CacheAddr is the Redis cache address
	CacheAddr string

	// Port is the HTTP server port
	Port int

	// Debug enables debug mode
	Debug bool
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		AppName:     "WebApp",
		DatabaseDSN: "postgres://localhost:5432/mydb",
		CacheAddr:   "redis://localhost:6379",
		Port:        8080,
		Debug:       true,
	}
}
