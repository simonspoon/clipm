package models

import "time"

// Task represents a task with its metadata and body content
type Task struct {
	ID          int64     `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Parent      *int64    `yaml:"parent"` // null if top-level
	Status      string    `yaml:"status"` // todo, in-progress, done, blocked
	Priority    string    `yaml:"priority"` // low, medium, high
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Tags        []string  `yaml:"tags"`
	Body        string    `yaml:"-"` // Markdown content after frontmatter
}

// Valid status values
const (
	StatusTodo       = "todo"
	StatusInProgress = "in-progress"
	StatusDone       = "done"
	StatusBlocked    = "blocked"
)

// Valid priority values
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
)

// IsValidStatus checks if a status value is valid
func IsValidStatus(status string) bool {
	return status == StatusTodo || status == StatusInProgress || status == StatusDone || status == StatusBlocked
}

// IsValidPriority checks if a priority value is valid
func IsValidPriority(priority string) bool {
	return priority == PriorityLow || priority == PriorityMedium || priority == PriorityHigh
}
