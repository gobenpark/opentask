package linear

import (
	"fmt"
	"opentask/pkg/platforms"
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
	return "linear"
}

func (f *Factory) GetName() string {
	return "Linear"
}

func (f *Factory) ValidateConfig(config map[string]any) error {
	_, err := parseConfig(config)
	return err
}

func parseConfig(config map[string]any) (Config, error) {
	cfg := Config{}

	// Extract token
	if token, ok := config["token"].(string); ok {
		cfg.Token = token
	} else {
		return cfg, fmt.Errorf("token is required and must be a string")
	}

	// Extract base URL (optional)
	if baseURL, ok := config["base_url"].(string); ok {
		cfg.BaseURL = baseURL
	}

	// Validate token is not empty
	if cfg.Token == "" {
		return cfg, fmt.Errorf("token cannot be empty")
	}

	return cfg, nil
}

// Register factory with the global registry
func init() {
	platforms.DefaultRegistry.Register(NewFactory())
}
