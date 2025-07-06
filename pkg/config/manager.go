package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Manager struct {
	config *Config
	path   string
}

func NewManager() *Manager {
	return &Manager{
		config: NewConfig(),
	}
}

func (m *Manager) Load(configPath string) error {
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath = filepath.Join(home, DefaultConfigFile)
	}

	m.path = configPath

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil
	}

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(m.config); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func (m *Manager) Save() error {
	if m.path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		m.path = filepath.Join(home, DefaultConfigFile)
	}

	if err := os.MkdirAll(filepath.Dir(m.path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.Set("version", m.config.Version)
	viper.Set("workspace", m.config.Workspace)
	viper.Set("platforms", m.config.Platforms)
	viper.Set("defaults", m.config.Defaults)
	if m.config.RemoteSync != nil {
		viper.Set("remote_sync", m.config.RemoteSync)
	}

	if err := viper.WriteConfigAs(m.path); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (m *Manager) GetConfig() *Config {
	return m.config
}

func (m *Manager) SetConfig(config *Config) {
	m.config = config
}

func (m *Manager) GetConfigPath() string {
	return m.path
}

func (m *Manager) Reset() {
	m.config = NewConfig()
}

func (m *Manager) Validate() error {
	if m.config.Version == "" {
		return fmt.Errorf("config version is required")
	}

	if m.config.Workspace == "" {
		return fmt.Errorf("workspace is required")
	}

	for name, platform := range m.config.Platforms {
		if platform.Type == "" {
			return fmt.Errorf("platform type is required for platform %s", name)
		}
	}

	return nil
}