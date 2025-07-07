package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"opentask/pkg/models"
	"opentask/pkg/platforms"

	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock server responses
var mockJiraIssue = jira.Issue{
	ID:  "12345",
	Key: "TEST-123",
	Fields: &jira.IssueFields{
		Summary:     "Test Issue",
		Description: "Test Description",
		Status: &jira.Status{
			Name: "To Do",
			StatusCategory: jira.StatusCategory{
				Key: "new",
			},
		},
		Priority: &jira.Priority{
			Name: "Medium",
		},
		Assignee: &jira.User{
			AccountID:    "user123",
			DisplayName:  "John Doe",
			EmailAddress: "john@example.com",
		},
		Created: jira.Time(time.Now()),
		Updated: jira.Time(time.Now()),
		Labels:  []string{"test", "bug"},
		Project: jira.Project{
			ID:   "proj1",
			Key:  "TEST",
			Name: "Test Project",
		},
	},
}

var mockJiraProject = jira.Project{
	ID:   "proj1",
	Key:  "TEST",
	Name: "Test Project",
	Self: "https://example.atlassian.net/rest/api/2/project/proj1",
}

var mockJiraUser = jira.User{
	AccountID:    "user123",
	DisplayName:  "John Doe",
	EmailAddress: "john@example.com",
	Active:       true,
	Self:         "https://example.atlassian.net/rest/api/2/user/user123",
	AvatarUrls: jira.AvatarUrls{
		Four8X48: "https://avatar.example.com/48x48.png",
	},
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name: "valid config",
			config: Config{
				BaseURL: "https://example.atlassian.net",
				Email:   "test@example.com",
				Token:   "token123",
			},
			expectError: false,
		},
		{
			name: "missing base URL",
			config: Config{
				Email: "test@example.com",
				Token: "token123",
			},
			expectError: true,
			errorCode:   platforms.ErrInvalidConfig,
		},
		{
			name: "missing email",
			config: Config{
				BaseURL: "https://example.atlassian.net",
				Token:   "token123",
			},
			expectError: true,
			errorCode:   platforms.ErrInvalidConfig,
		},
		{
			name: "missing token",
			config: Config{
				BaseURL: "https://example.atlassian.net",
				Email:   "test@example.com",
			},
			expectError: true,
			errorCode:   platforms.ErrInvalidConfig,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, client)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
				assert.Equal(t, "jira", platErr.Platform)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
				assert.Equal(t, tt.config.BaseURL, client.baseURL)
				assert.Equal(t, tt.config.Email, client.email)
			}
		})
	}
}

func TestClient_CreateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/issue":
			if r.Method == "POST" {
				response := mockJiraIssue
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		task        *models.Task
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name: "valid task",
			task: &models.Task{
				Title:       "Test Task",
				Description: "Test Description",
				ProjectID:   "TEST",
				Priority:    models.PriorityMedium,
				Labels:      []string{"test"},
				Assignee: &models.User{
					ID:       "user123",
					Metadata: map[string]any{"jira_account_id": "user123"},
				},
			},
			expectError: false,
		},
		{
			name: "missing project ID",
			task: &models.Task{
				Title:       "Test Task",
				Description: "Test Description",
				Priority:    models.PriorityMedium,
			},
			expectError: true,
			errorCode:   platforms.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdTask, err := client.CreateTask(context.Background(), tt.task)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, createdTask)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, createdTask)
				assert.Equal(t, "TEST-123", createdTask.ID)
				assert.Equal(t, "Test Issue", createdTask.Title)
				assert.Equal(t, models.PlatformJira, createdTask.Platform)
			}
		})
	}
}

func TestClient_GetTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/issue/TEST-123":
			if r.Method == "GET" {
				response := mockJiraIssue
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		case "/rest/api/2/issue/NOT-FOUND":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name:        "existing task",
			taskID:      "TEST-123",
			expectError: false,
		},
		{
			name:        "non-existing task",
			taskID:      "NOT-FOUND",
			expectError: true,
			errorCode:   platforms.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := client.GetTask(context.Background(), tt.taskID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, task)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
				assert.Equal(t, tt.taskID, platErr.TaskID)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, task)
				assert.Equal(t, "TEST-123", task.ID)
				assert.Equal(t, "Test Issue", task.Title)
				assert.Equal(t, models.PlatformJira, task.Platform)
			}
		})
	}
}

