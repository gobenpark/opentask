package jira

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"opentask/pkg/models"
	"opentask/pkg/platforms"

	"github.com/andygrunwald/go-jira"
)

type Client struct {
	client  *jira.Client
	baseURL string
	email   string
}

type Config struct {
	BaseURL string `json:"base_url" yaml:"base_url"`
	Email   string `json:"email" yaml:"email"`
	Token   string `json:"token" yaml:"token"`
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidConfig,
			"jira",
			"",
			fmt.Errorf("base URL is required"),
		)
	}

	if cfg.Email == "" || cfg.Token == "" {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidConfig,
			"jira",
			"",
			fmt.Errorf("email and token are required"),
		)
	}

	// Create basic auth transport
	tp := jira.BasicAuthTransport{
		Username: cfg.Email,
		Password: cfg.Token,
	}

	// Create Jira client
	jiraClient, err := jira.NewClient(tp.Client(), cfg.BaseURL)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidConfig,
			"jira",
			"",
			fmt.Errorf("failed to create Jira client: %w", err),
		)
	}

	return &Client{
		client:  jiraClient,
		baseURL: cfg.BaseURL,
		email:   cfg.Email,
	}, nil
}

// Implement PlatformClient interface
func (c *Client) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	// Create issue fields
	issueFields := &jira.IssueFields{
		Summary:     task.Title,
		Description: task.Description,
		Type: jira.IssueType{
			Name: "Task", // Default to Task type
		},
	}

	// Set project
	if task.ProjectID != "" {
		issueFields.Project = jira.Project{
			ID: task.ProjectID,
		}
	} else {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidInput,
			"jira",
			"",
			fmt.Errorf("project ID is required for Jira issues"),
		)
	}

	// Set priority
	if task.Priority != "" {
		issueFields.Priority = &jira.Priority{
			Name: convertToJiraPriority(task.Priority),
		}
	}

	// Set assignee
	if task.Assignee != nil {
		if accountID, ok := task.Assignee.GetMetadata("jira_account_id"); ok {
			if accountIDStr, ok := accountID.(string); ok {
				issueFields.Assignee = &jira.User{
					AccountID: accountIDStr,
				}
			}
		}
	}

	// Set labels
	if len(task.Labels) > 0 {
		issueFields.Labels = task.Labels
	}

	// Create the issue
	issue := &jira.Issue{
		Fields: issueFields,
	}

	createdIssue, resp, err := c.client.Issue.CreateWithContext(ctx, issue)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("failed to create issue: %w", err),
		)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("create issue failed with status %d", resp.StatusCode),
		)
	}

	// Convert created issue back to our task format
	jiraIssue := &JiraIssue{Issue: *createdIssue}
	createdTask := jiraIssue.ToTask()

	return createdTask, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (*models.Task, error) {
	issue, resp, err := c.client.Issue.Get(id, nil)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, platforms.NewPlatformError(
				platforms.ErrNotFound,
				"jira",
				id,
				fmt.Errorf("issue not found"),
			)
		}
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			id,
			fmt.Errorf("failed to get issue: %w", err),
		)
	}
	defer resp.Body.Close()

	jiraIssue := &JiraIssue{Issue: *issue}
	task := jiraIssue.ToTask()

	return task, nil
}

func (c *Client) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	jiraID, ok := task.GetMetadata("jira_id")
	if !ok {
		// If no jira_id, try using the task ID directly
		jiraID = task.ID
	}

	jiraIDStr, ok := jiraID.(string)
	if !ok {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidInput,
			"jira",
			task.ID,
			fmt.Errorf("invalid jira_id in task metadata"),
		)
	}

	// Get current issue to compare status
	currentIssue, _, err := c.client.Issue.Get(jiraIDStr, nil)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			task.ID,
			fmt.Errorf("failed to get current issue: %w", err),
		)
	}

	// Update status via transition if needed
	currentStatus := convertFromJiraStatus(currentIssue.Fields.Status.Name)
	if currentStatus != task.Status {
		err := c.transitionIssue(jiraIDStr, task.Status)
		if err != nil {
			return nil, err
		}
	}

	// Create update fields for other properties
	updateFields := &jira.IssueFields{
		Summary:     task.Title,
		Description: task.Description,
	}

	// Set priority
	if task.Priority != "" {
		updateFields.Priority = &jira.Priority{
			Name: convertToJiraPriority(task.Priority),
		}
	}

	// Set labels
	if len(task.Labels) > 0 {
		updateFields.Labels = task.Labels
	}

	// Update the issue fields
	issue := &jira.Issue{
		Key:    jiraIDStr,
		Fields: updateFields,
	}

	updatedIssue, resp, err := c.client.Issue.Update(issue)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			task.ID,
			fmt.Errorf("failed to update issue: %w", err),
		)
	}
	defer resp.Body.Close()

	// If update successful, get the updated issue
	if updatedIssue == nil {
		updatedIssue, _, err = c.client.Issue.Get(jiraIDStr, nil)
		if err != nil {
			return nil, platforms.NewPlatformError(
				platforms.ErrPlatformAPI,
				"jira",
				task.ID,
				fmt.Errorf("failed to get updated issue: %w", err),
			)
		}
	}

	jiraIssue := &JiraIssue{Issue: *updatedIssue}
	updatedTask := jiraIssue.ToTask()

	return updatedTask, nil
}

func (c *Client) DeleteTask(ctx context.Context, id string) error {
	resp, err := c.client.Issue.Delete(id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return platforms.NewPlatformError(
				platforms.ErrNotFound,
				"jira",
				id,
				fmt.Errorf("issue not found"),
			)
		}
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			id,
			fmt.Errorf("failed to delete issue: %w", err),
		)
	}
	defer resp.Body.Close()

	return nil
}

