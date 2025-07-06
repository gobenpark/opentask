package cmd

import (
	"fmt"

	"opentask/pkg/config"

	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect [platform]",
	Short: "Connect to task management platforms",
	Long: `Connect to various task management platforms like Linear, Jira, Slack, or GitHub.
	
This command helps you authenticate and configure connections to different platforms.
Use --list to see all available platforms.`,
	RunE: runConnect,
}

var (
	connectList   bool
	connectServer string
	connectToken  string
	connectForce  bool
)

func init() {
	rootCmd.AddCommand(connectCmd)

	connectCmd.Flags().BoolVarP(&connectList, "list", "l", false, "list available platforms")
	connectCmd.Flags().StringVarP(&connectServer, "server", "s", "", "server URL (for self-hosted platforms)")
	connectCmd.Flags().StringVarP(&connectToken, "token", "t", "", "authentication token")
	connectCmd.Flags().BoolVarP(&connectForce, "force", "f", false, "force reconnection")
}

func runConnect(cmd *cobra.Command, args []string) error {
	if connectList {
		return listPlatforms()
	}

	if len(args) == 0 {
		return fmt.Errorf("platform name is required. Use --list to see available platforms")
	}

	platformName := args[0]

	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	if !connectForce {
		if platform, exists := cfg.GetPlatform(platformName); exists && platform.Enabled {
			fmt.Printf("Platform %s is already connected.\n", platformName)
			fmt.Print("Do you want to reconnect? [y/N]: ")

			var response string
			fmt.Scanln(&response)

			if response != "y" && response != "Y" {
				fmt.Println("Connection cancelled.")
				return nil
			}
		}
	}

	return connectToPlatform(platformName, cfg, manager)
}

func listPlatforms() error {
	fmt.Println("Available platforms:")
	fmt.Println("  linear   - Linear (https://linear.app)")
	fmt.Println("  jira     - Jira (https://www.atlassian.com/software/jira)")
	fmt.Println("  slack    - Slack (https://slack.com)")
	fmt.Println("  github   - GitHub Issues (https://github.com)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  opentask connect linear")
	fmt.Println("  opentask connect jira --server https://company.atlassian.net")
	fmt.Println("  opentask connect slack --token xoxb-...")
	fmt.Println("  opentask connect github --token ghp_...")

	return nil
}

func connectToPlatform(platformName string, cfg *config.Config, manager *config.Manager) error {
	switch platformName {
	case "linear":
		return connectLinear(cfg, manager)
	case "jira":
		return connectJira(cfg, manager)
	case "slack":
		return connectSlack(cfg, manager)
	case "github":
		return connectGitHub(cfg, manager)
	default:
		return fmt.Errorf("unsupported platform: %s", platformName)
	}
}

func connectLinear(cfg *config.Config, manager *config.Manager) error {
	fmt.Println("Connecting to Linear...")

	token := connectToken
	if token == "" {
		fmt.Print("Enter your Linear API token: ")
		fmt.Scanln(&token)
	}

	if token == "" {
		return fmt.Errorf("API token is required for Linear")
	}

	platform := config.Platform{
		Type:    "linear",
		Enabled: true,
		Credentials: map[string]string{
			"token": token,
		},
		Settings: map[string]any{
			"base_url": "https://api.linear.app/graphql",
		},
	}

	cfg.AddPlatform("linear", platform)

	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✓ Successfully connected to Linear")
	return nil
}

func connectJira(cfg *config.Config, manager *config.Manager) error {
	fmt.Println("Connecting to Jira...")

	server := connectServer
	if server == "" {
		fmt.Print("Enter your Jira server URL: ")
		fmt.Scanln(&server)
	}

	token := connectToken
	if token == "" {
		fmt.Print("Enter your Jira API token: ")
		fmt.Scanln(&token)
	}

	if server == "" || token == "" {
		return fmt.Errorf("server URL and API token are required for Jira")
	}

	var email string
	fmt.Print("Enter your Jira email: ")
	fmt.Scanln(&email)

	platform := config.Platform{
		Type:    "jira",
		Enabled: true,
		Credentials: map[string]string{
			"token": token,
			"email": email,
		},
		Settings: map[string]any{
			"base_url": server,
		},
	}

	cfg.AddPlatform("jira", platform)

	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✓ Successfully connected to Jira")
	return nil
}

func connectSlack(cfg *config.Config, manager *config.Manager) error {
	fmt.Println("Connecting to Slack...")

	token := connectToken
	if token == "" {
		fmt.Print("Enter your Slack Bot Token: ")
		fmt.Scanln(&token)
	}

	if token == "" {
		return fmt.Errorf("bot token is required for Slack")
	}

	platform := config.Platform{
		Type:    "slack",
		Enabled: true,
		Credentials: map[string]string{
			"bot_token": token,
		},
		Settings: map[string]any{
			"base_url": "https://slack.com/api",
		},
	}

	cfg.AddPlatform("slack", platform)

	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✓ Successfully connected to Slack")
	return nil
}

func connectGitHub(cfg *config.Config, manager *config.Manager) error {
	fmt.Println("Connecting to GitHub...")

	token := connectToken
	if token == "" {
		fmt.Print("Enter your GitHub Personal Access Token: ")
		fmt.Scanln(&token)
	}

	if token == "" {
		return fmt.Errorf("personal access token is required for GitHub")
	}

	platform := config.Platform{
		Type:    "github",
		Enabled: true,
		Credentials: map[string]string{
			"token": token,
		},
		Settings: map[string]any{
			"base_url": "https://api.github.com",
		},
	}

	cfg.AddPlatform("github", platform)

	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✓ Successfully connected to GitHub")
	return nil
}
