package models

import "time"

// Task represents a task in the work queue
type Task struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Parent      *int64    `json:"parent"`
	Status      string    `json:"status"`
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
