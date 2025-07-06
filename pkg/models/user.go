package models

import (
	"time"
)

type User struct {
	ID          string            `json:"id" yaml:"id"`
	Name        string            `json:"name" yaml:"name"`
	Email       string            `json:"email" yaml:"email"`
	Username    string            `json:"username,omitempty" yaml:"username,omitempty"`
	Avatar      string            `json:"avatar,omitempty" yaml:"avatar,omitempty"`
	Platform    Platform          `json:"platform" yaml:"platform"`
	Active      bool              `json:"active" yaml:"active"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	Metadata    map[string]any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

func NewUser(id, name, email string, platform Platform) *User {
	now := time.Now()
	return &User{
		ID:        id,
		Name:      name,
		Email:     email,
		Platform:  platform,
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}
}

func (u *User) SetMetadata(key string, value any) {
	if u.Metadata == nil {
		u.Metadata = make(map[string]any)
	}
	u.Metadata[key] = value
	u.UpdatedAt = time.Now()
}

func (u *User) GetMetadata(key string) (any, bool) {
	if u.Metadata == nil {
		return nil, false
	}
	value, exists := u.Metadata[key]
	return value, exists
}

func (u *User) DisplayName() string {
	if u.Name != "" {
		return u.Name
	}
	if u.Username != "" {
		return u.Username
	}
	return u.Email
}