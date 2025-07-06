package linear

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hasura/go-graphql-client"
	"opentask/pkg/models"
	"opentask/pkg/platforms"
)

const (
	LinearAPIURL = "https://api.linear.app/graphql"
)

type Client struct {
	graphql *graphql.Client
	token   string
	baseURL string
}

type Config struct {
	Token   string `json:"token" yaml:"token"`
	BaseURL string `json:"base_url,omitempty" yaml:"base_url,omitempty"`
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.Token == "" {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidConfig,
			"linear",
			"",
			fmt.Errorf("API token is required"),
		)
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = LinearAPIURL
	}

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &authTransport{
			token: cfg.Token,
			base:  http.DefaultTransport,
		},
	}

	graphqlClient := graphql.NewClient(baseURL, httpClient)

	return &Client{
		graphql: graphqlClient,
		token:   cfg.Token,
		baseURL: baseURL,
	}, nil
}

// authTransport adds Authorization header to requests
type authTransport struct {
	token string
	base  http.RoundTripper
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.token))
	req.Header.Set("Content-Type", "application/json")
	return t.base.RoundTrip(req)
}

// Implement PlatformClient interface
func (c *Client) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	var mutation struct {
		IssueCreate struct {
			Success bool `graphql:"success"`
			Issue   struct {
				LinearIssue
			} `graphql:"issue"`
		} `graphql:"issueCreate(input: $input)"`
	}

	input := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"priority":    convertToLinearPriority(task.Priority),
	}

	// Add team ID if specified in metadata
	if teamID, ok := task.GetMetadata("team_id"); ok {
		input["teamId"] = teamID
	}

	// Add assignee if specified
	if task.Assignee != nil {
		if assigneeID, ok := task.Assignee.GetMetadata("linear_id"); ok {
			input["assigneeId"] = assigneeID
		}
	}

	// Add project if specified
	if task.ProjectID != "" {
		input["projectId"] = task.ProjectID
	}

	variables := map[string]interface{}{
		"input": input,
	}

	err := c.graphql.Mutate(ctx, &mutation, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to create issue: %w", err),
		)
	}

	if !mutation.IssueCreate.Success {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("issue creation failed"),
		)
	}

	createdTask := mutation.IssueCreate.Issue.LinearIssue.ToTask()
	return createdTask, nil
}

func (c *Client) GetTask(ctx context.Context, id string) (*models.Task, error) {
	var query struct {
		Issue LinearIssue `graphql:"issue(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": id,
	}

	err := c.graphql.Query(ctx, &query, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			id,
			fmt.Errorf("failed to get issue: %w", err),
		)
	}

	task := query.Issue.ToTask()
	return task, nil
}

func (c *Client) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	linearID, ok := task.GetMetadata("linear_id")
	if !ok {
		return nil, platforms.NewPlatformError(
			platforms.ErrInvalidInput,
			"linear",
			task.ID,
			fmt.Errorf("linear_id not found in task metadata"),
		)
	}

	var mutation struct {
		IssueUpdate struct {
			Success bool `graphql:"success"`
			Issue   struct {
				LinearIssue
			} `graphql:"issue"`
		} `graphql:"issueUpdate(id: $id, input: $input)"`
	}

	input := map[string]interface{}{
		"title":       task.Title,
		"description": task.Description,
		"priority":    convertToLinearPriority(task.Priority),
	}

	variables := map[string]interface{}{
		"id":    linearID,
		"input": input,
	}

	err := c.graphql.Mutate(ctx, &mutation, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			task.ID,
			fmt.Errorf("failed to update issue: %w", err),
		)
	}

	if !mutation.IssueUpdate.Success {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			task.ID,
			fmt.Errorf("issue update failed"),
		)
	}

	updatedTask := mutation.IssueUpdate.Issue.LinearIssue.ToTask()
	return updatedTask, nil
}

func (c *Client) DeleteTask(ctx context.Context, id string) error {
	var mutation struct {
		IssueDelete struct {
			Success bool `graphql:"success"`
		} `graphql:"issueDelete(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": id,
	}

	err := c.graphql.Mutate(ctx, &mutation, variables)
	if err != nil {
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			id,
			fmt.Errorf("failed to delete issue: %w", err),
		)
	}

	if !mutation.IssueDelete.Success {
		return platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			id,
			fmt.Errorf("issue deletion failed"),
		)
	}

	return nil
}

