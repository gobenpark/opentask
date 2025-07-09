package task

import (
	"context"
	"fmt"
	"opentask/pkg/config"
	"opentask/pkg/models"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var detailStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.RoundedBorder()).
	BorderForeground(lipgloss.Color("62")).
	Padding(1, 2)

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewDeleteConfirm
)

type model struct {
	table         table.Model
	viewport      viewport.Model
	plain         bool
	tasks         []*models.Task
	currentView   viewState
	selectedTask  *models.Task
	config        *config.Config
	deleteTask    *models.Task
	deleteMessage string
}

func (m model) Init() tea.Cmd {
	if m.plain {
		// In plain mode, immediately quit after initial render
		return tea.Quit
	}
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if m.currentView == viewDetail {
			m.viewport.Width = msg.Width - 4
			m.viewport.Height = msg.Height - 6
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.currentView == viewDetail {
				m.currentView = viewList
				return m, nil
			}
			if m.currentView == viewDeleteConfirm {
				m.currentView = viewList
				m.deleteTask = nil
				m.deleteMessage = ""
				return m, nil
			}
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.currentView == viewList {
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					taskID := selectedRow[0]
					for _, task := range m.tasks {
						if task.ID == taskID {
							m.selectedTask = task
							m.currentView = viewDetail
							m.viewport.SetContent(m.formatTaskDetail())
							return m, nil
						}
					}
				}
			}
		case "d":
			if m.currentView == viewList {
				selectedRow := m.table.SelectedRow()
				if len(selectedRow) > 0 {
					taskID := selectedRow[0]
					for _, task := range m.tasks {
						if task.ID == taskID {
							m.deleteTask = task
							m.currentView = viewDeleteConfirm
							return m, nil
						}
					}
				}
			} else if m.currentView == viewDetail && m.selectedTask != nil {
				m.deleteTask = m.selectedTask
				m.currentView = viewDeleteConfirm
				return m, nil
			}
		case "y":
			if m.currentView == viewDeleteConfirm && m.deleteTask != nil {
				return m.confirmDelete()
			}
		case "n":
			if m.currentView == viewDeleteConfirm {
				m.currentView = viewList
				m.deleteTask = nil
				m.deleteMessage = ""
				return m, nil
			}
		case "1":
			if m.currentView == viewList {
				return m.updateSelectedTaskStatus("open")
			} else if m.currentView == viewDetail && m.selectedTask != nil {
				return m.updateTaskStatus("open")
			}
		case "2":
			if m.currentView == viewList {
				return m.updateSelectedTaskStatus("in_progress")
			} else if m.currentView == viewDetail && m.selectedTask != nil {
				return m.updateTaskStatus("in_progress")
			}
		case "3":
			if m.currentView == viewList {
				return m.updateSelectedTaskStatus("done")
			} else if m.currentView == viewDetail && m.selectedTask != nil {
				return m.updateTaskStatus("done")
			}
		case "4":
			if m.currentView == viewList {
				return m.updateSelectedTaskStatus("cancelled")
			} else if m.currentView == viewDetail && m.selectedTask != nil {
				return m.updateTaskStatus("cancelled")
			}
		case "r":
			if m.currentView == viewList {
				return m.refreshTasks()
			}
			return m, nil
		}
	}

	if m.currentView == viewList {
		m.table, cmd = m.table.Update(msg)
	} else if m.currentView == viewDetail {
		m.viewport, cmd = m.viewport.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	if m.plain {
		// In plain mode, return just the table content without styling
		return m.table.View()
	}

	switch m.currentView {
	case viewDetail:
		return m.renderTaskDetail()
	case viewDeleteConfirm:
		return m.renderDeleteConfirm()
	default:
		return baseStyle.Render(m.table.View()) + "\n" + "Enter: details • d:delete • 1:open 2:in_progress 3:done 4:cancelled • r:refresh • q:quit"
	}
}