func TestClient_UpdateTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/issue/TEST-123":
			if r.Method == "PUT" {
				w.WriteHeader(http.StatusNoContent)
			} else if r.Method == "GET" {
				response := mockJiraIssue
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		task        *models.Task
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name: "valid update",
			task: &models.Task{
				ID:          "TEST-123",
				Title:       "Updated Task",
				Description: "Updated Description",
				Priority:    models.PriorityHigh,
				Labels:      []string{"updated"},
				Metadata:    map[string]any{"jira_id": "TEST-123"},
			},
			expectError: false,
		},
		{
			name: "invalid jira_id type",
			task: &models.Task{
				ID:       "TEST-123",
				Title:    "Updated Task",
				Metadata: map[string]any{"jira_id": 123}, // Invalid type
			},
			expectError: true,
			errorCode:   platforms.ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedTask, err := client.UpdateTask(context.Background(), tt.task)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, updatedTask)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedTask)
				assert.Equal(t, "TEST-123", updatedTask.ID)
				assert.Equal(t, models.PlatformJira, updatedTask.Platform)
			}
		})
	}
}

func TestClient_DeleteTask(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/issue/TEST-123":
			if r.Method == "DELETE" {
				w.WriteHeader(http.StatusNoContent)
			}
		case "/rest/api/2/issue/NOT-FOUND":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		taskID      string
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name:        "existing task",
			taskID:      "TEST-123",
			expectError: false,
		},
		{
			name:        "non-existing task",
			taskID:      "NOT-FOUND",
			expectError: true,
			errorCode:   platforms.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.DeleteTask(context.Background(), tt.taskID)

			if tt.expectError {
				assert.Error(t, err)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
				assert.Equal(t, tt.taskID, platErr.TaskID)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClient_ListTasks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/search":
			if r.Method == "GET" {
				// The Jira search API returns a searchResult structure
				response := map[string]interface{}{
					"expand":     "schema,names",
					"startAt":    0,
					"maxResults": 50,
					"total":      1,
					"issues":     []jira.Issue{mockJiraIssue},
				}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		filter      *models.TaskFilter
		expectError bool
		expectCount int
	}{
		{
			name:        "no filter",
			filter:      nil,
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with status filter",
			filter: &models.TaskFilter{
				Status: func() *models.TaskStatus { s := models.StatusOpen; return &s }(),
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with assignee filter",
			filter: &models.TaskFilter{
				Assignee: "me",
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with project filter",
			filter: &models.TaskFilter{
				ProjectID: "TEST",
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with labels filter",
			filter: &models.TaskFilter{
				Labels: []string{"test"},
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with query filter",
			filter: &models.TaskFilter{
				Query: "test",
			},
			expectError: false,
			expectCount: 1,
		},
		{
			name: "with pagination",
			filter: &models.TaskFilter{
				Limit:  10,
				Offset: 0,
			},
			expectError: false,
			expectCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, err := client.ListTasks(context.Background(), tt.filter)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, tasks)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tasks)
				assert.Len(t, tasks, tt.expectCount)

				if len(tasks) > 0 {
					assert.Equal(t, "TEST-123", tasks[0].ID)
					assert.Equal(t, models.PlatformJira, tasks[0].Platform)
				}
			}
		})
	}
}

func TestClient_ListProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/project":
			if r.Method == "GET" {
				response := []jira.Project{mockJiraProject}
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	projects, err := client.ListProjects(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, projects)
	assert.Len(t, projects, 1)
	assert.Equal(t, "proj1", projects[0].ID)
	assert.Equal(t, "TEST", projects[0].Key)
	assert.Equal(t, "Test Project", projects[0].Name)
	assert.Equal(t, models.PlatformJira, projects[0].Platform)
}

func TestClient_GetProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/project/TEST":
			if r.Method == "GET" {
				response := mockJiraProject
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		case "/rest/api/2/project/NOT-FOUND":
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		projectID   string
		expectError bool
		errorCode   platforms.ErrorCode
	}{
		{
			name:        "existing project",
			projectID:   "TEST",
			expectError: false,
		},
		{
			name:        "non-existing project",
			projectID:   "NOT-FOUND",
			expectError: true,
			errorCode:   platforms.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			project, err := client.GetProject(context.Background(), tt.projectID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, project)

				var platErr *platforms.PlatformError
				assert.ErrorAs(t, err, &platErr)
				assert.Equal(t, tt.errorCode, platErr.Code)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, project)
				assert.Equal(t, "proj1", project.ID)
				assert.Equal(t, "TEST", project.Key)
				assert.Equal(t, models.PlatformJira, project.Platform)
			}
		})
	}
}

