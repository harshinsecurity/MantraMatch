// internal/config.go
package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Service represents an API service with its verification details
type Service struct {
	Name         string            `yaml:"name"`
	Regex        string            `yaml:"regex"`
	VerifyURL    string            `yaml:"verify_url"`
	VerifyMethod string            `yaml:"verify_method"`
	Headers      map[string]string `yaml:"headers"`
	SuccessKey   string            `yaml:"success_key"`
}

// Config holds the configuration for the tool
type Config struct {
	Services []Service `yaml:"services"`
}

// LoadConfig reads and parses the configuration file
func LoadConfig(configPath string) (*Config, error) {
	// If no config path is provided, use the default
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("error getting user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".config", "mantramatch", "config.yaml")
	}

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Parse the YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	// Validate the config
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig checks if the loaded configuration is valid
func validateConfig(config *Config) error {
	if len(config.Services) == 0 {
		return fmt.Errorf("no services defined in the configuration")
	}

	for _, service := range config.Services {
		if service.Name == "" {
			return fmt.Errorf("service name cannot be empty")
		}
		if service.Regex == "" {
			return fmt.Errorf("regex for service '%s' cannot be empty", service.Name)
		}
		if service.VerifyURL == "" {
			return fmt.Errorf("verify URL for service '%s' cannot be empty", service.Name)
		}
		if service.VerifyMethod == "" {
			return fmt.Errorf("verify method for service '%s' cannot be empty", service.Name)
		}
	}

	return nil
}
