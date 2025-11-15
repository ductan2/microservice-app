package config

import (
	"sync"
)

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns the singleton configuration instance
func GetConfig() *Config {
	once.Do(func() {
		cfg, err := Load()
		if err != nil {
			panic("Failed to load configuration: " + err.Error())
		}

		if err := cfg.Validate(); err != nil {
			panic("Configuration validation failed: " + err.Error())
		}

		instance = cfg
	})
	return instance
}

// GetPort returns the server port from configuration
func GetPort() string {
	return GetConfig().Server.Port
}
