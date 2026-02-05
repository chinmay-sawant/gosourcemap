package config

import "os"

// Config holds application configuration.
type Config struct {
	Port                string
	OrderServiceURL     string
	InventoryServiceURL string
}

// Load returns configuration from environment variables.
func Load() *Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	orderURL := os.Getenv("ORDER_SERVICE_URL")
	if orderURL == "" {
		orderURL = "http://localhost:8081"
	}

	inventoryURL := os.Getenv("INVENTORY_SERVICE_URL")
	if inventoryURL == "" {
		inventoryURL = "http://localhost:8082"
	}

	return &Config{
		Port:                port,
		OrderServiceURL:     orderURL,
		InventoryServiceURL: inventoryURL,
	}
}
