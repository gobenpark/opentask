# OpenTask

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](#)

OpenTask is a unified command-line interface (CLI) tool for managing tasks across multiple platforms including Linear, Jira, Slack, and GitHub Issues. Unlike existing single-platform CLI tools, OpenTask provides a seamless developer experience by integrating all task management workflows into a single, consistent interface.

## 🎯 Key Features

- **Unified Interface**: Single CLI for multiple platforms (Linear, Jira, Slack, GitHub)
- **Developer-Centric**: Git-inspired workflow and command structure
- **Multiple Output Formats**: Interactive tables, JSON, CSV, and plain text
- **Platform Agnostic**: Work with any combination of task management platforms
- **Configuration Management**: YAML/JSON with environment variable support
- **Extensible**: Plugin architecture for adding new platforms

## 🚀 Quick Start

### Installation

#### Option 1: Direct Download (Recommended)
```bash
# Download and build from source
git clone https://github.com/your-org/opentask.git
cd opentask
go build -o opentask
sudo mv opentask /usr/local/bin/
```

#### Option 2: Go Install
```bash
go install github.com/your-org/opentask@latest
```

### Initial Setup

1. **Initialize OpenTask configuration:**
```bash
opentask init
```

2. **Connect to your platforms:**
```bash
# Connect to Jira
opentask connect jira --server https://your-domain.atlassian.net

# Connect to Linear (coming soon)
opentask connect linear

# Connect to GitHub Issues (coming soon)
opentask connect github
```

3. **List your tasks:**
```bash
opentask task list
```

## 📋 Usage

### Task Management

#### List Tasks
```bash
# List all tasks (interactive table)
opentask task list

# List tasks in plain text format (perfect for scripts)
opentask task list --plain

# List tasks in JSON format
opentask task list --format json

# List tasks in CSV format
opentask task list --format csv

# Filter by platform
opentask task list --platform jira

# Filter by status
opentask task list --status open

# Filter by assignee
opentask task list --assignee me
```

#### Create Tasks
```bash
# Create a task with title
opentask task create "Fix login bug"

# Create a task on specific platform
opentask task create "API documentation" --platform jira

# Create a task with assignee
opentask task create "Deploy v2.0" --assignee john
```

### Platform Management

#### Connect to Platforms
```bash
# Connect to Jira
opentask connect jira --server https://your-domain.atlassian.net

# Check connection status
opentask status

# Disconnect from a platform
opentask disconnect jira
```

### Configuration

#### View Current Configuration
```bash
# Show current configuration
cat ~/.opentask.yaml
```

#### Environment Variables
You can override configuration using environment variables:

```bash
export OPENTASK_JIRA_SERVER="https://your-domain.atlassian.net"
export OPENTASK_JIRA_EMAIL="your-email@company.com"
export OPENTASK_JIRA_TOKEN="your-api-token"
```

## ⚙️ Configuration

OpenTask uses a YAML configuration file located at `~/.opentask.yaml`. Here's an example configuration:

```yaml
version: "1.0"
workspace: "default"

platforms:
  jira:
    type: "jira"
    enabled: true
    credentials:
      server: "https://your-domain.atlassian.net"
      email: "your-email@company.com"
      token: "your-api-token"
    settings:
      project_key: "DEV"
      max_results: 100

  linear:
    type: "linear"
    enabled: false
    credentials:
      api_key: ""
    settings:
      team_id: ""

defaults:
  platform: "jira"
  format: "table"
  limit: 50
```

### Platform-Specific Configuration

#### Jira Configuration
1. Generate an API token at: https://id.atlassian.com/manage-profile/security/api-tokens
2. Configure Jira connection:
```bash
opentask connect jira \
  --server https://your-domain.atlassian.net \
  --email your-email@company.com \
  --token your-api-token
```

#### Linear Configuration (Coming Soon)
```bash
opentask connect linear --api-key your-linear-api-key
```

#### GitHub Configuration (Coming Soon)
```bash
opentask connect github --token your-github-token
```

## 🔧 Advanced Usage

### Scripting and Automation

OpenTask is designed to work well in scripts and automation workflows:

```bash
#!/bin/bash

# Get all open tasks in JSON format
TASKS=$(opentask task list --status open --format json)

# Process tasks with jq
echo "$TASKS" | jq '.[] | select(.priority == "high") | .title'

# Use plain format for simple text processing
opentask task list --plain | grep "bug" | wc -l
```

### Integration with Other Tools

#### Using with fzf for Interactive Selection
```bash
# Select a task interactively
TASK_ID=$(opentask task list --plain | fzf | awk '{print $1}')
echo "Selected task: $TASK_ID"
```

#### Export to CSV for Analysis
```bash
# Export all tasks to CSV for spreadsheet analysis
opentask task list --format csv > tasks.csv
```

## 🏗️ Architecture

OpenTask is built with a modular architecture:

```
opentask/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command and global flags
│   ├── init.go            # Project initialization
│   ├── connect.go         # Platform connection management
│   └── task/              # Task management commands
├── pkg/                   # Core packages
│   ├── auth/              # Authentication handlers
│   ├── platforms/         # Platform integrations
│   │   ├── jira/
│   │   ├── linear/
│   │   ├── slack/
│   │   └── github/
│   ├── config/            # Configuration management
│   ├── models/            # Unified data models
│   └── sync/              # Synchronization logic
└── internal/              # Internal packages
```

### Technology Stack

- **Language**: Go 1.24+
- **CLI Framework**: Cobra + Viper
- **UI Framework**: Bubble Tea (for interactive tables)
- **Configuration**: YAML/JSON with environment variable support
- **Authentication**: OAuth2, API tokens, and platform-specific auth methods

## 🚧 Development Status

OpenTask is currently in active development. Here's the current platform support:

| Platform | Status | Features |
|----------|---------|----------|
| Jira | ✅ Stable | List, Create, Filter |
| Linear | 🚧 In Progress | Coming Soon |
| GitHub Issues | 📋 Planned | Coming Soon |
| Slack | 📋 Planned | Coming Soon |

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Setup

1. Clone the repository:
```bash
git clone https://github.com/your-org/opentask.git
cd opentask
```

2. Install dependencies:
```bash
go mod download
```

3. Build the project:
```bash
make build
```

4. Run tests:
```bash
make test
```

### Adding New Platforms

To add support for a new platform:

1. Create a new package in `pkg/platforms/yourplatform/`
2. Implement the `PlatformClient` interface
3. Add configuration schema
4. Update the factory method in `pkg/platforms/platform.go`
5. Add tests and documentation

## 📚 Documentation

- [API Reference](docs/api.md)
- [Platform Integration Guide](docs/platforms.md)
- [Configuration Reference](docs/configuration.md)
- [Troubleshooting](docs/troubleshooting.md)

## 🐛 Troubleshooting

### Common Issues

#### Authentication Errors
```bash
# Check your configuration
opentask status

# Verify API tokens are correct
cat ~/.opentask.yaml
```

#### Connection Issues
```bash
# Test connectivity
curl -u email@domain.com:api-token https://your-domain.atlassian.net/rest/api/2/myself
```

#### No Tasks Displayed
```bash
# Check if platform is enabled
opentask status

# Try with debug mode
opentask task list --debug
```

### Debug Mode

Enable debug mode for detailed logging:
```bash
opentask task list --debug
```

### Getting Help

- 📖 Check the [documentation](docs/)
- 🐛 Report bugs on [GitHub Issues](https://github.com/your-org/opentask/issues)
- 💬 Ask questions in [Discussions](https://github.com/your-org/opentask/discussions)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for beautiful terminal UIs
- [go-jira](https://github.com/andygrunwald/go-jira) for Jira API integration
- All contributors who help make OpenTask better

---

**OpenTask** - Unifying your task management workflow, one command at a time.