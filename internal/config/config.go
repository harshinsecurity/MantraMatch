package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// SuccessIndicator defines the criteria for a successful API key validation
type SuccessIndicator struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

// Validation defines the validation rules for an API service
type Validation struct {
	StatusCode       int              `yaml:"status_code"`
	ContentType      string           `yaml:"content_type,omitempty"`
	SuccessIndicator SuccessIndicator `yaml:"success_indicator"`
}

// Service represents an API service configuration
type Service struct {
	Name         string            `yaml:"name"`
	Regex        string            `yaml:"regex"`
	VerifyURL    string            `yaml:"verify_url"`
	VerifyMethod string            `yaml:"verify_method"`
	Headers      map[string]string `yaml:"headers"`
	Validation   Validation        `yaml:"validation"`
}

// Config holds the entire configuration for MantraMatch
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
		if err := validateService(service); err != nil {
			return fmt.Errorf("invalid service '%s': %w", service.Name, err)
		}
	}

	return nil
}

// validateService checks if a single service configuration is valid
func validateService(service Service) error {
	if service.Name == "" {
		return fmt.Errorf("service name cannot be empty")
	}
	if service.Regex == "" {
		return fmt.Errorf("regex cannot be empty")
	}
	if service.VerifyURL == "" {
		return fmt.Errorf("verify URL cannot be empty")
	}
	if service.VerifyMethod == "" {
		return fmt.Errorf("verify method cannot be empty")
	}
	if service.Validation.StatusCode == 0 {
		return fmt.Errorf("status code cannot be 0")
	}
	if err := validateSuccessIndicator(service.Validation.SuccessIndicator); err != nil {
		return fmt.Errorf("invalid success indicator: %w", err)
	}
	return nil
}

// validateSuccessIndicator checks if the success indicator is valid
func validateSuccessIndicator(indicator SuccessIndicator) error {
	validTypes := map[string]bool{
		"status_code_only": true,
		"json_key_exists":  true,
		"json_key_value":   true,
		"contains_string":  true,
		"regex_match":      true,
		"header_exists":    true,
		"header_value":     true,
	}

	if !validTypes[indicator.Type] {
		return fmt.Errorf("invalid success indicator type: %s", indicator.Type)
	}

	switch indicator.Type {
	case "json_key_exists", "json_key_value", "header_exists", "header_value":
		if indicator.Key == "" {
			return fmt.Errorf("key is required for type %s", indicator.Type)
		}
	case "json_key_value", "contains_string", "regex_match", "header_value":
		if indicator.Value == "" {
			return fmt.Errorf("value is required for type %s", indicator.Type)
		}
	}

	return nil
}
