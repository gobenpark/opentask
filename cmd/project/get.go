package project

import (
	"fmt"

	"opentask/pkg/config"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current default project",
	Long: `Display the current default project configuration.
	
Shows the project ID that is currently set as the default for 
the current workspace.`,
	RunE: runProjectGet,
}

func runProjectGet(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	if cfg.Defaults.Project == "" {
		fmt.Println("No default project is currently set.")
		fmt.Println("Use 'opentask project set <project-id>' to set a default project.")
		return nil
	}

	fmt.Printf("Default project: %s\n", cfg.Defaults.Project)

	// Show workspace info
	if cfg.Workspace != "" {
		fmt.Printf("Workspace: %s\n", cfg.Workspace)
	}

	return nil
}