func (m model) formatTaskDetail() string {
	if m.selectedTask == nil {
		return "No task selected"
	}

	task := m.selectedTask

	var details strings.Builder
	details.WriteString(fmt.Sprintf("Task ID: %s\n", task.ID))
	details.WriteString(fmt.Sprintf("Platform: %s\n", task.Platform))
	details.WriteString(fmt.Sprintf("Title: %s\n", task.Title))
	details.WriteString(fmt.Sprintf("Status: %s\n", task.Status))
	details.WriteString(fmt.Sprintf("Priority: %s\n", task.Priority))

	if task.Assignee != nil {
		details.WriteString(fmt.Sprintf("Assignee: %s\n", task.Assignee.Name))
	} else {
		details.WriteString("Assignee: None\n")
	}

	if task.ProjectID != "" {
		details.WriteString(fmt.Sprintf("Project: %s\n", task.ProjectID))
	}

	if len(task.Labels) > 0 {
		details.WriteString(fmt.Sprintf("Labels: %s\n", strings.Join(task.Labels, ", ")))
	}

	details.WriteString(fmt.Sprintf("Created: %s\n", task.CreatedAt.Format("2006-01-02 15:04:05")))
	details.WriteString(fmt.Sprintf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05")))

	if task.DueDate != nil {
		details.WriteString(fmt.Sprintf("Due Date: %s\n", task.DueDate.Format("2006-01-02 15:04:05")))
	}

	if task.Description != "" {
		details.WriteString(fmt.Sprintf("\nDescription:\n%s\n", task.Description))
	}

	if len(task.Metadata) > 0 {
		details.WriteString("\nMetadata:\n")
		for key, value := range task.Metadata {
			details.WriteString(fmt.Sprintf("  %s: %v\n", key, value))
		}
	}

	return details.String()
}

func (m model) renderTaskDetail() string {
	if m.selectedTask == nil {
		return "No task selected"
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		MarginBottom(1).
		Render(fmt.Sprintf("Task Details: %s", m.selectedTask.Title))

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		MarginTop(1).
		Render("↑↓ scroll • d:delete • 1:open 2:in_progress 3:done 4:cancelled • ESC back • q quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		detailStyle.Render(m.viewport.View()),
		footer,
	)
}

func NewTaskListModel(tasks []*models.Task, plain bool, cfg *config.Config) model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "PLATFORM", Width: 10},
		{Title: "STATUS", Width: 12},
		{Title: "PRIORITY", Width: 10},
		{Title: "TITLE", Width: 50},
		{Title: "ASSIGNEE", Width: 10},
	}

	rows := make([]table.Row, len(tasks))
	for i, task := range tasks {
		assignee := "none"
		if task.Assignee != nil {
			assignee = task.Assignee.Name
		}
		rows[i] = table.Row{
			task.ID,
			task.Platform.String(),
			task.Status.String(),
			task.Priority.String(),
			task.Title,
			assignee,
		}
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	vp := viewport.New(100, 30)

	return model{
		table:       t,
		viewport:    vp,
		plain:       plain,
		tasks:       tasks,
		currentView: viewList,
		config:      cfg,
	}
}

func (m model) updateTaskStatus(statusStr string) (tea.Model, tea.Cmd) {
	if m.selectedTask == nil {
		return m, nil
	}

	status := models.TaskStatus(statusStr)
	if !status.IsValid() {
		return m, nil
	}

	// Load configuration to find platform client
	manager := config.NewManager()
	if err := manager.Load(""); err != nil {
		return m, nil
	}

	cfg := manager.GetConfig()
	platformName := string(m.selectedTask.Platform)
	platform, exists := cfg.GetPlatform(platformName)
	if !exists || !platform.Enabled {
		return m, nil
	}

	// Create platform client
	client, err := createPlatformClient(platformName, platform)
	if err != nil {
		return m, nil
	}

	// Update task status
	originalStatus := m.selectedTask.Status
	m.selectedTask.SetStatus(status)

	// Update task via API
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updatedTask, err := client.UpdateTask(ctx, m.selectedTask)
	if err != nil {
		// Revert status on error
		m.selectedTask.SetStatus(originalStatus)
		return m, nil
	}

	// Update the task in our local list
	for i, task := range m.tasks {
		if task.ID == updatedTask.ID {
			m.tasks[i] = updatedTask
			m.selectedTask = updatedTask
			break
		}
	}

	// Refresh the viewport content
	m.viewport.SetContent(m.formatTaskDetail())

	return m, nil
}

