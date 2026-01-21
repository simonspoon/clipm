package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/simonspoon/clipm/internal/models"
)

// Storage directory and file names.
const (
	ClipmDir  = ".clipm"
	TasksFile = "tasks.json"
)

// Storage errors.
var (
	ErrNotInProject = errors.New("not in a clipm project. Run 'clipm init' first")
	ErrTaskNotFound = errors.New("task not found")
)

// TaskStore is the root structure for the tasks.json file
type TaskStore struct {
	Version string        `json:"version"`
	Tasks   []models.Task `json:"tasks"`
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

// NextResult represents the result of GetNextTask
type NextResult struct {
	Task       *models.Task  `json:"task,omitempty"`
	Candidates []models.Task `json:"candidates,omitempty"`
}

// GetNextTask returns the next task using depth-first traversal.
// When in-progress tasks exist: returns todo children or siblings of the deepest in-progress task.
// When no in-progress tasks: returns root-level todos as candidates.
func (s *Storage) GetNextTask() (*NextResult, error) {
	store, err := s.loadStore()
	if err != nil {
		return nil, err
	}

	deepest := getDeepestInProgress(store.Tasks)
	if deepest == nil {
		// No in-progress context - return root-level todos as candidates
		candidates := getRootTodos(store.Tasks)
		return &NextResult{Candidates: candidates}, nil
	}

	// Walk up from deepest, looking for todo children first, then siblings
	current := deepest
	for {
		// First, check for todo children of current task
		children := getTodoChildren(store.Tasks, current.ID)
		if len(children) > 0 {
			return &NextResult{Task: &children[0]}, nil
		}

		// Then, check for todo siblings
		siblings := getTodoSiblings(store.Tasks, current.ID)
		if len(siblings) > 0 {
			return &NextResult{Task: &siblings[0]}, nil
		}

		// Move up to parent
		if current.Parent == nil {
			break
		}
		parent := findTask(store.Tasks, *current.Parent)
		if parent == nil {
			break
		}
		current = parent
	}
	return &NextResult{}, nil
}

// getDeepestInProgress finds the in-progress task that has no in-progress children
func getDeepestInProgress(tasks []models.Task) *models.Task {
	// Build map of tasks that have in-progress children
	hasInProgressChild := make(map[int64]bool)
	for _, t := range tasks {
		if t.Status == models.StatusInProgress && t.Parent != nil {
			hasInProgressChild[*t.Parent] = true
		}
	}

	// Find in-progress task with no in-progress children (deepest)
	var deepest *models.Task
	for i := range tasks {
		if tasks[i].Status == models.StatusInProgress && !hasInProgressChild[tasks[i].ID] {
			if deepest == nil || tasks[i].Created.Before(deepest.Created) {
				deepest = &tasks[i]
			}
		}
	}
	return deepest
}

// getTodoChildren returns todo tasks that are children of the given task, sorted by created time
func getTodoChildren(tasks []models.Task, parentID int64) []models.Task {
	var children []models.Task
	for _, t := range tasks {
		if t.Status == models.StatusTodo && t.Parent != nil && *t.Parent == parentID {
			children = append(children, t)
		}
	}
	sort.Slice(children, func(i, j int) bool {
		return children[i].Created.Before(children[j].Created)
	})
	return children
}

// getTodoSiblings returns todo tasks with the same parent as the given task, sorted by created time
func getTodoSiblings(tasks []models.Task, taskID int64) []models.Task {
	// Find the task to get its parent
	var targetParent *int64
	for i := range tasks {
		if tasks[i].ID == taskID {
			targetParent = tasks[i].Parent
			break
		}
	}

	// Find all todo tasks with the same parent
	var siblings []models.Task
	for _, t := range tasks {
		if t.Status != models.StatusTodo {
			continue
		}
		// Check if same parent (both nil or both point to same ID)
		sameParent := (targetParent == nil && t.Parent == nil) ||
			(targetParent != nil && t.Parent != nil && *targetParent == *t.Parent)
		if sameParent {
			siblings = append(siblings, t)
		}
	}

	// Sort by created time (oldest first)
	sort.Slice(siblings, func(i, j int) bool {
		return siblings[i].Created.Before(siblings[j].Created)
	})

	return siblings
}

// getRootTodos returns all todo tasks with no parent, sorted by created time
func getRootTodos(tasks []models.Task) []models.Task {
	var roots []models.Task
	for _, t := range tasks {
		if t.Status == models.StatusTodo && t.Parent == nil {
			roots = append(roots, t)
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		return roots[i].Created.Before(roots[j].Created)
	})
	return roots
}

// findTask finds a task by ID
func findTask(tasks []models.Task, id int64) *models.Task {
	for i := range tasks {
		if tasks[i].ID == id {
			return &tasks[i]
		}
	}
	return nil
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
