package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type SuccessIndicator struct {
	Type  string `yaml:"type"`
	Key   string `yaml:"key,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type Validation struct {
	StatusCode       int              `yaml:"status_code"`
	ContentType      string           `yaml:"content_type,omitempty"`
	SuccessIndicator SuccessIndicator `yaml:"success_indicator"`
}

type Service struct {
	Name         string            `yaml:"name"`
	Regex        string            `yaml:"regex"`
	VerifyURL    string            `yaml:"verify_url"`
	VerifyMethod string            `yaml:"verify_method"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	Validation   Validation        `yaml:"validation"`
	Note         string            `yaml:"note,omitempty"`
}

type Config struct {
	Services []Service `yaml:"services"`
}

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func CreateDefaultConfig(configPath string) error {
	defaultConfig := Config{
		Services: []Service{
			{
				Name:         "Example Service",
				Regex:        "^[a-zA-Z0-9]{32}$",
				VerifyURL:    "https://api.example.com/verify",
				VerifyMethod: "GET",
				Headers: map[string]string{
					"Authorization": "Bearer %s",
				},
				Validation: Validation{
					StatusCode: 200,
					SuccessIndicator: SuccessIndicator{
						Type: "json_key_exists",
						Key:  "success",
					},
				},
				Note: "This is an example service configuration.",
			},
		},
	}

	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("error marshaling default config: %w", err)
	}

	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing default config file: %w", err)
	}

	return nil
}

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

	if indicator.Type == "json_key_exists" || indicator.Type == "json_key_value" ||
		indicator.Type == "header_exists" || indicator.Type == "header_value" {
		if indicator.Key == "" {
			return fmt.Errorf("key is required for type %s", indicator.Type)
		}
	}

	if indicator.Type == "json_key_value" || indicator.Type == "contains_string" ||
		indicator.Type == "regex_match" || indicator.Type == "header_value" {
		if indicator.Value == "" {
			return fmt.Errorf("value is required for type %s", indicator.Type)
		}
	}

	return nil
}
