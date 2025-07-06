package task

import (
	"github.com/spf13/cobra"
)

var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks across platforms",
	Long: `Manage tasks across multiple platforms including Linear, Jira, Slack, and GitHub.
	
This command provides subcommands for creating, listing, updating, and deleting tasks.`,
}

func init() {
	TaskCmd.AddCommand(createCmd)
	TaskCmd.AddCommand(listCmd)
}