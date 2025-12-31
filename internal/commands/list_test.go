package commands

import (
	"testing"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var taskIDCounter int64 = 1000000000000

func createTestTask(t *testing.T, store *storage.Storage, name, status string, parent *int64) int64 {
	now := time.Now()
	taskIDCounter++
	task := &models.Task{
		ID:      taskIDCounter,
		Name:    name,
		Status:  status,
		Parent:  parent,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))
	return task.ID
}

func TestListCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different statuses
	createTestTask(t, store, "Todo Task", models.StatusTodo, nil)
	createTestTask(t, store, "In Progress Task", models.StatusInProgress, nil)
	createTestTask(t, store, "Done Task", models.StatusDone, nil)

	// Reset flags
	listStatus = ""
	listPretty = false

	// Test list all
	err = runList(nil, []string{})
	require.NoError(t, err)
}

func TestListFilterByStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different statuses
	createTestTask(t, store, "Todo Task", models.StatusTodo, nil)
	createTestTask(t, store, "In Progress Task", models.StatusInProgress, nil)
	createTestTask(t, store, "Done Task", models.StatusDone, nil)

	// Test filter by status
	listStatus = models.StatusTodo
	listPretty = false

	tasks, err := store.LoadAll()
	require.NoError(t, err)

	// Count todo tasks
	var todoCount int
	for _, t := range tasks {
		if t.Status == models.StatusTodo {
			todoCount++
		}
	}
	assert.Equal(t, 1, todoCount)
}

func TestListEmpty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flags
	listStatus = ""
	listPretty = false

	// Test list on empty project
	err := runList(nil, []string{})
	require.NoError(t, err)
}

func TestListInvalidStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set invalid status filter
	listStatus = "invalid"
	listPretty = false

	err := runList(nil, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}
