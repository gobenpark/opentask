package main

import (
	"opentask/cmd"

	_ "opentask/pkg/platforms/jira"
	// Import platform implementations to register them
	_ "opentask/pkg/platforms/linear"
)

func main() {
	cmd.Execute()
}
