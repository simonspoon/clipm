package storage

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"clipm/internal/models"
	"gopkg.in/yaml.v3"
)

const (
	ClipmDir     = ".clipm"
	IndexFile    = "index.json"
	ArchiveDir   = "archive"
	TaskFileExt  = ".md"
	TaskFilePrefix = "task-"
)

var (
	ErrNotInProject = errors.New("not in a clipm project. Run 'clipm init' first")
	ErrTaskNotFound = errors.New("task not found")
)

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

	// Create archive subdirectory
	archivePath := filepath.Join(clipmPath, ArchiveDir)
	if err := os.Mkdir(archivePath, 0755); err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	// Create empty index
	index := models.NewIndex()
	if err := s.SaveIndex(index); err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}

	return nil
}

// LoadIndex loads the index from disk
func (s *Storage) LoadIndex() (*models.Index, error) {
	indexPath := filepath.Join(s.rootDir, ClipmDir, IndexFile)

	data, err := os.ReadFile(indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Index doesn't exist, rebuild from files
			return s.RebuildIndex()
		}
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index models.Index
	if err := json.Unmarshal(data, &index); err != nil {
		// Corrupted index, rebuild
		return s.RebuildIndex()
	}

	return &index, nil
}

// SaveIndex saves the index to disk
func (s *Storage) SaveIndex(index *models.Index) error {
	indexPath := filepath.Join(s.rootDir, ClipmDir, IndexFile)

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// RebuildIndex rebuilds the index from all task files
func (s *Storage) RebuildIndex() (*models.Index, error) {
	index := models.NewIndex()
	clipmPath := filepath.Join(s.rootDir, ClipmDir)

	// Scan active tasks
	if err := s.scanTaskFiles(clipmPath, index, false); err != nil {
		return nil, err
	}

	// Scan archived tasks
	archivePath := filepath.Join(clipmPath, ArchiveDir)
	if err := s.scanTaskFiles(archivePath, index, true); err != nil {
		return nil, err
	}

	// Save the rebuilt index
	if err := s.SaveIndex(index); err != nil {
		return nil, err
	}

	return index, nil
}

// scanTaskFiles scans a directory for task files and adds them to the index
func (s *Storage) scanTaskFiles(dir string, index *models.Index, archived bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), TaskFileExt) {
			continue
		}

		taskPath := filepath.Join(dir, entry.Name())
		task, err := s.readTaskFile(taskPath)
		if err != nil {
			// Skip corrupted files
			continue
		}

		index.AddTask(task, archived)
	}

	return nil
}

// LoadTask loads a task by ID
func (s *Storage) LoadTask(id int64) (*models.Task, error) {
	// Check active tasks first
	taskPath := s.getTaskPath(id, false)
	task, err := s.readTaskFile(taskPath)
	if err == nil {
		return task, nil
	}

	// Check archived tasks
	taskPath = s.getTaskPath(id, true)
	task, err = s.readTaskFile(taskPath)
	if err != nil {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

// SaveTask saves a task to disk and updates the index
func (s *Storage) SaveTask(task *models.Task, archived bool) error {
	taskPath := s.getTaskPath(task.ID, archived)

	if err := s.writeTaskFile(taskPath, task); err != nil {
		return err
	}

	// Update index
	index, err := s.LoadIndex()
	if err != nil {
		return err
	}

	index.AddTask(task, archived)
	return s.SaveIndex(index)
}

// DeleteTask deletes a task file and removes it from the index
func (s *Storage) DeleteTask(id int64) error {
	// Try to delete from active tasks
	taskPath := s.getTaskPath(id, false)
	err := os.Remove(taskPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// If not found, try archived tasks
	if os.IsNotExist(err) {
		taskPath = s.getTaskPath(id, true)
		if err := os.Remove(taskPath); err != nil {
			return fmt.Errorf("failed to delete task: %w", err)
		}
	}

	// Update index
	index, err := s.LoadIndex()
	if err != nil {
		return err
	}

	index.RemoveTask(id)
	return s.SaveIndex(index)
}

// ArchiveTask moves a task to the archive directory
func (s *Storage) ArchiveTask(id int64) error {
	// Load the task
	task, err := s.LoadTask(id)
	if err != nil {
		return err
	}

	// Delete from active location
	activePath := s.getTaskPath(id, false)
	if err := os.Remove(activePath); err != nil {
		return fmt.Errorf("failed to remove active task: %w", err)
	}

	// Save to archive
	task.Status = models.StatusDone
	return s.SaveTask(task, true)
}

// getTaskPath returns the file path for a task
func (s *Storage) getTaskPath(id int64, archived bool) string {
	filename := fmt.Sprintf("%s%d%s", TaskFilePrefix, id, TaskFileExt)
	clipmPath := filepath.Join(s.rootDir, ClipmDir)

	if archived {
		return filepath.Join(clipmPath, ArchiveDir, filename)
	}
	return filepath.Join(clipmPath, filename)
}

// readTaskFile reads a task from a markdown file with YAML frontmatter
func (s *Storage) readTaskFile(path string) (*models.Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Split frontmatter and body
	parts := bytes.SplitN(data, []byte("---"), 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid task file format: missing frontmatter")
	}

	// Parse frontmatter
	var task models.Task
	if err := yaml.Unmarshal(parts[1], &task); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter: %w", err)
	}

	// Store body content
	task.Body = strings.TrimSpace(string(parts[2]))

	return &task, nil
}

// writeTaskFile writes a task to a markdown file with YAML frontmatter
func (s *Storage) writeTaskFile(path string, task *models.Task) error {
	// Marshal frontmatter
	frontmatter, err := yaml.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal frontmatter: %w", err)
	}

	// Build file content
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(frontmatter)
	buf.WriteString("---\n\n")
	if task.Body != "" {
		buf.WriteString(task.Body)
		buf.WriteString("\n")
	}

	// Write to file
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write task file: %w", err)
	}

	return nil
}

// GetRootDir returns the project root directory
func (s *Storage) GetRootDir() string {
	return s.rootDir
}
