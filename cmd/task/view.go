package task

import (
	"fmt"
	"opentask/pkg/models"
	"strings"

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
)

type model struct {
	table        table.Model
	viewport     viewport.Model
	plain        bool
	tasks        []*models.Task
	currentView  viewState
	selectedTask *models.Task
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
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "r":
			return m, nil
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
	default:
		return baseStyle.Render(m.table.View()) + "\n" + "Press Enter to view task details • Press q to quit"
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
		Render("↑↓ scroll • ESC to go back • q to quit")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		detailStyle.Render(m.viewport.View()),
		footer,
	)
}

func NewTaskListModel(tasks []*models.Task, plain bool) model {
	columns := []table.Column{
		{Title: "ID", Width: 4},
		{Title: "PLATFORM", Width: 10},
		{Title: "STATUS", Width: 10},
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
		table.WithHeight(7),
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
	}
}
