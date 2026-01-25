package models

import "time"

// Note represents an observation or progress update on a task
type Note struct {
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Task represents a task in the work queue
type Task struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Parent      *int64    `json:"parent"`
	Status      string    `json:"status"`
	BlockedBy   []int64   `json:"blockedBy,omitempty"`
	Owner       *string   `json:"owner,omitempty"`
	Notes       []Note    `json:"notes,omitempty"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
}

// Valid status values
const (
	StatusTodo       = "todo"
	StatusInProgress = "in-progress"
	StatusDone       = "done"
)

// IsValidStatus checks if a status value is valid
func IsValidStatus(status string) bool {
	return status == StatusTodo || status == StatusInProgress || status == StatusDone
}
