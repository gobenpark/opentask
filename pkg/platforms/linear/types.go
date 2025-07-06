package linear

import (
	"opentask/pkg/models"
	"time"
)

// Linear API response types
type LinearIssue struct {
	ID          string           `json:"id"`
	Identifier  string           `json:"identifier"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Priority    float64          `json:"priority"`
	State       LinearIssueState `json:"state"`
	Assignee    *LinearUser      `json:"assignee"`
	Team        LinearTeam       `json:"team"`
	Project     *LinearProject   `json:"project"`
	Labels      []LinearLabel    `json:"labels"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
	DueDate     *time.Time       `json:"dueDate"`
	URL         string           `json:"url"`
}

type LinearIssueState struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"`
	Color string `json:"color"`
}

type LinearUser struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	Active      bool   `json:"active"`
}

type LinearTeam struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Key         string `json:"key"`
	Description string `json:"description"`
}

type LinearProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SlugID      string `json:"slugId"`
}

type LinearLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type LinearWorkflowState struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// Conversion methods to unified models
func (li *LinearIssue) ToTask() *models.Task {
	task := &models.Task{
		ID:          li.Identifier,
		Title:       li.Title,
		Description: li.Description,
		Status:      convertLinearStatus(li.State.Type),
		Priority:    convertLinearPriority(li.Priority),
		Platform:    models.PlatformLinear,
		CreatedAt:   li.CreatedAt,
		UpdatedAt:   li.UpdatedAt,
		DueDate:     li.DueDate,
		Metadata:    make(map[string]any),
	}

	// Set assignee
	if li.Assignee != nil {
		task.Assignee = li.Assignee.ToUser()
	}

	// Set project
	if li.Project != nil {
		task.ProjectID = li.Project.ID
	}

	// Set labels
	for _, label := range li.Labels {
		task.Labels = append(task.Labels, label.Name)
	}

	// Set metadata
	task.Metadata["linear_id"] = li.ID
	task.Metadata["linear_url"] = li.URL
	task.Metadata["team"] = li.Team.Key
	task.Metadata["state_id"] = li.State.ID
	task.Metadata["state_color"] = li.State.Color

	return task
}

func (lu *LinearUser) ToUser() *models.User {
	return &models.User{
		ID:       lu.ID,
		Name:     lu.DisplayName,
		Email:    lu.Email,
		Avatar:   lu.AvatarURL,
		Platform: models.PlatformLinear,
		Active:   lu.Active,
		Metadata: map[string]any{
			"linear_id": lu.ID,
		},
	}
}

func (lp *LinearProject) ToProject() *models.Project {
	return &models.Project{
		ID:       lp.ID,
		Name:     lp.Name,
		Platform: models.PlatformLinear,
		Active:   true,
		Metadata: map[string]any{
			"linear_id": lp.ID,
			"slug_id":   lp.SlugID,
		},
	}
}

// Helper functions for status/priority conversion
func convertLinearStatus(stateType string) models.TaskStatus {
	switch stateType {
	case "unstarted", "backlog":
		return models.StatusOpen
	case "started":
		return models.StatusInProgress
	case "completed":
		return models.StatusDone
	case "canceled", "cancelled":
		return models.StatusCancelled
	default:
		return models.StatusOpen
	}
}

func convertLinearPriority(priority float64) models.Priority {
	// Linear priority: 0 = No priority, 1 = Urgent, 2 = High, 3 = Medium, 4 = Low
	switch {
	case priority == 1:
		return models.PriorityUrgent
	case priority == 2:
		return models.PriorityHigh
	case priority == 3:
		return models.PriorityMedium
	case priority == 4:
		return models.PriorityLow
	default:
		return models.PriorityMedium
	}
}

func convertToLinearPriority(priority models.Priority) float64 {
	switch priority {
	case models.PriorityUrgent:
		return 1
	case models.PriorityHigh:
		return 2
	case models.PriorityMedium:
		return 3
	case models.PriorityLow:
		return 4
	default:
		return 3
	}
}
