package provider

import (
	"cinemaos-backend/internal/config"
)

// ProvideConfig loads and returns the application configuration
func ProvideConfig(configPath string) (*config.Config, error) {
	return config.Load(configPath)
}
