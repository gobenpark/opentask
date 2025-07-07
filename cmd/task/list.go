package task

import (
	"context"
	"fmt"
	"os"
	"time"

	"opentask/pkg/config"
	"opentask/pkg/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long: `List tasks from configured platforms.
	
You can filter tasks by platform, status, assignee, and other criteria.
By default, tasks from all enabled platforms are shown.`,
	RunE: runList,
}

var (
	listPlatform string
	listStatus   string
	listAssignee string
	listProject  string
	listLabels   []string
	listLimit    int
	listOffset   int
	listFormat   string
	listAll      bool
	listPlain    bool
)

func init() {
	listCmd.Flags().StringVarP(&listPlatform, "platform", "p", "", "filter by platform")
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "filter by status (open, in_progress, done, cancelled)")
	listCmd.Flags().StringVarP(&listAssignee, "assignee", "a", "", "filter by assignee")
	listCmd.Flags().StringVar(&listProject, "project", "", "filter by project")
	listCmd.Flags().StringSliceVarP(&listLabels, "labels", "l", []string{}, "filter by labels")
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "maximum number of tasks to show")
	listCmd.Flags().IntVar(&listOffset, "offset", 0, "number of tasks to skip")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "output format (table, json, csv)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "show tasks from all platforms")
	listCmd.Flags().BoolVar(&listPlain, "plain", false, "disable interactive mode and output plain text")
}

func runList(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	platforms := determinePlatformsForList(cfg)
	if len(platforms) == 0 {
		return fmt.Errorf("no platforms configured or enabled")
	}

	filter := createTaskFilter()

	var allTasks []*models.Task

	for _, platformName := range platforms {
		platform, exists := cfg.GetPlatform(platformName)
		if !exists {
			continue
		}

		if !platform.Enabled {
			continue
		}

		// Create platform client
		client, err := createPlatformClient(platformName, platform)
		if err != nil {
			fmt.Printf("⚠ Failed to create %s client: %v\n", platformName, err)
			continue
		}

		// Fetch tasks from platform
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		tasks, err := client.ListTasks(ctx, filter)
		if err != nil {
			fmt.Printf("⚠ Failed to list tasks from %s: %v\n", platformName, err)
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	if len(allTasks) == 0 {
		fmt.Println("No tasks found matching the criteria.")
		return nil
	}

	// Apply pagination
	start := listOffset
	end := start + listLimit
	if end > len(allTasks) {
		end = len(allTasks)
	}

	if start >= len(allTasks) {
		fmt.Println("No more tasks to show.")
		return nil
	}

	paginatedTasks := allTasks[start:end]

	switch listFormat {
	case "json":
		return printTasksJSON(paginatedTasks)
	case "csv":
		return printTasksCSV(paginatedTasks)
	default:
		return printBubbleTasksTable(paginatedTasks)
	}
}

func determinePlatformsForList(cfg *config.Config) []string {
	if listPlatform != "" {
		return []string{listPlatform}
	}

	if listAll {
		return cfg.GetEnabledPlatforms()
	}

	// Default to enabled platforms
	return cfg.GetEnabledPlatforms()
}

func createTaskFilter() *models.TaskFilter {
	filter := &models.TaskFilter{
		Limit:  listLimit,
		Offset: listOffset,
	}

	if listPlatform != "" {
		platform := models.Platform(listPlatform)
		filter.Platform = &platform
	}

	if listStatus != "" {
		status := models.TaskStatus(listStatus)
		filter.Status = &status
	}

	if listAssignee != "" {
		filter.Assignee = listAssignee
	}

	if listProject != "" {
		filter.ProjectID = listProject
	}

	if len(listLabels) > 0 {
		filter.Labels = listLabels
	}

	return filter
}

func applyFilter(tasks []*models.Task, filter *models.TaskFilter) []*models.Task {
	var filtered []*models.Task

	for _, task := range tasks {
		if filter.Platform != nil && task.Platform != *filter.Platform {
			continue
		}

		if filter.Status != nil && task.Status != *filter.Status {
			continue
		}

		if filter.ProjectID != "" && task.ProjectID != filter.ProjectID {
			continue
		}

		if len(filter.Labels) > 0 {
			hasAllLabels := true
			for _, filterLabel := range filter.Labels {
				found := false
				for _, taskLabel := range task.Labels {
					if taskLabel == filterLabel {
						found = true
						break
					}
				}
				if !found {
					hasAllLabels = false
					break
				}
			}
			if !hasAllLabels {
				continue
			}
		}

		filtered = append(filtered, task)
	}

	return filtered
}

func printBubbleTasksTable(tasks []*models.Task) error {
	m := NewTaskListModel(tasks, listPlain)

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

func printTasksJSON(tasks []*models.Task) error {
	// In a real implementation, we would use json.Marshal
	fmt.Println("[")
	for i, task := range tasks {
		fmt.Printf(`  {"id": "%s", "title": "%s", "status": "%s", "platform": "%s"}`,
			task.ID, task.Title, task.Status, task.Platform)
		if i < len(tasks)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("]")

	return nil
}

func printTasksCSV(tasks []*models.Task) error {
	// Print header
	fmt.Println("ID,Platform,Status,Priority,Title")

	// Print tasks
	for _, task := range tasks {
		fmt.Printf("%s,%s,%s,%s,%s\n",
			task.ID,
			task.Platform,
			task.Status,
			task.Priority,
			task.Title)
	}

	return nil
}