func (m model) updateSelectedTaskStatus(statusStr string) (tea.Model, tea.Cmd) {
	selectedRow := m.table.SelectedRow()
	if len(selectedRow) == 0 {
		return m, nil
	}

	taskID := selectedRow[0]
	var targetTask *models.Task

	for _, task := range m.tasks {
		if task.ID == taskID {
			targetTask = task
			break
		}
	}

	if targetTask == nil {
		return m, nil
	}

	status := models.TaskStatus(statusStr)
	if !status.IsValid() {
		return m, nil
	}

	platformName := string(targetTask.Platform)
	platform, exists := m.config.GetPlatform(platformName)
	if !exists || !platform.Enabled {
		return m, nil
	}

	// Create platform client
	client, err := createPlatformClient(platformName, platform)
	if err != nil {
		return m, nil
	}

	// Update task status
	originalStatus := targetTask.Status
	targetTask.SetStatus(status)

	// Update task via API
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	updatedTask, err := client.UpdateTask(ctx, targetTask)
	if err != nil {
		// Revert status on error
		targetTask.SetStatus(originalStatus)
		return m, nil
	}

	// Update the task in our local list
	for i, task := range m.tasks {
		if task.ID == updatedTask.ID {
			m.tasks[i] = updatedTask
			break
		}
	}

	// Refresh the table
	m = m.refreshTable()

	return m, nil
}

func (m model) refreshTasks() (tea.Model, tea.Cmd) {
	if m.config == nil {
		return m, nil
	}

	platforms := m.config.GetEnabledPlatforms()
	var allTasks []*models.Task

	for _, platformName := range platforms {
		platform, exists := m.config.GetPlatform(platformName)
		if !exists || !platform.Enabled {
			continue
		}

		client, err := createPlatformClient(platformName, platform)
		if err != nil {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Use a basic filter for refresh
		filter := &models.TaskFilter{
			Limit: 100, // Get more tasks for refresh
		}

		tasks, err := client.ListTasks(ctx, filter)
		if err != nil {
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	// Update model with new tasks
	m.tasks = allTasks
	m = m.refreshTable()

	return m, nil
}

func (m model) refreshTable() model {
	rows := make([]table.Row, len(m.tasks))
	for i, task := range m.tasks {
		assignee := "none"
		if task.Assignee != nil {
			assignee = task.Assignee.Name
		}
		rows[i] = table.Row{
			task.ID,
			task.Platform.String(),
			task.Status.String(),
			task.Priority.String(),
			task.Title,
			assignee,
		}
	}

	m.table.SetRows(rows)
	return m
}

func (m model) renderDeleteConfirm() string {
	if m.deleteTask == nil {
		return "No task selected for deletion"
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Padding(1, 2).
		MarginTop(5).
		MarginLeft(10)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("196"))

	content := fmt.Sprintf(
		"%s\n\n%s\n\n%s\n\n%s",
		titleStyle.Render("⚠ Delete Task"),
		fmt.Sprintf("Are you sure you want to delete this task?"),
		fmt.Sprintf("ID: %s\nTitle: %s", m.deleteTask.ID, m.deleteTask.Title),
		"Press 'y' to confirm, 'n' to cancel, or ESC to go back",
	)

	if m.deleteMessage != "" {
		content += fmt.Sprintf("\n\n%s", m.deleteMessage)
	}

	return style.Render(content)
}

func (m model) confirmDelete() (tea.Model, tea.Cmd) {
	if m.deleteTask == nil {
		return m, nil
	}

	// Find platform for the task
	platformName := string(m.deleteTask.Platform)
	platform, exists := m.config.GetPlatform(platformName)
	if !exists || !platform.Enabled {
		m.deleteMessage = "Platform not found or not enabled"
		return m, nil
	}

	// Create platform client
	client, err := createPlatformClient(platformName, platform)
	if err != nil {
		m.deleteMessage = fmt.Sprintf("Failed to create client: %v", err)
		return m, nil
	}

	// Delete task via API
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.DeleteTask(ctx, m.deleteTask.ID)
	if err != nil {
		m.deleteMessage = fmt.Sprintf("Failed to delete task: %v", err)
		return m, nil
	}

	// Remove task from local list
	for i, task := range m.tasks {
		if task.ID == m.deleteTask.ID {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			break
		}
	}

	// Clear selected task if it was the one being deleted
	if m.selectedTask != nil && m.selectedTask.ID == m.deleteTask.ID {
		m.selectedTask = nil
	}

	// Reset delete state
	m.deleteTask = nil
	m.deleteMessage = ""
	m.currentView = viewList

	// Refresh the table
	m = m.refreshTable()

	return m, nil
}