func TestClient_GetCurrentUser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/myself":
			if r.Method == "GET" {
				response := mockJiraUser
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	user, err := client.GetCurrentUser(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user123", user.ID)
	assert.Equal(t, "John Doe", user.Name)
	assert.Equal(t, "john@example.com", user.Email)
	assert.Equal(t, models.PlatformJira, user.Platform)
	assert.True(t, user.Active)
}

func TestClient_SearchUsers(t *testing.T) {
	config := Config{
		BaseURL: "https://example.atlassian.net",
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	users, err := client.SearchUsers(context.Background(), "john")

	assert.NoError(t, err)
	assert.NotNil(t, users)
	assert.Len(t, users, 0) // Current implementation returns empty slice
}

func TestClient_GetPlatformInfo(t *testing.T) {
	config := Config{
		BaseURL: "https://example.atlassian.net",
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	info := client.GetPlatformInfo()

	assert.Equal(t, "Jira", info.Name)
	assert.Equal(t, "jira", info.Type)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, "Atlassian Jira issue tracking and project management", info.Description)
	assert.Equal(t, "https://example.atlassian.net", info.BaseURL)
}

func TestClient_HealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/myself":
			if r.Method == "GET" {
				response := mockJiraUser
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := Config{
		BaseURL: server.URL,
		Email:   "test@example.com",
		Token:   "token123",
	}

	client, err := NewClient(config)
	require.NoError(t, err)

	err = client.HealthCheck(context.Background())
	assert.NoError(t, err)
}

func TestBuildJQLQuery(t *testing.T) {
	tests := []struct {
		name     string
		filter   *models.TaskFilter
		expected string
	}{
		{
			name:     "nil filter",
			filter:   nil,
			expected: "ORDER BY created DESC",
		},
		{
			name:     "empty filter",
			filter:   &models.TaskFilter{},
			expected: "ORDER BY created DESC",
		},
		{
			name: "status filter",
			filter: &models.TaskFilter{
				Status: func() *models.TaskStatus { s := models.StatusOpen; return &s }(),
			},
			expected: `status = "To Do" ORDER BY created DESC`,
		},
		{
			name: "assignee filter - me",
			filter: &models.TaskFilter{
				Assignee: "me",
			},
			expected: "assignee = currentUser() ORDER BY created DESC",
		},
		{
			name: "assignee filter - specific user",
			filter: &models.TaskFilter{
				Assignee: "john.doe",
			},
			expected: `assignee = "john.doe" ORDER BY created DESC`,
		},
		{
			name: "project filter",
			filter: &models.TaskFilter{
				ProjectID: "TEST",
			},
			expected: `project = "TEST" ORDER BY created DESC`,
		},
		{
			name: "labels filter",
			filter: &models.TaskFilter{
				Labels: []string{"bug", "urgent"},
			},
			expected: `(labels = "bug" AND labels = "urgent") ORDER BY created DESC`,
		},
		{
			name: "query filter",
			filter: &models.TaskFilter{
				Query: "test search",
			},
			expected: `text ~ "test search" ORDER BY created DESC`,
		},
		{
			name: "combined filters",
			filter: &models.TaskFilter{
				Status:    func() *models.TaskStatus { s := models.StatusInProgress; return &s }(),
				Assignee:  "me",
				ProjectID: "TEST",
				Labels:    []string{"bug"},
				Query:     "urgent",
			},
			expected: `status = "In Progress" AND assignee = currentUser() AND project = "TEST" AND (labels = "bug") AND text ~ "urgent" ORDER BY created DESC`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildJQLQuery(tt.filter)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkNewClient(b *testing.B) {
	config := Config{
		BaseURL: "https://example.atlassian.net",
		Email:   "test@example.com",
		Token:   "token123",
	}

	for i := 0; i < b.N; i++ {
		client, err := NewClient(config)
		if err != nil {
			b.Fatal(err)
		}
		_ = client
	}
}

func BenchmarkBuildJQLQuery(b *testing.B) {
	filter := &models.TaskFilter{
		Status:    func() *models.TaskStatus { s := models.StatusInProgress; return &s }(),
		Assignee:  "me",
		ProjectID: "TEST",
		Labels:    []string{"bug", "urgent"},
		Query:     "test search",
	}

	for i := 0; i < b.N; i++ {
		_ = buildJQLQuery(filter)
	}
}

// Test helper functions
func createTestTask() *models.Task {
	return &models.Task{
		Title:       "Test Task",
		Description: "Test Description",
		Status:      models.StatusOpen,
		Priority:    models.PriorityMedium,
		Platform:    models.PlatformJira,
		ProjectID:   "TEST",
		Labels:      []string{"test"},
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Metadata:    make(map[string]any),
	}
}

func createTestUser() *models.User {
	return &models.User{
		ID:       "user123",
		Name:     "John Doe",
		Email:    "john@example.com",
		Platform: models.PlatformJira,
		Active:   true,
		Metadata: map[string]any{
			"jira_account_id": "user123",
		},
	}
}

func createTestProject() *models.Project {
	return &models.Project{
		ID:       "proj1",
		Name:     "Test Project",
		Key:      "TEST",
		Platform: models.PlatformJira,
		Active:   true,
		Metadata: map[string]any{
			"jira_id": "proj1",
		},
	}
}

// Integration test helper - only runs if JIRA_INTEGRATION_TEST env var is set
func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// This would require actual Jira credentials and should only run in CI/CD
	// or when explicitly enabled for integration testing
	t.Skip("Integration test requires real Jira instance")
}
