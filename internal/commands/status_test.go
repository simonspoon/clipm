package commands

import (
	"fmt"
	"testing"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test task
	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Test Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	statusPretty = false

	// Test updating status
	err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), models.StatusInProgress})
	require.NoError(t, err)

	// Verify status was updated
	updated, err := store.LoadTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusInProgress, updated.Status)
	assert.True(t, updated.Updated.After(now))
}

func TestStatusCommand_InvalidStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test task
	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Test Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	statusPretty = false

	// Test invalid status
	err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), "invalid-status"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestStatusCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	statusPretty = false

	// Test non-existent task
	err := runStatus(nil, []string{"999999999999", models.StatusDone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStatusCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	statusPretty = false

	// Test invalid ID format
	err := runStatus(nil, []string{"not-a-number", models.StatusDone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestStatusCommand_AllStatuses(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Test Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	statusPretty = false

	// Test each valid status
	statuses := []string{models.StatusTodo, models.StatusInProgress, models.StatusDone}
	for _, status := range statuses {
		err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), status})
		require.NoError(t, err)

		updated, err := store.LoadTask(task.ID)
		require.NoError(t, err)
		assert.Equal(t, status, updated.Status)
	}
}

func TestStatusCommand_CannotMarkDoneWithUndoneChildren(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create parent task
	parent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child task
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child Task",
		Status:  models.StatusTodo,
		Parent:  &parent.ID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child))

	// Reset flag
	statusPretty = false

	// Try to mark parent as done - should fail
	err = runStatus(nil, []string{fmt.Sprintf("%d", parent.ID), models.StatusDone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undone children")

	// Mark child as done
	err = runStatus(nil, []string{fmt.Sprintf("%d", child.ID), models.StatusDone})
	require.NoError(t, err)

	// Now parent can be marked done
	err = runStatus(nil, []string{fmt.Sprintf("%d", parent.ID), models.StatusDone})
	require.NoError(t, err)
}
