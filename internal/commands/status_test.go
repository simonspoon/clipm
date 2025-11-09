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
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

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
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Test invalid status
	err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), "invalid-status"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}

func TestStatusCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test non-existent task
	err := runStatus(nil, []string{"999999999999", models.StatusDone})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestStatusCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

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
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Test each valid status
	statuses := []string{models.StatusTodo, models.StatusInProgress, models.StatusDone, models.StatusBlocked}
	for _, status := range statuses {
		err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), status})
		require.NoError(t, err)

		updated, err := store.LoadTask(task.ID)
		require.NoError(t, err)
		assert.Equal(t, status, updated.Status)
	}
}

func TestStatusCommand_ArchivedTask(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create and archive a task
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, true))

	// Should be able to update archived task status
	err = runStatus(nil, []string{fmt.Sprintf("%d", task.ID), models.StatusTodo})
	require.NoError(t, err)

	updated, err := store.LoadTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusTodo, updated.Status)
}
