package cmd

import (
	"context"
	"fmt"
	"os"

	"opentask/cmd/project"
	"opentask/cmd/task"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "opentask",
	Short: "OpenTask - Multi-Platform Task Management CLI",
	Long: `OpenTask is a unified command-line interface for managing tasks across 
multiple platforms including Linear, Jira, Slack, and GitHub Issues.

Unlike existing single-platform CLI tools, OpenTask provides a seamless 
developer experience by integrating all task management workflows into 
a single, consistent interface.`,
	Version: "0.1.0",
}

func Execute() {
	if err := fang.Execute(context.Background(), rootCmd); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.opentask.yaml)")
	rootCmd.PersistentFlags().StringP("workspace", "w", "", "workspace to use")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "debug mode")

	viper.BindPFlag("workspace", rootCmd.PersistentFlags().Lookup("workspace"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Add subcommands
	rootCmd.AddCommand(task.TaskCmd)
	rootCmd.AddCommand(project.ProjectCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".opentask")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("debug") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
