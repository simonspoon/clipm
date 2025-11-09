package commands

import (
	"os"
	"testing"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestEnv(t *testing.T) (string, func()) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-cmd-test-*")
	require.NoError(t, err)

	// Change to temp directory
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))

	// Initialize clipm
	store := storage.NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Return cleanup function
	cleanup := func() {
		os.Chdir(origDir)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestAddCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0

	// Test basic add
	err := runAdd(nil, []string{"Test Task"})
	require.NoError(t, err)

	// Verify task was created
	store, err := storage.NewStorage()
	require.NoError(t, err)

	index, err := store.LoadIndex()
	require.NoError(t, err)
	assert.Len(t, index.Tasks, 1)

	// Get the task
	var taskID int64
	for id := range index.Tasks {
		taskID = id
	}

	task, err := store.LoadTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, "Test Task", task.Name)
	assert.Equal(t, models.StatusTodo, task.Status)
	assert.Equal(t, models.PriorityMedium, task.Priority)
	assert.Empty(t, task.Description)
	assert.Empty(t, task.Tags)
	assert.Nil(t, task.Parent)
}

func TestAddCommandWithAllFlags(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set flags
	addDescription = "Test description"
	addPriority = models.PriorityHigh
	addTags = "tag1,tag2,tag3"
	addParent = 0

	// Test add with flags
	err := runAdd(nil, []string{"Task with flags"})
	require.NoError(t, err)

	// Verify task
	store, err := storage.NewStorage()
	require.NoError(t, err)

	index, err := store.LoadIndex()
	require.NoError(t, err)

	var taskID int64
	for id := range index.Tasks {
		taskID = id
	}

	task, err := store.LoadTask(taskID)
	require.NoError(t, err)
	assert.Equal(t, "Task with flags", task.Name)
	assert.Equal(t, "Test description", task.Description)
	assert.Equal(t, models.PriorityHigh, task.Priority)
	assert.Equal(t, []string{"tag1", "tag2", "tag3"}, task.Tags)
}

func TestAddCommandWithParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create parent task
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0

	err := runAdd(nil, []string{"Parent Task"})
	require.NoError(t, err)

	// Get parent ID
	store, err := storage.NewStorage()
	require.NoError(t, err)

	index, err := store.LoadIndex()
	require.NoError(t, err)

	var parentID int64
	for id := range index.Tasks {
		parentID = id
	}

	// Create child task
	addParent = parentID
	err = runAdd(nil, []string{"Child Task"})
	require.NoError(t, err)

	// Verify child has parent
	index, err = store.LoadIndex()
	require.NoError(t, err)
	assert.Len(t, index.Tasks, 2)

	// Find child task
	var childID int64
	for id, entry := range index.Tasks {
		if entry.Name == "Child Task" {
			childID = id
		}
	}

	child, err := store.LoadTask(childID)
	require.NoError(t, err)
	require.NotNil(t, child.Parent)
	assert.Equal(t, parentID, *child.Parent)
}

func TestAddCommandInvalidPriority(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set invalid priority
	addPriority = "invalid"
	addDescription = ""
	addTags = ""
	addParent = 0

	// Test add should fail
	err := runAdd(nil, []string{"Test Task"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid priority")
}

func TestAddCommandNonExistentParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set non-existent parent
	addParent = 999999999999
	addPriority = models.PriorityMedium
	addDescription = ""
	addTags = ""

	// Test add should fail
	err := runAdd(nil, []string{"Test Task"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent task")
}

func TestAddCommandNotInProject(t *testing.T) {
	// Create temp directory without initializing
	tmpDir, err := os.MkdirTemp("", "clipm-cmd-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(origDir)

	require.NoError(t, os.Chdir(tmpDir))

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0

	// Test add should fail
	err = runAdd(nil, []string{"Test Task"})
	assert.Error(t, err)
}

func TestAddCommandTagsParsing(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	tests := []struct {
		name     string
		tagInput string
		expected []string
	}{
		{
			name:     "single tag",
			tagInput: "backend",
			expected: []string{"backend"},
		},
		{
			name:     "multiple tags",
			tagInput: "backend,frontend,api",
			expected: []string{"backend", "frontend", "api"},
		},
		{
			name:     "tags with spaces",
			tagInput: "backend, frontend , api",
			expected: []string{"backend", "frontend", "api"},
		},
		{
			name:     "empty tag",
			tagInput: "backend,,frontend",
			expected: []string{"backend", "frontend"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addTags = tt.tagInput
			addDescription = ""
			addPriority = models.PriorityMedium
			addParent = 0

			err := runAdd(nil, []string{"Test Task"})
			require.NoError(t, err)

			// Get the latest task
			store, err := storage.NewStorage()
			require.NoError(t, err)

			index, err := store.LoadIndex()
			require.NoError(t, err)

			var maxID int64
			for id := range index.Tasks {
				if id > maxID {
					maxID = id
				}
			}

			task, err := store.LoadTask(maxID)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, task.Tags)
		})
	}
}
