package main

import (
	"opentask/cmd"
	
	// Import platform implementations to register them
	_ "opentask/pkg/platforms/linear"
	_ "opentask/pkg/platforms/jira"
)

func main() {
	cmd.Execute()
}
