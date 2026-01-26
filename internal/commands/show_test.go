package commands

import (
	"os"
	"testing"

	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShowCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a test task
	taskID := createTestTask(t, store, "Test Task", models.StatusTodo, nil)

	// Reset flag
	showPretty = false

	// Test show command
	err = runShow(nil, []string{taskID})
	require.NoError(t, err)
}

func TestShowCommandWithAllFields(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent task
	parentID := createTestTask(t, store, "Parent Task", models.StatusTodo, nil)

	// Create task with parent
	taskID := createTestTask(t, store, "Child Task", models.StatusInProgress, &parentID)

	// Load task to add description
	task, err := store.LoadTask(taskID)
	require.NoError(t, err)
	task.Description = "Full description"
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	showPretty = false

	// Show the task - just verify it doesn't error
	err = runShow(nil, []string{taskID})
	require.NoError(t, err)
}

func TestShowCommandTaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	showPretty = false

	// Test show non-existent task
	err := runShow(nil, []string{"zzzz"})
	assert.Error(t, err)
}

func TestShowCommandInvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	showPretty = false

	// Test show with invalid ID (wrong length)
	err := runShow(nil, []string{"not-valid"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
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

	// Reset flag
	showPretty = false

	// Test show should fail
	err = runShow(nil, []string{"aaaa"})
	assert.Error(t, err)
}
