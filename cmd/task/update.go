package task

import (
	"context"
	"fmt"
	"time"

	"opentask/pkg/config"
	"opentask/pkg/models"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <task-id>",
	Short: "Update a task",
	Long: `Update a task by ID. Currently supports updating task status.

Available statuses:
- open
- in_progress  
- done
- cancelled

Examples:
  opentask task update TASK-123 --status done
  opentask task update LIN-456 --status in_progress`,
	Args: cobra.ExactArgs(1),
	RunE: runUpdate,
}

var (
	updateStatus   string
	updatePlatform string
)

func init() {
	updateCmd.Flags().StringVarP(&updateStatus, "status", "s", "", "update task status (open, in_progress, done, cancelled)")
	updateCmd.Flags().StringVarP(&updatePlatform, "platform", "p", "", "specify platform if task ID is ambiguous")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	if updateStatus == "" {
		return fmt.Errorf("no updates specified. Use --status to update task status")
	}

	// Validate status
	status := models.TaskStatus(updateStatus)
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %s. Valid statuses: open, in_progress, done, cancelled", updateStatus)
	}

	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	// Find the task across all platforms
	task, platform, err := findTaskByID(cfg, taskID, updatePlatform)
	if err != nil {
		return err
	}

	// Create platform client
	client, err := createPlatformClient(platform, cfg.Platforms[platform])
	if err != nil {
		return fmt.Errorf("failed to create %s client: %w", platform, err)
	}

	// Update the task
	originalStatus := task.Status
	task.SetStatus(status)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	updatedTask, err := client.UpdateTask(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	fmt.Printf("✅ Task %s updated successfully\n", taskID)
	fmt.Printf("   Status: %s → %s\n", originalStatus, updatedTask.Status)

	return nil
}

func findTaskByID(cfg *config.Config, taskID string, preferredPlatform string) (*models.Task, string, error) {
	var foundTasks []*models.Task
	var foundPlatforms []string

	// If platform is specified, only search in that platform
	platforms := cfg.GetEnabledPlatforms()
	if preferredPlatform != "" {
		if _, exists := cfg.GetPlatform(preferredPlatform); !exists {
			return nil, "", fmt.Errorf("platform %s not configured", preferredPlatform)
		}
		platforms = []string{preferredPlatform}
	}

	for _, platformName := range platforms {
		platform, exists := cfg.GetPlatform(platformName)
		if !exists || !platform.Enabled {
			continue
		}

		client, err := createPlatformClient(platformName, platform)
		if err != nil {
			fmt.Printf("⚠ Failed to create %s client: %v\n", platformName, err)
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			// Task not found in this platform, continue to next
			continue
		}

		foundTasks = append(foundTasks, task)
		foundPlatforms = append(foundPlatforms, platformName)
	}

	if len(foundTasks) == 0 {
		return nil, "", fmt.Errorf("task %s not found in any configured platform", taskID)
	}

	if len(foundTasks) > 1 {
		fmt.Printf("Multiple tasks found with ID %s:\n", taskID)
		for i, task := range foundTasks {
			fmt.Printf("  %d. %s (%s) - %s\n", i+1, task.ID, foundPlatforms[i], task.Title)
		}
		return nil, "", fmt.Errorf("ambiguous task ID. Use --platform to specify which platform")
	}

	return foundTasks[0], foundPlatforms[0], nil
}
