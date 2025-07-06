# CLAUDE.md
# OpenTask CLI - Multi-Platform Task Management Tool

## Project Overview

OpenTask is a unified command-line interface (CLI) tool for managing tasks across multiple platforms including Linear, Jira, Slack, and GitHub Issues. Unlike existing single-platform CLI tools, OpenTask provides a seamless developer experience by integrating all task management workflows into a single, consistent interface.

## Vision & Goals

### Primary Goal
Create a Git-like CLI experience for task management that allows developers to work with multiple task management platforms without context switching between different tools and interfaces.

### Key Differentiators
- **Unified Interface**: Single CLI for multiple platforms (Linear, Jira, Slack, GitHub)
- **Remote State Management**: Team-wide configuration synchronization
- **Developer-Centric**: Git-inspired workflow and command structure
- **Plugin Architecture**: Easily extensible for new platforms
- **Context-Aware**: Project and workspace-specific configurations

## Technical Architecture

### Core Technology Stack
- **Language**: Go (Golang)
- **CLI Framework**: Cobra + Viper + Bubble tea
- **Configuration**: YAML/JSON with environment variable support
- **Authentication**: OAuth2, API tokens, and platform-specific auth methods
- **Storage**: Local file system with optional remote synchronization

### Project Structure
```
opentask/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── init.go            # Project initialization
│   ├── connect.go         # Platform connection management
│   ├── task/              # Task management commands
│   │   ├── create.go
│   │   ├── list.go
│   │   ├── update.go
│   │   └── delete.go
│   ├── sync.go            # Synchronization commands
│   └── workspace.go       # Workspace management
├── pkg/                   # Core packages
│   ├── auth/              # Authentication handlers
│   ├── platforms/         # Platform integrations
│   │   ├── linear/
│   │   ├── jira/
│   │   ├── slack/
│   │   └── github/
│   ├── config/            # Configuration management
│   ├── models/            # Unified data models
│   └── sync/              # Synchronization logic
├── internal/              # Internal packages
├── docs/                  # Documentation
├── examples/              # Usage examples
└── scripts/               # Build and deployment scripts
```

## Core Features & Commands

### 1. Project Initialization
```bash
opentask init                    # Initialize project configuration
opentask init --template <name>  # Use predefined template
```

### 2. Platform Connection Management
```bash
opentask connect linear          # Connect to Linear
opentask connect jira --server <url>  # Connect to Jira instance
opentask connect slack --workspace <name>  # Connect to Slack workspace
opentask disconnect <platform>   # Disconnect from platform
opentask status                  # Show connection status
```

### 3. Task Management
```bash
# Create tasks
opentask task create "Fix login bug" --platform linear
opentask task create "API documentation" --platform jira --assignee john
opentask task create "Deploy v2.0" --sync-to linear,jira

# List and filter tasks
opentask task list                     # List all tasks
opentask task list --platform linear  # Platform-specific
opentask task list --status open      # Filter by status
opentask task list --assignee me      # Filter by assignee

# Update tasks
opentask task update TASK-123 --status done
opentask task update LIN-456 --assignee alice
opentask task assign TASK-123 --to bob

# Delete tasks
opentask task delete TASK-123
```

### 4. Workspace Management
```bash
opentask workspace create prod       # Create workspace
opentask workspace switch dev        # Switch workspace
opentask workspace list              # List workspaces
opentask workspace sync              # Sync workspace config
```

### 5. Synchronization
```bash
opentask sync                        # Sync all platforms
opentask sync --platform linear     # Sync specific platform
opentask sync --push                # Push local changes
opentask sync --pull                # Pull remote changes
```

## Platform Integration Specifications

### Linear Integration
- **API**: GraphQL API
- **Authentication**: OAuth2 or Personal Access Token
- **Features**: Issues, projects, teams, workflows
- **Library**: Custom GraphQL client or existing Go SDK

### Jira Integration
- **API**: REST API v2/v3
- **Authentication**: Basic Auth with API Token
- **Features**: Issues, projects, workflows, custom fields
- **Library**: `github.com/andygrunwald/go-jira`

### Slack Integration
- **API**: Web API + Socket Mode
- **Authentication**: Bot Token + App-Level Token
- **Features**: Message posting, channel management, workflow notifications
- **Library**: `github.com/slack-go/slack`

### GitHub Integration
- **API**: REST API v4
- **Authentication**: Personal Access Token or GitHub App
- **Features**: Issues, pull requests, projects
- **Library**: `github.com/google/go-github`

## Data Models

