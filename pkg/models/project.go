package models

import (
	"time"
)

type Project struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Key         string            `json:"key,omitempty" yaml:"key,omitempty"`
	Platform    Platform          `json:"platform" yaml:"platform"`
	Active      bool              `json:"active" yaml:"active"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	Metadata    map[string]any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func NewProject(id, name string, platform Platform) *Project {
	now := time.Now()
	return &Project{
		ID:        id,
		Name:      name,
		Platform:  platform,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}
}

func (p *Project) SetMetadata(key string, value any) {
	if p.Metadata == nil {
		p.Metadata = make(map[string]any)
	}
	p.Metadata[key] = value
	p.UpdatedAt = time.Now()
}

func (p *Project) GetMetadata(key string) (any, bool) {
	if p.Metadata == nil {
		return nil, false
	}
	value, exists := p.Metadata[key]
	return value, exists
}

func (p *Project) DisplayName() string {
	if p.Key != "" {
		return p.Key
	}
	return p.Name
}