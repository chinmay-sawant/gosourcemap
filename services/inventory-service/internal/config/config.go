package config

import "os"

// Config holds application configuration.
type Config struct {
	Port string
}

// Load returns configuration from environment variables.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	return &Config{Port: port}
}
