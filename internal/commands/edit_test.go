package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEditCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test non-existent task
	err := runEdit(nil, []string{"999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestEditCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test invalid ID format
	err := runEdit(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestEditCommand_GetTaskFilePath(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	taskID := int64(1234567890000)

	// Test active task path
	activePath := getTaskFilePath(store, taskID, false)
	assert.Contains(t, activePath, storage.ClipmDir)
	assert.Contains(t, activePath, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, taskID, storage.TaskFileExt))
	assert.NotContains(t, activePath, storage.ArchiveDir)

	// Test archived task path
	archivedPath := getTaskFilePath(store, taskID, true)
	assert.Contains(t, archivedPath, storage.ClipmDir)
	assert.Contains(t, archivedPath, storage.ArchiveDir)
	assert.Contains(t, archivedPath, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, taskID, storage.TaskFileExt))
}

func TestEditCommand_ValidateYAML(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a test task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Get task file path
	taskPath := filepath.Join(tmpDir, storage.ClipmDir, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, task.ID, storage.TaskFileExt))

	// Corrupt the YAML by removing the closing ---
	content, err := os.ReadFile(taskPath)
	require.NoError(t, err)

	// Write invalid YAML
	invalidYAML := "---\ninvalid: yaml: content:\n\nBody content"
	require.NoError(t, os.WriteFile(taskPath, []byte(invalidYAML), 0644))

	// Try to load the task - should fail
	_, err = store.LoadTask(task.ID)
	assert.Error(t, err)

	// Restore original content
	require.NoError(t, os.WriteFile(taskPath, content, 0644))
}

func TestEditCommand_NoEditorFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a test task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Save original PATH and EDITOR
	origPath := os.Getenv("PATH")
	origEditor := os.Getenv("EDITOR")

	// Clear PATH and EDITOR to simulate no editor found
	os.Setenv("PATH", "")
	os.Setenv("EDITOR", "")

	// Restore after test
	defer func() {
		os.Setenv("PATH", origPath)
		os.Setenv("EDITOR", origEditor)
	}()

	// Should fail with no editor found
	err = runEdit(nil, []string{fmt.Sprintf("%d", task.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no editor found")
}

func TestEditCommand_ArchivedTask(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create an archived task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Archived Task",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, true))

	// Verify getTaskFilePath returns correct path for archived task
	path := getTaskFilePath(store, task.ID, true)
	assert.Contains(t, path, storage.ClipmDir)
	assert.Contains(t, path, storage.ArchiveDir)
	assert.Contains(t, path, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, task.ID, storage.TaskFileExt))

	// Verify the file exists at that path
	_, err = os.Stat(path)
	require.NoError(t, err)
}
