package jira

import (
	"strings"
	"time"

	"opentask/pkg/models"

	"github.com/andygrunwald/go-jira"
)

// Extended Jira types for our needs
type JiraIssue struct {
	jira.Issue
}

type JiraProject jira.Project

type JiraUser jira.User

// Conversion methods to unified models
func (ji *JiraIssue) ToTask() *models.Task {
	task := &models.Task{
		ID:          ji.Key,
		Title:       ji.Fields.Summary,
		Description: ji.Fields.Description,
		Platform:    models.PlatformJira,
		ProjectID:   ji.Fields.Project.Key,
		CreatedAt:   time.Time(ji.Fields.Created),
		UpdatedAt:   time.Time(ji.Fields.Updated),
		Metadata:    make(map[string]any),
	}

	// Set status
	if ji.Fields.Status != nil {
		task.Status = convertJiraStatus(ji.Fields.Status.StatusCategory.Key)
	} else {
		task.Status = models.StatusOpen // Default status
	}

	// Set priority
	if ji.Fields.Priority != nil {
		task.Priority = convertJiraPriority(ji.Fields.Priority.Name)
	} else {
		task.Priority = models.PriorityMedium // Default priority
	}

	// Set assignee
	if ji.Fields.Assignee != nil {
		task.Assignee = &models.User{
			ID:       ji.Fields.Assignee.AccountID,
			Name:     ji.Fields.Assignee.DisplayName,
			Email:    ji.Fields.Assignee.EmailAddress,
			Platform: models.PlatformJira,
			Active:   ji.Fields.Assignee.Active,
			Metadata: map[string]any{
				"jira_account_id": ji.Fields.Assignee.AccountID,
			},
		}
	}

	// Set labels
	if ji.Fields.Labels != nil {
		task.Labels = ji.Fields.Labels
	}

	// Set due date (Jira Date type handling)
	dueDate := time.Time(ji.Fields.Duedate)
	if !dueDate.IsZero() {
		task.DueDate = &dueDate
	}

	// Set metadata
	task.Metadata["jira_id"] = ji.ID
	task.Metadata["jira_self"] = ji.Self
	if ji.Fields.Type.Name != "" {
		task.Metadata["issue_type"] = ji.Fields.Type.Name
	}
	if ji.Fields.Status != nil {
		task.Metadata["status_name"] = ji.Fields.Status.Name
		task.Metadata["status_category"] = ji.Fields.Status.StatusCategory.Key
	}
	if ji.Fields.Priority != nil {
		task.Metadata["priority_name"] = ji.Fields.Priority.Name
	}

	return task
}

func (jp *JiraProject) ToProject() *models.Project {
	project := jira.Project(*jp)
	return &models.Project{
		ID:       project.ID,
		Name:     project.Name,
		Key:      project.Key,
		Platform: models.PlatformJira,
		Active:   true,
		Metadata: map[string]any{
			"jira_id":   project.ID,
			"jira_self": project.Self,
		},
	}
}

func (ju *JiraUser) ToUser() *models.User {
	user := jira.User(*ju)
	return &models.User{
		ID:       user.AccountID,
		Name:     user.DisplayName,
		Email:    user.EmailAddress,
		Avatar:   user.AvatarUrls.Four8X48,
		Platform: models.PlatformJira,
		Active:   user.Active,
		Metadata: map[string]any{
			"jira_account_id": user.AccountID,
			"jira_self":       user.Self,
		},
	}
}

// Helper functions for status/priority conversion
func convertJiraStatus(statusCategory string) models.TaskStatus {
	switch strings.ToLower(statusCategory) {
	case "new", "to do", "todo":
		return models.StatusOpen
	case "indeterminate", "in progress":
		return models.StatusInProgress
	case "done", "complete":
		return models.StatusDone
	case "cancelled", "canceled":
		return models.StatusCancelled
	default:
		return models.StatusOpen
	}
}

func convertJiraPriority(priority string) models.Priority {
	switch strings.ToLower(priority) {
	case "highest", "critical", "blocker":
		return models.PriorityUrgent
	case "high", "major":
		return models.PriorityHigh
	case "medium", "normal":
		return models.PriorityMedium
	case "low", "minor", "trivial", "lowest":
		return models.PriorityLow
	default:
		return models.PriorityMedium
	}
}

func convertToJiraPriority(priority models.Priority) string {
	switch priority {
	case models.PriorityUrgent:
		return "Highest"
	case models.PriorityHigh:
		return "High"
	case models.PriorityMedium:
		return "Medium"
	case models.PriorityLow:
		return "Low"
	default:
		return "Medium"
	}
}

// Convert task status to Jira transition
func convertToJiraStatus(status models.TaskStatus) string {
	switch status {
	case models.StatusOpen:
		return "To Do"
	case models.StatusInProgress:
		return "In Progress"
	case models.StatusDone:
		return "Done"
	case models.StatusCancelled:
		return "Cancelled"
	default:
		return "To Do"
	}
}
