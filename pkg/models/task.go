package models

import (
	"time"
)

type Task struct {
	ID          string            `json:"id" yaml:"id"`
	Title       string            `json:"title" yaml:"title"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Status      TaskStatus        `json:"status" yaml:"status"`
	Priority    Priority          `json:"priority,omitempty" yaml:"priority,omitempty"`
	Assignee    *User             `json:"assignee,omitempty" yaml:"assignee,omitempty"`
	Platform    Platform          `json:"platform" yaml:"platform"`
	ProjectID   string            `json:"project_id,omitempty" yaml:"project_id,omitempty"`
	Labels      []string          `json:"labels,omitempty" yaml:"labels,omitempty"`
	CreatedAt   time.Time         `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" yaml:"updated_at"`
	DueDate     *time.Time        `json:"due_date,omitempty" yaml:"due_date,omitempty"`
	Metadata    map[string]any    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

type TaskStatus string

const (
	StatusOpen       TaskStatus = "open"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
	StatusCancelled  TaskStatus = "cancelled"
)

func (ts TaskStatus) String() string {
	return string(ts)
}

func (ts TaskStatus) IsValid() bool {
	switch ts {
	case StatusOpen, StatusInProgress, StatusDone, StatusCancelled:
		return true
	default:
		return false
	}
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) String() string {
	return string(p)
}

func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

type Platform string

const (
	PlatformLinear Platform = "linear"
	PlatformJira   Platform = "jira"
	PlatformSlack  Platform = "slack"
	PlatformGitHub Platform = "github"
)

func (p Platform) String() string {
	return string(p)
}

func (p Platform) IsValid() bool {
	switch p {
	case PlatformLinear, PlatformJira, PlatformSlack, PlatformGitHub:
		return true
	default:
		return false
	}
}

type TaskFilter struct {
	Platform  *Platform   `json:"platform,omitempty"`
	Status    *TaskStatus `json:"status,omitempty"`
	Priority  *Priority   `json:"priority,omitempty"`
	Assignee  string      `json:"assignee,omitempty"`
	ProjectID string      `json:"project_id,omitempty"`
	Labels    []string    `json:"labels,omitempty"`
	Query     string      `json:"query,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

func NewTask(title string, platform Platform) *Task {
	now := time.Now()
	return &Task{
		Title:     title,
		Platform:  platform,
		Status:    StatusOpen,
		Priority:  PriorityMedium,
		CreatedAt: now,
		UpdatedAt: now,
		Metadata:  make(map[string]any),
	}
}

func (t *Task) SetStatus(status TaskStatus) {
	if status.IsValid() {
		t.Status = status
		t.UpdatedAt = time.Now()
	}
}

func (t *Task) SetPriority(priority Priority) {
	if priority.IsValid() {
		t.Priority = priority
		t.UpdatedAt = time.Now()
	}
}

func (t *Task) SetAssignee(user *User) {
	t.Assignee = user
	t.UpdatedAt = time.Now()
}

func (t *Task) AddLabel(label string) {
	if t.Labels == nil {
		t.Labels = []string{}
	}
	
	for _, existing := range t.Labels {
		if existing == label {
			return
		}
	}
	
	t.Labels = append(t.Labels, label)
	t.UpdatedAt = time.Now()
}

func (t *Task) RemoveLabel(label string) {
	if t.Labels == nil {
		return
	}
	
	for i, existing := range t.Labels {
		if existing == label {
			t.Labels = append(t.Labels[:i], t.Labels[i+1:]...)
			t.UpdatedAt = time.Now()
			return
		}
	}
}

func (t *Task) SetMetadata(key string, value any) {
	if t.Metadata == nil {
		t.Metadata = make(map[string]any)
	}
	t.Metadata[key] = value
	t.UpdatedAt = time.Now()
}

func (t *Task) GetMetadata(key string) (any, bool) {
	if t.Metadata == nil {
		return nil, false
	}
	value, exists := t.Metadata[key]
	return value, exists
}