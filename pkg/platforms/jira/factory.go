package jira

import (
	"fmt"
	"opentask/pkg/platforms"
	"strings"
)

type Factory struct{}

func NewFactory() *Factory {
	return &Factory{}
}

func (f *Factory) Create(config map[string]any) (platforms.PlatformClient, error) {
	cfg, err := parseConfig(config)
	if err != nil {
		return nil, err
	}

	return NewClient(cfg)
}

func (f *Factory) GetType() string {
	return "jira"
}

func (f *Factory) GetName() string {
	return "Jira"
}

func (f *Factory) ValidateConfig(config map[string]any) error {
	_, err := parseConfig(config)
	return err
}

func parseConfig(config map[string]any) (Config, error) {
	cfg := Config{}

	// Extract base URL
	if baseURL, ok := config["base_url"].(string); ok {
		cfg.BaseURL = strings.TrimSuffix(baseURL, "/")
	} else {
		return cfg, fmt.Errorf("base_url is required and must be a string")
	}

	// Extract email
	if email, ok := config["email"].(string); ok {
		cfg.Email = email
	} else {
		return cfg, fmt.Errorf("email is required and must be a string")
	}

	// Extract token
	if token, ok := config["token"].(string); ok {
		cfg.Token = token
	} else {
		return cfg, fmt.Errorf("token is required and must be a string")
	}

	// Validate required fields
	if cfg.BaseURL == "" {
		return cfg, fmt.Errorf("base_url cannot be empty")
	}

	if cfg.Email == "" {
		return cfg, fmt.Errorf("email cannot be empty")
	}

	if cfg.Token == "" {
		return cfg, fmt.Errorf("token cannot be empty")
	}

	// Validate base URL format
	if !strings.HasPrefix(cfg.BaseURL, "http://") && !strings.HasPrefix(cfg.BaseURL, "https://") {
		return cfg, fmt.Errorf("base_url must start with http:// or https://")
	}

	return cfg, nil
}

// Register factory with the global registry
func init() {
	platforms.DefaultRegistry.Register(NewFactory())
}