func (c *Client) ListTasks(ctx context.Context, filter *models.TaskFilter) ([]*models.Task, error) {
	// Build JQL query
	jql := buildJQLQuery(filter)

	// Set options
	options := &jira.SearchOptions{
		StartAt:    0,
		MaxResults: 50,
	}

	if filter != nil {
		if filter.Limit > 0 {
			options.MaxResults = filter.Limit
		}
		if filter.Offset > 0 {
			options.StartAt = filter.Offset
		}
	}

	// Search issues
	issues, resp, err := c.client.Issue.Search(jql, options)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("failed to search issues: %w", err),
		)
	}
	defer resp.Body.Close()

	// Convert to tasks
	var tasks []*models.Task
	for _, issue := range issues {
		jiraIssue := &JiraIssue{Issue: issue}
		tasks = append(tasks, jiraIssue.ToTask())
	}

	return tasks, nil
}

func (c *Client) ListProjects(ctx context.Context) ([]*models.Project, error) {
	projects, resp, err := c.client.Project.GetList()
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("failed to list projects: %w", err),
		)
	}
	defer resp.Body.Close()

	var result []*models.Project
	for _, project := range *projects {
		convertedProject := &models.Project{
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
		result = append(result, convertedProject)
	}

	return result, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*models.Project, error) {
	project, resp, err := c.client.Project.Get(id)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, platforms.NewPlatformError(
				platforms.ErrNotFound,
				"jira",
				"",
				fmt.Errorf("project not found"),
			)
		}
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("failed to get project: %w", err),
		)
	}
	defer resp.Body.Close()

	convertedProject := &models.Project{
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
	return convertedProject, nil
}

func (c *Client) GetCurrentUser(ctx context.Context) (*models.User, error) {
	user, resp, err := c.client.User.GetSelfWithContext(ctx)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			"",
			fmt.Errorf("failed to get current user: %w", err),
		)
	}
	defer resp.Body.Close()

	convertedUser := &models.User{
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
	return convertedUser, nil
}

func (c *Client) SearchUsers(ctx context.Context, query string) ([]*models.User, error) {
	// For now, return empty result as user search API varies by Jira version
	// In a real implementation, we'd use the appropriate API endpoint
	return []*models.User{}, nil
}

func (c *Client) GetPlatformInfo() platforms.PlatformInfo {
	return platforms.PlatformInfo{
		Name:        "Jira",
		Type:        "jira",
		Version:     "1.0",
		Description: "Atlassian Jira issue tracking and project management",
		BaseURL:     c.baseURL,
	}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.GetCurrentUser(ctx)
	return err
}

// convertFromJiraStatus converts Jira status name to our TaskStatus
func convertFromJiraStatus(jiraStatus string) models.TaskStatus {
	switch strings.ToLower(jiraStatus) {
	case "to do", "open", "new", "created":
		return models.StatusOpen
	case "in progress", "in development", "doing":
		return models.StatusInProgress
	case "done", "closed", "resolved", "completed":
		return models.StatusDone
	case "cancelled", "canceled", "rejected":
		return models.StatusCancelled
	default:
		return models.StatusOpen
	}
}

// transitionIssue transitions a Jira issue to the specified status
func (c *Client) transitionIssue(issueID string, targetStatus models.TaskStatus) error {
	// Get available transitions
	transitions, resp, err := c.client.Issue.GetTransitions(issueID)
	if err != nil {
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			issueID,
			fmt.Errorf("failed to get transitions: %w", err),
		)
	}
	defer resp.Body.Close()

	// Find the transition that leads to the target status
	targetJiraStatus := convertToJiraStatus(targetStatus)
	var targetTransition *jira.Transition

	for _, transition := range transitions {
		if transition.To.Name == targetJiraStatus {
			targetTransition = &transition
			break
		}
	}

	if targetTransition == nil {
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			issueID,
			fmt.Errorf("no transition available to status: %s", targetJiraStatus),
		)
	}

	// Perform the transition
	resp, err = c.client.Issue.DoTransition(issueID, targetTransition.ID)
	if err != nil {
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"jira",
			issueID,
			fmt.Errorf("failed to transition issue: %w", err),
		)
	}
	defer resp.Body.Close()

	return nil
}

// Helper function to build JQL query from filter
func buildJQLQuery(filter *models.TaskFilter) string {
	var conditions []string

	if filter == nil {
		return "ORDER BY created DESC"
	}

	// Add status filter
	if filter.Status != nil {
		statusName := convertToJiraStatus(*filter.Status)
		conditions = append(conditions, fmt.Sprintf("status = \"%s\"", statusName))
	}

	// Add assignee filter
	if filter.Assignee != "" {
		if filter.Assignee == "me" {
			conditions = append(conditions, "assignee = currentUser()")
		} else {
			conditions = append(conditions, fmt.Sprintf("assignee = \"%s\"", filter.Assignee))
		}
	}

	// Add project filter
	if filter.ProjectID != "" {
		conditions = append(conditions, fmt.Sprintf("project = \"%s\"", filter.ProjectID))
	}

	// Add labels filter
	if len(filter.Labels) > 0 {
		labelConditions := make([]string, len(filter.Labels))
		for i, label := range filter.Labels {
			labelConditions[i] = fmt.Sprintf("labels = \"%s\"", label)
		}
		conditions = append(conditions, "("+strings.Join(labelConditions, " AND ")+")")
	}

	// Add text search
	if filter.Query != "" {
		conditions = append(conditions, fmt.Sprintf("text ~ \"%s\"", filter.Query))
	}

	query := strings.Join(conditions, " AND ")
	if query == "" {
		query = "ORDER BY created DESC"
	} else {
		query += " ORDER BY created DESC"
	}

	return query
}
