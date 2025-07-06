package platforms

import (
	"context"
	"opentask/pkg/models"
)

type PlatformClient interface {
	// Task operations
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTask(ctx context.Context, id string) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	DeleteTask(ctx context.Context, id string) error
	ListTasks(ctx context.Context, filter *models.TaskFilter) ([]*models.Task, error)

	// Project operations
	ListProjects(ctx context.Context) ([]*models.Project, error)
	GetProject(ctx context.Context, id string) (*models.Project, error)

	// User operations
	GetCurrentUser(ctx context.Context) (*models.User, error)
	SearchUsers(ctx context.Context, query string) ([]*models.User, error)

	// Platform-specific
	GetPlatformInfo() PlatformInfo
	HealthCheck(ctx context.Context) error
}

type PlatformInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Version     string `json:"version"`
	Description string `json:"description"`
	BaseURL     string `json:"base_url,omitempty"`
}

type AuthProvider interface {
	Authenticate(ctx context.Context) (*AuthToken, error)
	RefreshToken(ctx context.Context, token *AuthToken) (*AuthToken, error)
	RevokeToken(ctx context.Context, token *AuthToken) error
	ValidateToken(ctx context.Context, token *AuthToken) error
}

type AuthToken struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenType    string   `json:"token_type"`
	ExpiresAt    int64    `json:"expires_at"`
	Scopes       []string `json:"scopes"`
}

type PlatformFactory interface {
	Create(config map[string]any) (PlatformClient, error)
	GetType() string
	GetName() string
	ValidateConfig(config map[string]any) error
}

type Registry struct {
	factories map[string]PlatformFactory
}

func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]PlatformFactory),
	}
}

func (r *Registry) Register(factory PlatformFactory) {
	r.factories[factory.GetType()] = factory
}

func (r *Registry) Create(platformType string, config map[string]any) (PlatformClient, error) {
	factory, exists := r.factories[platformType]
	if !exists {
		return nil, NewPlatformError(ErrPlatformNotSupported, platformType, "", nil)
	}

	if err := factory.ValidateConfig(config); err != nil {
		return nil, NewPlatformError(ErrInvalidConfig, platformType, "", err)
	}

	return factory.Create(config)
}

func (r *Registry) GetSupportedPlatforms() []string {
	platforms := make([]string, 0, len(r.factories))
	for platformType := range r.factories {
		platforms = append(platforms, platformType)
	}
	return platforms
}

func (r *Registry) IsSupported(platformType string) bool {
	_, exists := r.factories[platformType]
	return exists
}

var DefaultRegistry = NewRegistry()