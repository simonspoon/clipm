package models

import "time"

// IndexEntry represents a task entry in the index for fast lookups
type IndexEntry struct {
	ID       int64     `json:"id"`
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Priority string    `json:"priority"`
	Parent   *int64    `json:"parent"`
	Created  time.Time `json:"created"`
	Updated  time.Time `json:"updated"`
	Archived bool      `json:"archived"`
}

// Index represents the fast lookup index
type Index struct {
	Version string                 `json:"version"`
	Tasks   map[int64]*IndexEntry `json:"tasks"`
}

// NewIndex creates a new empty index
func NewIndex() *Index {
	return &Index{
		Version: "1.0.0",
		Tasks:   make(map[int64]*IndexEntry),
	}
}

// AddTask adds or updates a task in the index
func (idx *Index) AddTask(task *Task, archived bool) {
	idx.Tasks[task.ID] = &IndexEntry{
		ID:       task.ID,
		Name:     task.Name,
		Status:   task.Status,
		Priority: task.Priority,
		Parent:   task.Parent,
		Created:  task.Created,
		Updated:  task.Updated,
		Archived: archived,
	}
}

// RemoveTask removes a task from the index
func (idx *Index) RemoveTask(id int64) {
	delete(idx.Tasks, id)
}

// GetTask retrieves a task from the index
func (idx *Index) GetTask(id int64) (*IndexEntry, bool) {
	entry, exists := idx.Tasks[id]
	return entry, exists
}

// GetChildren returns all task IDs that have the given task as their parent
func (idx *Index) GetChildren(parentID int64) []int64 {
	var children []int64
	for _, entry := range idx.Tasks {
		if entry.Parent != nil && *entry.Parent == parentID {
			children = append(children, entry.ID)
		}
	}
	return children
}