func (c *Client) ListTasks(ctx context.Context, filter *models.TaskFilter) ([]*models.Task, error) {
	var query struct {
		Issues struct {
			Nodes []LinearIssue `graphql:"nodes"`
		} `graphql:"issues(first: $first, after: $after, filter: $filter)"`
	}

	first := 50
	if filter != nil && filter.Limit > 0 {
		first = filter.Limit
	}

	var after *string
	if filter != nil && filter.Offset > 0 {
		// In GraphQL pagination, we'd need to convert offset to cursor
		// For simplicity, we'll skip this for now
	}

	linearFilter := map[string]interface{}{}
	if filter != nil {
		if filter.Status != nil {
			linearFilter["state"] = map[string]interface{}{
				"type": map[string]interface{}{
					"eq": convertToLinearStateType(*filter.Status),
				},
			}
		}
		if filter.Assignee != "" {
			linearFilter["assignee"] = map[string]interface{}{
				"email": map[string]interface{}{
					"eq": filter.Assignee,
				},
			}
		}
	}

	variables := map[string]interface{}{
		"first":  first,
		"after":  after,
		"filter": linearFilter,
	}

	err := c.graphql.Query(ctx, &query, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to list issues: %w", err),
		)
	}

	var tasks []*models.Task
	for _, issue := range query.Issues.Nodes {
		tasks = append(tasks, issue.ToTask())
	}

	return tasks, nil
}

func (c *Client) ListProjects(ctx context.Context) ([]*models.Project, error) {
	var query struct {
		Projects struct {
			Nodes []LinearProject `graphql:"nodes"`
		} `graphql:"projects(first: 100)"`
	}

	err := c.graphql.Query(ctx, &query, nil)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to list projects: %w", err),
		)
	}

	var projects []*models.Project
	for _, project := range query.Projects.Nodes {
		projects = append(projects, project.ToProject())
	}

	return projects, nil
}

func (c *Client) GetProject(ctx context.Context, id string) (*models.Project, error) {
	var query struct {
		Project LinearProject `graphql:"project(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": id,
	}

	err := c.graphql.Query(ctx, &query, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to get project: %w", err),
		)
	}

	project := query.Project.ToProject()
	return project, nil
}

func (c *Client) GetCurrentUser(ctx context.Context) (*models.User, error) {
	var query struct {
		Viewer LinearUser `graphql:"viewer"`
	}

	err := c.graphql.Query(ctx, &query, nil)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to get current user: %w", err),
		)
	}

	user := query.Viewer.ToUser()
	return user, nil
}

func (c *Client) SearchUsers(ctx context.Context, query string) ([]*models.User, error) {
	var gqlQuery struct {
		Users struct {
			Nodes []LinearUser `graphql:"nodes"`
		} `graphql:"users(first: 20, filter: $filter)"`
	}

	filter := map[string]interface{}{
		"name": map[string]interface{}{
			"contains": query,
		},
	}

	variables := map[string]interface{}{
		"filter": filter,
	}

	err := c.graphql.Query(ctx, &gqlQuery, variables)
	if err != nil {
		return nil, platforms.NewPlatformError(
			platforms.ErrPlatformAPI,
			"linear",
			"",
			fmt.Errorf("failed to search users: %w", err),
		)
	}

	var users []*models.User
	for _, user := range gqlQuery.Users.Nodes {
		users = append(users, user.ToUser())
	}

	return users, nil
}

func (c *Client) GetPlatformInfo() platforms.PlatformInfo {
	return platforms.PlatformInfo{
		Name:        "Linear",
		Type:        "linear",
		Version:     "1.0",
		Description: "Linear issue tracking and project management",
		BaseURL:     c.baseURL,
	}
}

func (c *Client) HealthCheck(ctx context.Context) error {
	_, err := c.GetCurrentUser(ctx)
	return err
}

// Helper function to convert task status to Linear state type
func convertToLinearStateType(status models.TaskStatus) string {
	switch status {
	case models.StatusOpen:
		return "unstarted"
	case models.StatusInProgress:
		return "started"
	case models.StatusDone:
		return "completed"
	case models.StatusCancelled:
		return "canceled"
	default:
		return "unstarted"
	}
}
