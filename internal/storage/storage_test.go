package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clipm/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)

	// Test initialization
	err = store.Init()
	require.NoError(t, err)

	// Verify .clipm directory exists
	clipmPath := filepath.Join(tmpDir, ClipmDir)
	info, err := os.Stat(clipmPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify archive directory exists
	archivePath := filepath.Join(clipmPath, ArchiveDir)
	info, err = os.Stat(archivePath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify index.json exists and is valid
	index, err := store.LoadIndex()
	require.NoError(t, err)
	assert.Equal(t, "1.0.0", index.Version)
	assert.Empty(t, index.Tasks)

	// Test duplicate init fails
	err = store.Init()
	assert.Error(t, err)
}

func TestSaveAndLoadTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create a test task
	now := time.Now()
	task := &models.Task{
		ID:          now.UnixMilli(),
		Name:        "Test Task",
		Description: "Test Description",
		Status:      models.StatusTodo,
		Priority:    models.PriorityHigh,
		Created:     now,
		Updated:     now,
		Tags:        []string{"test", "example"},
		Body:        "## Notes\n\nThis is a test task.",
	}

	// Save the task
	err = store.SaveTask(task, false)
	require.NoError(t, err)

	// Load the task
	loaded, err := store.LoadTask(task.ID)
	require.NoError(t, err)

	// Verify task fields
	assert.Equal(t, task.ID, loaded.ID)
	assert.Equal(t, task.Name, loaded.Name)
	assert.Equal(t, task.Description, loaded.Description)
	assert.Equal(t, task.Status, loaded.Status)
	assert.Equal(t, task.Priority, loaded.Priority)
	assert.Equal(t, task.Tags, loaded.Tags)
	assert.Equal(t, task.Body, loaded.Body)

	// Verify index was updated
	index, err := store.LoadIndex()
	require.NoError(t, err)
	entry, exists := index.GetTask(task.ID)
	require.True(t, exists)
	assert.Equal(t, task.Name, entry.Name)
	assert.False(t, entry.Archived)
}

func TestArchiveTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create and save a test task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Task to Archive",
		Status:   models.StatusInProgress,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Archive the task
	err = store.ArchiveTask(task.ID)
	require.NoError(t, err)

	// Verify task is in archive
	loaded, err := store.LoadTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, loaded.Status)

	// Verify index shows archived
	index, err := store.LoadIndex()
	require.NoError(t, err)
	entry, exists := index.GetTask(task.ID)
	require.True(t, exists)
	assert.True(t, entry.Archived)

	// Verify task file is in archive directory
	archivePath := store.getTaskPath(task.ID, true)
	_, err = os.Stat(archivePath)
	assert.NoError(t, err)
}

func TestDeleteTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create and save a test task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Task to Delete",
		Status:   models.StatusTodo,
		Priority: models.PriorityLow,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Delete the task
	err = store.DeleteTask(task.ID)
	require.NoError(t, err)

	// Verify task file is deleted
	taskPath := store.getTaskPath(task.ID, false)
	_, err = os.Stat(taskPath)
	assert.True(t, os.IsNotExist(err))

	// Verify task is removed from index
	index, err := store.LoadIndex()
	require.NoError(t, err)
	_, exists := index.GetTask(task.ID)
	assert.False(t, exists)
}

func TestRebuildIndex(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create multiple tasks
	now := time.Now()
	for i := 0; i < 3; i++ {
		task := &models.Task{
			ID:       now.UnixMilli() + int64(i),
			Name:     "Test Task " + string(rune(i)),
			Status:   models.StatusTodo,
			Priority: models.PriorityMedium,
			Created:  now,
			Updated:  now,
		}
		require.NoError(t, store.SaveTask(task, false))
	}

	// Delete the index file
	indexPath := filepath.Join(tmpDir, ClipmDir, IndexFile)
	require.NoError(t, os.Remove(indexPath))

	// Rebuild index
	index, err := store.LoadIndex()
	require.NoError(t, err)

	// Verify all tasks are in rebuilt index
	assert.Len(t, index.Tasks, 3)
}

func TestTaskWithParent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create parent task
	now := time.Now()
	parentID := now.UnixMilli()
	parent := &models.Task{
		ID:       parentID,
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityHigh,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Create child task
	child := &models.Task{
		ID:       parentID + 1,
		Name:     "Child Task",
		Parent:   &parentID,
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(child, false))

	// Load and verify child has parent
	loaded, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.Parent)
	assert.Equal(t, parentID, *loaded.Parent)
}
