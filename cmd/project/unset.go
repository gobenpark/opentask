package project

import (
	"fmt"

	"opentask/pkg/config"

	"github.com/spf13/cobra"
)

var unsetCmd = &cobra.Command{
	Use:   "unset",
	Short: "Unset default project",
	Long: `Remove the default project configuration.
	
After unsetting the default project, you will need to specify
the project explicitly when listing or creating tasks.`,
	RunE: runProjectUnset,
}

func runProjectUnset(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	if cfg.Defaults.Project == "" {
		fmt.Println("No default project is currently set.")
		return nil
	}

	previousProject := cfg.Defaults.Project

	// Clear the default project
	cfg.Defaults.Project = ""

	// Update the configuration
	manager.SetConfig(cfg)

	// Save the configuration
	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("✓ Default project unset (was: %s)\n", previousProject)
	return nil
}
