package commands

import (
	"os"
	"strconv"
	"testing"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a test task
	taskID := createTestTask(t, store, "Test Task", models.StatusTodo, models.PriorityHigh, []string{"test"}, nil)

	// Test show command
	err = runShow(nil, []string{strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
}

func TestShowCommandWithAllFields(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent task
	parentID := createTestTask(t, store, "Parent Task", models.StatusTodo, models.PriorityMedium, nil, nil)

	// Create task with all fields
	taskID := createTestTask(t, store, "Full Task", models.StatusInProgress, models.PriorityHigh, []string{"backend", "security"}, &parentID)

	// Load task to add body
	task, err := store.LoadTask(taskID)
	require.NoError(t, err)
	task.Description = "Full description"
	task.Body = "## Notes\n\nThis is a test task with body content."
	require.NoError(t, store.SaveTask(task, false))

	// Show the task - just verify it doesn't error
	err = runShow(nil, []string{strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
}

func TestShowCommandTaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test show non-existent task
	err := runShow(nil, []string{"999999999999"})
	assert.Error(t, err)
}

func TestShowCommandInvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test show with invalid ID
	err := runShow(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestShowCommandArchivedTask(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create and archive a task
	taskID := createTestTask(t, store, "Archived Task", models.StatusTodo, models.PriorityMedium, nil, nil)
	require.NoError(t, store.ArchiveTask(taskID))

	// Show should still work for archived tasks
	err = runShow(nil, []string{strconv.FormatInt(taskID, 10)})
	require.NoError(t, err)
}

func TestShowCommandNotInProject(t *testing.T) {
	// Create temp directory without initializing
	tmpDir, err := os.MkdirTemp("", "clipm-cmd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(tmpDir))

	// Test show should fail
	err = runShow(nil, []string{"123456789"})
	assert.Error(t, err)
}
