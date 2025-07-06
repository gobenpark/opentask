package task

import (
	"context"
	"fmt"
	"time"

	"opentask/pkg/config"
	"opentask/pkg/models"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [title]",
	Short: "Create a new task",
	Long: `Create a new task on the specified platform.
	
If no platform is specified, the default platform from configuration will be used.
You can specify multiple platforms to create the task on all of them.`,
	RunE: runCreate,
}

var (
	createPlatform  string
	createPlatforms []string
	createAssignee  string
	createPriority  string
	createProject   string
	createLabels    []string
	createDueDate   string
	createSyncTo    []string
)

func init() {
	createCmd.Flags().StringVarP(&createPlatform, "platform", "p", "", "platform to create task on")
	createCmd.Flags().StringSliceVar(&createPlatforms, "platforms", []string{}, "platforms to create task on")
	createCmd.Flags().StringVarP(&createAssignee, "assignee", "a", "", "task assignee")
	createCmd.Flags().StringVar(&createPriority, "priority", "", "task priority (low, medium, high, urgent)")
	createCmd.Flags().StringVar(&createProject, "project", "", "project ID or key")
	createCmd.Flags().StringSliceVarP(&createLabels, "labels", "l", []string{}, "task labels")
	createCmd.Flags().StringVar(&createDueDate, "due", "", "due date (YYYY-MM-DD)")
	createCmd.Flags().StringSliceVar(&createSyncTo, "sync-to", []string{}, "sync task to additional platforms")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task title is required")
	}

	title := args[0]
	description := ""
	if len(args) > 1 {
		description = args[1]
	}

	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	platforms := determinePlatforms(cfg)
	if len(platforms) == 0 {
		return fmt.Errorf("no platforms configured. Use 'opentask connect' to add platforms")
	}

	priority := determinePriority(cfg)
	assignee := determineAssignee(cfg)

	var createdTasks []*models.Task

	for _, platformName := range platforms {
		platform, exists := cfg.GetPlatform(platformName)
		if !exists {
			fmt.Printf("⚠ Platform %s not configured, skipping\n", platformName)
			continue
		}

		if !platform.Enabled {
			fmt.Printf("⚠ Platform %s is disabled, skipping\n", platformName)
			continue
		}

		task := createTask(title, description, platformName, priority, assignee)

		// Create platform client
		client, err := createPlatformClient(platformName, platform)
		if err != nil {
			fmt.Printf("⚠ Failed to create %s client: %v\n", platformName, err)
			continue
		}

		// Create task on platform
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		createdTask, err := client.CreateTask(ctx, task)
		if err != nil {
			fmt.Printf("⚠ Failed to create task on %s: %v\n", platformName, err)
			continue
		}

		createdTasks = append(createdTasks, createdTask)
		fmt.Printf("✓ Created task %s on %s: %s\n", createdTask.ID, platformName, createdTask.Title)
	}

	if len(createdTasks) == 0 {
		return fmt.Errorf("failed to create task on any platform")
	}

	fmt.Printf("\nSuccessfully created %d task(s)\n", len(createdTasks))

	return nil
}

func determinePlatforms(cfg *config.Config) []string {
	var platforms []string

	// Use explicit platforms first
	if len(createPlatforms) > 0 {
		platforms = append(platforms, createPlatforms...)
	} else if createPlatform != "" {
		platforms = append(platforms, createPlatform)
	} else if cfg.Defaults.Platform != "" {
		platforms = append(platforms, cfg.Defaults.Platform)
	} else {
		// Use first enabled platform
		for name, platform := range cfg.Platforms {
			if platform.Enabled {
				platforms = append(platforms, name)
				break
			}
		}
	}

	// Add sync-to platforms
	if len(createSyncTo) > 0 {
		platforms = append(platforms, createSyncTo...)
	}

	return platforms
}

func determinePriority(cfg *config.Config) models.Priority {
	if createPriority != "" {
		return models.Priority(createPriority)
	}

	if cfg.Defaults.Priority != "" {
		return models.Priority(cfg.Defaults.Priority)
	}

	return models.PriorityMedium
}

func determineAssignee(cfg *config.Config) string {
	if createAssignee != "" {
		return createAssignee
	}

	if cfg.Defaults.Assignee != "" {
		return cfg.Defaults.Assignee
	}

	return ""
}

func createTask(title, description, platformName string, priority models.Priority, assignee string) *models.Task {
	platform := models.Platform(platformName)
	task := models.NewTask(title, platform)

	if description != "" {
		task.Description = description
	}

	task.SetPriority(priority)

	if assignee != "" {
		// In a real implementation, we would resolve the assignee to a User object
		// For now, we just store the assignee string in metadata
		task.SetMetadata("assignee_query", assignee)
	}

	if createProject != "" {
		task.ProjectID = createProject
	}

	for _, label := range createLabels {
		task.AddLabel(label)
	}

	if createDueDate != "" {
		task.SetMetadata("due_date_string", createDueDate)
	}

	return task
}
