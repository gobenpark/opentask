package project

import (
	"github.com/spf13/cobra"
)

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
	Long: `Manage projects across configured platforms.
	
You can list available projects, set a default project for your workspace,
and view the current default project configuration.`,
}

func init() {
	// Add subcommands
	ProjectCmd.AddCommand(listCmd)
	ProjectCmd.AddCommand(setCmd)
	ProjectCmd.AddCommand(getCmd)
	ProjectCmd.AddCommand(unsetCmd)
}
