package project

import (
	"context"
	"fmt"
	"os"
	"time"

	"opentask/pkg/config"
	"opentask/pkg/models"
	"opentask/pkg/platforms"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List projects",
	Long: `List projects from configured platforms.
	
You can filter projects by platform or show projects from all enabled platforms.`,
	RunE: runProjectList,
}

var (
	listPlatform string
	listFormat   string
	listPlain    bool
)

func init() {
	listCmd.Flags().StringVarP(&listPlatform, "platform", "p", "", "filter by platform")
	listCmd.Flags().StringVarP(&listFormat, "format", "f", "table", "output format (table, json, csv)")
	listCmd.Flags().BoolVar(&listPlain, "plain", false, "disable interactive mode and output plain text")
}

func runProjectList(cmd *cobra.Command, args []string) error {
	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg := manager.GetConfig()

	platforms := determinePlatformsForProjectList(cfg)
	if len(platforms) == 0 {
		return fmt.Errorf("no platforms configured or enabled")
	}

	var allProjects []*models.Project

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

		// Fetch projects from platform
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		projects, err := client.ListProjects(ctx)
		if err != nil {
			fmt.Printf("⚠ Failed to list projects from %s: %v\n", platformName, err)
			continue
		}

		allProjects = append(allProjects, projects...)
	}

	if len(allProjects) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	switch listFormat {
	case "json":
		return printProjectsJSON(allProjects)
	case "csv":
		return printProjectsCSV(allProjects)
	default:
		return printProjectsTable(allProjects)
	}
}

func determinePlatformsForProjectList(cfg *config.Config) []string {
	if listPlatform != "" {
		return []string{listPlatform}
	}

	// Default to enabled platforms
	return cfg.GetEnabledPlatforms()
}

func printProjectsTable(projects []*models.Project) error {
	if listPlain {
		return printProjectsPlainTable(projects)
	}

	m := NewProjectListModel(projects)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	return nil
}

func printProjectsPlainTable(projects []*models.Project) error {
	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Headers("ID", "KEY", "NAME", "PLATFORM", "ACTIVE")

	for _, project := range projects {
		activeStr := "✓"
		if !project.Active {
			activeStr = "✗"
		}

		t.Row(
			project.ID,
			project.Key,
			project.Name,
			string(project.Platform),
			activeStr,
		)
	}

	fmt.Println(t)
	return nil
}

func printProjectsJSON(projects []*models.Project) error {
	fmt.Println("[")
	for i, project := range projects {
		fmt.Printf(`  {"id": "%s", "key": "%s", "name": "%s", "platform": "%s", "active": %t}`,
			project.ID, project.Key, project.Name, project.Platform, project.Active)
		if i < len(projects)-1 {
			fmt.Println(",")
		} else {
			fmt.Println()
		}
	}
	fmt.Println("]")
	return nil
}

func printProjectsCSV(projects []*models.Project) error {
	// Print header
	fmt.Println("ID,Key,Name,Platform,Active")

	// Print projects
	for _, project := range projects {
		fmt.Printf("%s,%s,%s,%s,%t\n",
			project.ID,
			project.Key,
			project.Name,
			project.Platform,
			project.Active)
	}

	return nil
}

// ProjectListModel for bubble tea interactive display
type ProjectListModel struct {
	projects []*models.Project
	cursor   int
	selected bool
}

func NewProjectListModel(projects []*models.Project) ProjectListModel {
	return ProjectListModel{
		projects: projects,
		cursor:   0,
		selected: false,
	}
}

func (m ProjectListModel) Init() tea.Cmd {
	return nil
}

func (m ProjectListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.projects)-1 {
				m.cursor++
			}
		case "enter", " ":
			m.selected = !m.selected
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m ProjectListModel) View() string {
	s := "Projects:\n\n"

	// Create table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		Headers("ID", "KEY", "NAME", "PLATFORM", "ACTIVE")

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Background(lipgloss.Color("57")).
		Bold(true)

	for i, project := range m.projects {
		activeStr := "✓"
		if !project.Active {
			activeStr = "✗"
		}

		row := []string{
			project.ID,
			project.Key,
			project.Name,
			string(project.Platform),
			activeStr,
		}

		if i == m.cursor {
			// Apply selected style to each cell
			for j, cell := range row {
				row[j] = selectedStyle.Render(cell)
			}
		}

		t.Row(row...)
	}

	s += t.String()
	s += "\n\nUse arrow keys to navigate, Enter to select, q to quit\n"

	return s
}

// Helper function to create platform client (copied from task package)
func createPlatformClient(platformName string, platform config.Platform) (platforms.PlatformClient, error) {
	// Prepare configuration for platform factory
	clientConfig := make(map[string]any)

	// Copy credentials
	for key, value := range platform.Credentials {
		clientConfig[key] = value
	}

	// Copy settings
	for key, value := range platform.Settings {
		clientConfig[key] = value
	}

	// Create client using registry
	client, err := platforms.DefaultRegistry.Create(platform.Type, clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", platformName, err)
	}

	return client, nil
}
