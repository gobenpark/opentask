package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"opentask/pkg/config"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize OpenTask configuration",
	Long: `Initialize OpenTask configuration in the current directory or home directory.
	
This command creates a new .opentask.yaml configuration file with default settings.
If a configuration file already exists, it will ask for confirmation before overwriting.`,
	RunE: runInit,
}

var (
	initForce    bool
	initTemplate string
	initGlobal   bool
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "overwrite existing configuration")
	initCmd.Flags().StringVarP(&initTemplate, "template", "t", "", "use configuration template")
	initCmd.Flags().BoolVarP(&initGlobal, "global", "g", false, "initialize global configuration")
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := getConfigPath()

	if !initForce && configExists(configPath) {
		fmt.Printf("Configuration file already exists at: %s\n", configPath)
		fmt.Print("Do you want to overwrite it? [y/N]: ")

		var response string
		fmt.Scanln(&response)

		if response != "y" && response != "Y" {
			fmt.Println("Configuration initialization cancelled.")
			return nil
		}
	}

	manager := config.NewManager()
	cfg := config.NewConfig()

	if initTemplate != "" {
		var err error
		cfg, err = loadTemplate(initTemplate)
		if err != nil {
			return fmt.Errorf("failed to load template: %w", err)
		}
	}

	manager.SetConfig(cfg)

	if err := manager.Save(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("OpenTask configuration initialized at: %s\n", configPath)

	if initTemplate != "" {
		fmt.Printf("Using template: %s\n", initTemplate)
	}

	fmt.Println("\nNext steps:")
	fmt.Println("1. Connect to your platforms: opentask connect <platform>")
	fmt.Println("2. List available platforms: opentask connect --list")
	fmt.Println("3. View configuration: opentask config show")

	return nil
}

func getConfigPath() string {
	if initGlobal {
		home, err := os.UserHomeDir()
		if err != nil {
			return config.DefaultConfigFile
		}
		return filepath.Join(home, config.DefaultConfigFile)
	}

	return config.DefaultConfigFile
}

func configExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadTemplate(templateName string) (*config.Config, error) {
	switch templateName {
	case "basic":
		return createBasicTemplate(), nil
	case "dev":
		return createDevTemplate(), nil
	case "enterprise":
		return createEnterpriseTemplate(), nil
	default:
		return nil, fmt.Errorf("unknown template: %s", templateName)
	}
}

func createBasicTemplate() *config.Config {
	cfg := config.NewConfig()
	cfg.Defaults.Platform = "linear"
	cfg.Defaults.Priority = "medium"
	return cfg
}

func createDevTemplate() *config.Config {
	cfg := config.NewConfig()
	cfg.Defaults.Platform = "github"
	cfg.Defaults.Priority = "high"
	cfg.Defaults.Assignee = "me"
	return cfg
}

func createEnterpriseTemplate() *config.Config {
	cfg := config.NewConfig()
	cfg.Defaults.Platform = "jira"
	cfg.Defaults.Priority = "medium"
	cfg.RemoteSync = &config.RemoteSync{
		Type:     "git",
		Branch:   "main",
		Enabled:  false,
		Interval: "1h",
	}
	return cfg
}
