package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"clipm/internal/models"
)

const (
	ClipmDir  = ".clipm"
	TasksFile = "tasks.json"
)

var (
	ErrNotInProject = errors.New("not in a clipm project. Run 'clipm init' first")
	ErrTaskNotFound = errors.New("task not found")
)

// TaskStore is the root structure for the tasks.json file
type TaskStore struct {
	Version string         `json:"version"`
	Tasks   []models.Task  `json:"tasks"`
}

// Storage handles all file operations for clipm
type Storage struct {
	rootDir string
}

// NewStorage creates a new storage instance
func NewStorage() (*Storage, error) {
	rootDir, err := findProjectRoot()
	if err != nil {
		return nil, err
	}
	return &Storage{rootDir: rootDir}, nil
}

// NewStorageAt creates a storage instance at a specific directory
func NewStorageAt(dir string) *Storage {
	return &Storage{rootDir: dir}
}

// findProjectRoot searches for the .clipm directory in current or parent directories
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		clipmPath := filepath.Join(dir, ClipmDir)
		if info, err := os.Stat(clipmPath); err == nil && info.IsDir() {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", ErrNotInProject
		}
		dir = parent
	}
}

// Init initializes a new clipm project
func (s *Storage) Init() error {
	clipmPath := filepath.Join(s.rootDir, ClipmDir)

	// Check if already exists
	if _, err := os.Stat(clipmPath); err == nil {
		return fmt.Errorf(".clipm directory already exists")
	}

	// Create .clipm directory
	if err := os.Mkdir(clipmPath, 0755); err != nil {
		return fmt.Errorf("failed to create .clipm directory: %w", err)
	}

	// Create empty task store
	store := &TaskStore{
		Version: "2.0.0",
		Tasks:   []models.Task{},
	}
	return s.saveStore(store)
}

// LoadAll loads all tasks from the store
func (s *Storage) LoadAll() ([]models.Task, error) {
	store, err := s.loadStore()
	if err != nil {
		return nil, err
	}
	return store.Tasks, nil
}

// LoadTask loads a task by ID
func (s *Storage) LoadTask(id int64) (*models.Task, error) {
	store, err := s.loadStore()
	if err != nil {
		return nil, err
	}

	for i := range store.Tasks {
		if store.Tasks[i].ID == id {
			return &store.Tasks[i], nil
		}
	}
	return nil, ErrTaskNotFound
}

// SaveTask saves a task (creates or updates)
func (s *Storage) SaveTask(task *models.Task) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	// Check if task exists (update) or is new (create)
	found := false
	for i := range store.Tasks {
		if store.Tasks[i].ID == task.ID {
			store.Tasks[i] = *task
			found = true
			break
		}
	}

	if !found {
		store.Tasks = append(store.Tasks, *task)
	}

	return s.saveStore(store)
}

// DeleteTask deletes a task by ID
func (s *Storage) DeleteTask(id int64) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	newTasks := make([]models.Task, 0, len(store.Tasks))
	found := false
	for _, t := range store.Tasks {
		if t.ID == id {
			found = true
			continue
		}
		newTasks = append(newTasks, t)
	}

	if !found {
		return ErrTaskNotFound
	}

	store.Tasks = newTasks
	return s.saveStore(store)
}

// DeleteTasks deletes multiple tasks by ID
func (s *Storage) DeleteTasks(ids []int64) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	idSet := make(map[int64]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	newTasks := make([]models.Task, 0, len(store.Tasks))
	for _, t := range store.Tasks {
		if !idSet[t.ID] {
			newTasks = append(newTasks, t)
		}
	}

	store.Tasks = newTasks
	return s.saveStore(store)
}

// GetChildren returns all tasks that have the given task as their parent
func (s *Storage) GetChildren(parentID int64) ([]models.Task, error) {
	store, err := s.loadStore()
	if err != nil {
		return nil, err
	}

	var children []models.Task
	for _, t := range store.Tasks {
		if t.Parent != nil && *t.Parent == parentID {
			children = append(children, t)
		}
	}
	return children, nil
}

// GetNextTask returns the oldest todo task (FIFO)
func (s *Storage) GetNextTask() (*models.Task, error) {
	store, err := s.loadStore()
	if err != nil {
		return nil, err
	}

	var todoTasks []models.Task
	for _, t := range store.Tasks {
		if t.Status == models.StatusTodo {
			todoTasks = append(todoTasks, t)
		}
	}

	if len(todoTasks) == 0 {
		return nil, nil
	}

	// Sort by created time (oldest first)
	sort.Slice(todoTasks, func(i, j int) bool {
		return todoTasks[i].Created.Before(todoTasks[j].Created)
	})

	return &todoTasks[0], nil
}

// HasUndoneChildren checks recursively if a task has any descendants that are not done
func (s *Storage) HasUndoneChildren(parentID int64) (bool, error) {
	children, err := s.GetChildren(parentID)
	if err != nil {
		return false, err
	}

	for _, child := range children {
		if child.Status != models.StatusDone {
			return true, nil
		}
		// Check grandchildren recursively
		hasUndone, err := s.HasUndoneChildren(child.ID)
		if err != nil {
			return false, err
		}
		if hasUndone {
			return true, nil
		}
	}
	return false, nil
}

// OrphanChildren sets Parent to nil for all direct children of the given task
func (s *Storage) OrphanChildren(parentID int64) error {
	store, err := s.loadStore()
	if err != nil {
		return err
	}

	for i := range store.Tasks {
		if store.Tasks[i].Parent != nil && *store.Tasks[i].Parent == parentID {
			store.Tasks[i].Parent = nil
		}
	}

	return s.saveStore(store)
}

// loadStore reads the tasks.json file
func (s *Storage) loadStore() (*TaskStore, error) {
	storePath := filepath.Join(s.rootDir, ClipmDir, TasksFile)

	data, err := os.ReadFile(storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskStore{Version: "2.0.0", Tasks: []models.Task{}}, nil
		}
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	var store TaskStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, fmt.Errorf("failed to parse tasks file: %w", err)
	}

	return &store, nil
}

// saveStore writes the tasks.json file
func (s *Storage) saveStore(store *TaskStore) error {
	storePath := filepath.Join(s.rootDir, ClipmDir, TasksFile)

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(storePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// GetRootDir returns the project root directory
func (s *Storage) GetRootDir() string {
	return s.rootDir
}
