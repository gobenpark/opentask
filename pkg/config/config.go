package config

import (
	"time"
)

type Config struct {
	Version    string                 `yaml:"version" json:"version"`
	Workspace  string                 `yaml:"workspace" json:"workspace"`
	Platforms  map[string]Platform    `yaml:"platforms" json:"platforms"`
	Defaults   Defaults               `yaml:"defaults" json:"defaults"`
	RemoteSync *RemoteSync            `yaml:"remote_sync,omitempty" json:"remote_sync,omitempty"`
}

type Platform struct {
	Type        string            `yaml:"type" json:"type"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Credentials map[string]string `yaml:"credentials" json:"credentials"`
	Settings    map[string]any    `yaml:"settings" json:"settings"`
}

type Defaults struct {
	Platform string `yaml:"platform" json:"platform"`
	Assignee string `yaml:"assignee,omitempty" json:"assignee,omitempty"`
	Priority string `yaml:"priority,omitempty" json:"priority,omitempty"`
	Project  string `yaml:"project,omitempty" json:"project,omitempty"`
}

type RemoteSync struct {
	Type     string `yaml:"type" json:"type"`
	URL      string `yaml:"url" json:"url"`
	Branch   string `yaml:"branch,omitempty" json:"branch,omitempty"`
	Enabled  bool   `yaml:"enabled" json:"enabled"`
	Interval string `yaml:"interval,omitempty" json:"interval,omitempty"`
}

type Workspace struct {
	Name        string    `yaml:"name" json:"name"`
	Description string    `yaml:"description,omitempty" json:"description,omitempty"`
	CreatedAt   time.Time `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time `yaml:"updated_at" json:"updated_at"`
	Config      *Config   `yaml:"config,omitempty" json:"config,omitempty"`
}

const (
	DefaultConfigVersion = "1.0"
	DefaultWorkspace     = "default"
	DefaultConfigFile    = ".opentask.yaml"
)

func NewConfig() *Config {
	return &Config{
		Version:   DefaultConfigVersion,
		Workspace: DefaultWorkspace,
		Platforms: make(map[string]Platform),
		Defaults: Defaults{
			Platform: "",
			Priority: "medium",
		},
	}
}

func (c *Config) GetPlatform(name string) (Platform, bool) {
	platform, exists := c.Platforms[name]
	return platform, exists
}

func (c *Config) AddPlatform(name string, platform Platform) {
	if c.Platforms == nil {
		c.Platforms = make(map[string]Platform)
	}
	c.Platforms[name] = platform
}

func (c *Config) RemovePlatform(name string) {
	delete(c.Platforms, name)
}

func (c *Config) GetEnabledPlatforms() []string {
	var enabled []string
	for name, platform := range c.Platforms {
		if platform.Enabled {
			enabled = append(enabled, name)
		}
	}
	return enabled
}