### Unified Task Model
```go
type Task struct {
    ID          string            `json:"id" yaml:"id"`
    Title       string            `json:"title" yaml:"title"`
    Description string            `json:"description,omitempty" yaml:"description,omitempty"`
    Status      TaskStatus        `json:"status" yaml:"status"`
    Priority    Priority          `json:"priority,omitempty" yaml:"priority,omitempty"`
    Assignee    *User             `json:"assignee,omitempty" yaml:"assignee,omitempty"`
    Platform    Platform          `json:"platform" yaml:"platform"`
    ProjectID   string            `json:"project_id,omitempty" yaml:"project_id,omitempty"`
    Labels      []string          `json:"labels,omitempty" yaml:"labels,omitempty"`
    CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
    UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
    DueDate     *time.Time        `json:"due_date,omitempty" yaml:"due_date,omitempty"`
    Metadata    map[string]any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type TaskStatus string
const (
    StatusOpen       TaskStatus = "open"
    StatusInProgress TaskStatus = "in_progress"
    StatusDone       TaskStatus = "done"
    StatusCancelled  TaskStatus = "cancelled"
)

type Priority string
const (
    PriorityLow    Priority = "low"
    PriorityMedium Priority = "medium"
    PriorityHigh   Priority = "high"
    PriorityUrgent Priority = "urgent"
)
```

### Configuration Model
```go
type Config struct {
    Version    string              `yaml:"version"`
    Workspace  string              `yaml:"workspace"`
    Platforms  map[string]Platform `yaml:"platforms"`
    Defaults   Defaults            `yaml:"defaults"`
    RemoteSync *RemoteSync         `yaml:"remote_sync,omitempty"`
}

type Platform struct {
    Type         string            `yaml:"type"`
    Enabled      bool              `yaml:"enabled"`
    Credentials  map[string]string `yaml:"credentials"`
    Settings     map[string]any    `yaml:"settings"`
}
```

## Development Phases

### Phase 1: MVP (Months 1-3)
**Goal**: Basic functionality with Linear and Jira support

**Deliverables**:
- [ ] Project scaffolding with Cobra/Viper
- [ ] Configuration management system
- [ ] Linear integration (OAuth + basic CRUD)
- [ ] Jira integration (API token + basic CRUD)
- [ ] Local task storage and caching
- [ ] Basic CLI commands (init, connect, task CRUD)
- [ ] Unit tests and documentation

**Success Criteria**:
- Users can connect to Linear and Jira
- Users can create, list, update, delete tasks
- Configuration persists across sessions

### Phase 2: Enhanced Features (Months 4-6)
**Goal**: Slack integration and improved UX

**Deliverables**:
- [ ] Slack integration (notifications, task creation from messages)
- [ ] GitHub Issues integration
- [ ] Workspace management
- [ ] Advanced filtering and search
- [ ] Task synchronization between platforms
- [ ] Configuration templates
- [ ] Shell completion (bash, zsh, fish)

**Success Criteria**:
- Multi-platform task synchronization works
- Users can manage multiple workspaces
- Slack notifications are reliable

### Phase 3: Advanced Features (Months 7-9)
**Goal**: Remote synchronization and team collaboration

**Deliverables**:
- [ ] Remote configuration synchronization (Git-based)
- [ ] Team collaboration features
- [ ] Plugin system architecture
- [ ] Custom workflow definitions
- [ ] Advanced reporting and analytics
- [ ] CI/CD integrations
- [ ] Performance optimizations

**Success Criteria**:
- Teams can share configurations
- Plugin system supports third-party extensions
- Performance meets enterprise requirements

## Technical Requirements

### Authentication Management
```go
type AuthProvider interface {
    Authenticate(ctx context.Context) (*AuthToken, error)
    RefreshToken(ctx context.Context, token *AuthToken) (*AuthToken, error)
    RevokeToken(ctx context.Context, token *AuthToken) error
    ValidateToken(ctx context.Context, token *AuthToken) error
}

type AuthToken struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token,omitempty"`
    TokenType    string    `json:"token_type"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scopes       []string  `json:"scopes"`
}
```

### Platform Abstraction
```go
type PlatformClient interface {
    // Task operations
    CreateTask(ctx context.Context, task *Task) (*Task, error)
    GetTask(ctx context.Context, id string) (*Task, error)
    UpdateTask(ctx context.Context, task *Task) (*Task, error)
    DeleteTask(ctx context.Context, id string) error
    ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)
    
    // Project operations
    ListProjects(ctx context.Context) ([]*Project, error)
    GetProject(ctx context.Context, id string) (*Project, error)
    
    // User operations
    GetCurrentUser(ctx context.Context) (*User, error)
    SearchUsers(ctx context.Context, query string) ([]*User, error)
    
    // Platform-specific
    GetPlatformInfo() PlatformInfo
    HealthCheck(ctx context.Context) error
}
```

