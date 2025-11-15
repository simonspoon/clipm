package commands

import (
	"os"
	"strings"
	"testing"
	"time"

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

	// Create child task (with slight delay to ensure unique ID)
	time.Sleep(2 * time.Millisecond)
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

func TestAddCommandWithBodyFlag(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = "This is the task body content"

	// Test add with body flag
	err := runAdd(nil, []string{"Task with body"})
	require.NoError(t, err)

	// Verify task body was set
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
	assert.Equal(t, "Task with body", task.Name)
	assert.Equal(t, "This is the task body content", task.Body)
}

func TestAddCommandWithBodyFlagMultiline(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = "# Heading\n\nThis is a paragraph.\n\n- Item 1\n- Item 2"

	// Test add with multiline body
	err := runAdd(nil, []string{"Task with multiline body"})
	require.NoError(t, err)

	// Verify task body preserves formatting
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
	assert.Equal(t, "# Heading\n\nThis is a paragraph.\n\n- Item 1\n- Item 2", task.Body)
}

func TestAddCommandWithStdin(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = ""

	// Mock stdin with pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	// Write test data to stdin
	stdinContent := "Body from stdin"
	_, err = w.WriteString(stdinContent)
	require.NoError(t, err)
	w.Close()

	// Test add with stdin
	err = runAdd(nil, []string{"Task with stdin body"})
	require.NoError(t, err)

	// Verify task body was read from stdin
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
	assert.Equal(t, "Task with stdin body", task.Name)
	assert.Equal(t, strings.TrimSpace(stdinContent), task.Body)
}

func TestAddCommandBodyFlagTakesPrecedenceOverStdin(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = "Body from flag"

	// Mock stdin with pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	// Write test data to stdin (should be ignored)
	_, err = w.WriteString("Body from stdin (should be ignored)")
	require.NoError(t, err)
	w.Close()

	// Test add - flag should take precedence
	err = runAdd(nil, []string{"Task with both flag and stdin"})
	require.NoError(t, err)

	// Verify flag value was used, not stdin
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
	assert.Equal(t, "Body from flag", task.Body)
}

func TestAddCommandEmptyBody(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags (no body flag, no stdin)
	addDescription = "Description only"
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = ""

	// Test add without body (backward compatibility)
	err := runAdd(nil, []string{"Task without body"})
	require.NoError(t, err)

	// Verify task body is empty
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
	assert.Equal(t, "Task without body", task.Name)
	assert.Equal(t, "Description only", task.Description)
	assert.Empty(t, task.Body)
}

func TestAddCommandStdinWithMultilineContent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	addDescription = ""
	addPriority = models.PriorityMedium
	addTags = ""
	addParent = 0
	addBody = ""

	// Mock stdin with pipe
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdin = r

	// Write multiline markdown to stdin
	stdinContent := `# Task Details

This is a detailed description of the task.

## Implementation Steps

1. Step one
2. Step two
3. Step three

## Notes

- Important note 1
- Important note 2`

	_, err = w.WriteString(stdinContent)
	require.NoError(t, err)
	w.Close()

	// Test add with multiline stdin
	err = runAdd(nil, []string{"Task with multiline stdin"})
	require.NoError(t, err)

	// Verify multiline content preserved
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
	assert.Equal(t, strings.TrimSpace(stdinContent), task.Body)
	assert.Contains(t, task.Body, "# Task Details")
	assert.Contains(t, task.Body, "## Implementation Steps")
	assert.Contains(t, task.Body, "1. Step one")
}
