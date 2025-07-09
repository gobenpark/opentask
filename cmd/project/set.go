package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"opentask/pkg/config"
	"opentask/pkg/platforms"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <project-id>",
	Short: "Set default project",
	Long: `Set the default project for the current workspace.
	
The project ID should be a valid project identifier from one of your 
configured platforms. You can use "opentask project list" to see 
available projects.`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectSet,
}

var (
	setPlatform string
	setValidate bool
)

func init() {
	setCmd.Flags().StringVarP(&setPlatform, "platform", "p", "", "specify platform for project lookup")
	setCmd.Flags().BoolVar(&setValidate, "validate", true, "validate project exists before setting")
}

func runProjectSet(cmd *cobra.Command, args []string) error {
	projectID := args[0]

	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	// Validate project exists if validation is enabled
	if setValidate {
		if err := validateProjectExists(cfg, projectID, setPlatform); err != nil {
			return fmt.Errorf("project validation failed: %w", err)
		}
	}

	// Set the default project
	cfg.Defaults.Project = projectID

	// Update the configuration
	manager.SetConfig(cfg)

	// Save the configuration
	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Default project set to: %s\n", projectID)
	return nil
}

func validateProjectExists(cfg *config.Config, projectID string, platformFilter string) error {
	platforms := cfg.GetEnabledPlatforms()

	if platformFilter != "" {
		platforms = []string{platformFilter}
	}

	if len(platforms) == 0 {
		return fmt.Errorf("no platforms configured or enabled")
	}

	for _, platformName := range platforms {
		platform, exists := cfg.GetPlatform(platformName)
		if !exists {
			continue
		}

		if !platform.Enabled {
			continue
		}

		// Create platform client
		client, err := createPlatformClientForProject(platformName, platform)
		if err != nil {
			fmt.Printf("⚠ Failed to create %s client: %v\n", platformName, err)
			continue
		}

		// Check if project exists
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		project, err := client.GetProject(ctx, projectID)
		if err != nil {
			// Check if it's a "not found" error
			if isNotFoundError(err) {
				continue // Try next platform
			}
			fmt.Printf("⚠ Failed to check project in %s: %v\n", platformName, err)
			continue
		}

		if project != nil {
			fmt.Printf("✓ Project found: %s (%s) on %s\n", project.DisplayName(), project.Name, platformName)
			return nil
		}
	}

	return fmt.Errorf("project '%s' not found in any configured platform", projectID)
}

func isNotFoundError(err error) bool {
	// Check if the error is a "not found" type error
	// This would need to be implemented based on your platform error types
	errorMsg := strings.ToLower(err.Error())
	return strings.Contains(errorMsg, "not found") ||
		strings.Contains(errorMsg, "404") ||
		strings.Contains(errorMsg, "does not exist")
}

// Helper function to create platform client
func createPlatformClientForProject(platformName string, platform config.Platform) (platforms.PlatformClient, error) {
	// Prepare configuration for platform factory
	clientConfig := make(map[string]any)

	// Copy credentials
	for key, value := range platform.Credentials {
		clientConfig[key] = value
	}

	// Copy settings
	for key, value := range platform.Settings {
		clientConfig[key] = value
	}

	// Create client using registry
	client, err := platforms.DefaultRegistry.Create(platform.Type, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", platformName, err)
	}

	return client, nil
}
