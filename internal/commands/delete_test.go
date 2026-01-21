package commands

import (
	"fmt"
	"testing"
	"time"

	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteCommand(t *testing.T) {
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
	deletePretty = false

	// Delete the task
	err = runDelete(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)

	// Verify task was deleted
	_, err = store.LoadTask(task.ID)
	assert.Error(t, err)
	assert.Equal(t, storage.ErrTaskNotFound, err)
}

func TestDeleteCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	deletePretty = false

	err := runDelete(nil, []string{"999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	deletePretty = false

	err := runDelete(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestDeleteCommand_BlockedByUndoneChildren(t *testing.T) {
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
	deletePretty = false

	// Try to delete parent - should fail
	err = runDelete(nil, []string{fmt.Sprintf("%d", parent.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undone children")

	// Verify parent still exists
	_, err = store.LoadTask(parent.ID)
	require.NoError(t, err)
}

func TestDeleteCommand_AllowedWithDoneChildren(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create parent task
	parent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child task that is done
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child Task",
		Status:  models.StatusDone,
		Parent:  &parent.ID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child))

	// Reset flag
	deletePretty = false

	// Delete parent - should succeed since children are done
	err = runDelete(nil, []string{fmt.Sprintf("%d", parent.ID)})
	require.NoError(t, err)

	// Verify parent was deleted
	_, err = store.LoadTask(parent.ID)
	assert.Error(t, err)

	// Child should still exist and be orphaned
	orphanedChild, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	assert.Nil(t, orphanedChild.Parent)
}

func TestDeleteCommand_BlockedByUndoneGrandchildren(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create grandparent task
	grandparent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Grandparent Task",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(grandparent))

	// Create parent task (done)
	time.Sleep(2 * time.Millisecond)
	parent := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusDone,
		Parent:  &grandparent.ID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child task (undone)
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
	deletePretty = false

	// Try to delete grandparent - should fail due to undone grandchild
	err = runDelete(nil, []string{fmt.Sprintf("%d", grandparent.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "undone children")

	// Verify grandparent still exists
	_, err = store.LoadTask(grandparent.ID)
	require.NoError(t, err)
}

func TestDeleteCommand_PrettyOutput(t *testing.T) {
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

	// Set pretty flag
	deletePretty = true

	// Delete with pretty output
	err = runDelete(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)
}