### Error Handling
```go
type OpenTaskError struct {
    Code     ErrorCode `json:"code"`
    Message  string    `json:"message"`
    Platform string    `json:"platform,omitempty"`
    TaskID   string    `json:"task_id,omitempty"`
    Cause    error     `json:"-"`
}

type ErrorCode string
const (
    ErrAuthentication ErrorCode = "authentication_failed"
    ErrNotFound      ErrorCode = "not_found"
    ErrInvalidInput  ErrorCode = "invalid_input"
    ErrPlatformAPI   ErrorCode = "platform_api_error"
    ErrSyncConflict  ErrorCode = "sync_conflict"
)
```

## Testing Strategy

### Unit Tests
- All core functionality with >80% coverage
- Mock platform clients for isolated testing
- Configuration parsing and validation
- Authentication flow testing

### Integration Tests
- Real API calls with test accounts/sandboxes
- End-to-end command testing
- Multi-platform synchronization scenarios
- Error handling and recovery

### Performance Tests
- Large task list handling
- Concurrent API operations
- Memory and CPU usage profiling
- Startup time optimization

## Documentation Requirements

### User Documentation
- [ ] Installation guide (Homebrew, direct download, go install)
- [ ] Quick start tutorial
- [ ] Platform connection guides
- [ ] Command reference
- [ ] Configuration reference
- [ ] Troubleshooting guide

### Developer Documentation
- [ ] Architecture overview
- [ ] Platform integration guide
- [ ] Plugin development guide
- [ ] API reference
- [ ] Contributing guidelines
- [ ] Code style guide

## Security Considerations

### Credential Management
- Store tokens securely using OS keychain/keyring
- Support for credential rotation
- Secure token transmission (HTTPS only)
- Option to use environment variables

### Data Privacy
- Local-first approach (data stays on user's machine)
- Optional remote sync with encryption
- No unnecessary data collection
- GDPR compliance for EU users

### Platform Permissions
- Principle of least privilege
- Clear permission requirements documentation
- Regular security audit of dependencies
- Vulnerability scanning in CI/CD

## Deployment & Distribution

### Build System
```bash
# Cross-platform builds
make build-all          # Build for all platforms
make build-linux        # Linux amd64/arm64
make build-darwin       # macOS amd64/arm64
make build-windows      # Windows amd64

# Release process
make release VERSION=v1.0.0  # Create release artifacts
```

### Distribution Channels
- **GitHub Releases**: Primary distribution
- **Homebrew**: macOS package manager
- **Chocolatey**: Windows package manager
- **AUR**: Arch Linux user repository
- **Docker**: Containerized version
- **go install**: Direct Go installation

### CI/CD Pipeline
- GitHub Actions for automated testing
- Cross-platform build verification
- Security scanning (gosec, nancy)
- Automated release generation
- Documentation deployment

## Success Metrics

### Technical Metrics
- Command execution time < 2 seconds
- API response caching efficiency
- Error rate < 1% for platform operations
- Test coverage > 80%
- Zero critical security vulnerabilities

### User Metrics
- Daily active users
- Platform connection success rate
- Task operation success rate
- User retention (weekly/monthly)
- GitHub stars and community engagement

### Business Metrics
- Community growth rate
- Enterprise adoption
- Plugin ecosystem development
- Documentation completeness
- Support ticket resolution time

## Risk Assessment & Mitigation

### Technical Risks
1. **Platform API Changes**: Implement versioning and graceful degradation
2. **Rate Limiting**: Implement backoff strategies and caching
3. **Authentication Complexity**: Use well-tested OAuth libraries
4. **Data Synchronization**: Implement conflict resolution strategies

### Business Risks
1. **Platform Policy Changes**: Maintain good relationships with platform vendors
2. **Competition**: Focus on unique value proposition and user experience
3. **Maintenance Burden**: Build sustainable development practices
4. **Security Issues**: Implement comprehensive security practices

## Getting Started for Contributors

### Prerequisites
- Go 1.21+ installed
- Git configured
- Platform accounts for testing (Linear, Jira, etc.)
- IDE with Go support (VS Code, GoLand, etc.)

### Development Setup
```bash
# Clone repository
git clone https://github.com/your-org/opentask.git
cd opentask

# Install dependencies
go mod download

# Install development tools
make install-tools

# Run tests
make test

# Build local version
make build

# Install locally
make install
```

### First Contribution
1. Check out the [good first issue](https://github.com/your-org/opentask/labels/good%20first%20issue) label
2. Read the contributing guidelines
3. Set up development environment
4. Make your changes
5. Run tests and linting
6. Submit a pull request

## Conclusion

OpenTask represents a significant opportunity to improve developer productivity by unifying task management across multiple platforms. By leveraging Go's strengths and the mature CLI ecosystem, we can create a tool that developers actually want to use.

The phased approach ensures we can deliver value quickly while building toward a comprehensive solution. The focus on developer experience, extensibility, and performance will differentiate OpenTask from existing solutions.

This project aligns with current trends toward developer tooling consolidation and the growing importance of developer experience in tool adoption. With proper execution, OpenTask can become an essential tool in the modern developer's toolkit.