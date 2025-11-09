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

func TestDeleteCommand_WithForce(t *testing.T) {
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

	// Set force flag and delete
	deleteForce = true
	deleteOrphanKids = false
	err = runDelete(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)

	// Verify task was deleted
	_, err = store.LoadTask(task.ID)
	assert.Error(t, err)
	assert.Equal(t, storage.ErrTaskNotFound, err)

	// Verify task is not in index
	index, err := store.LoadIndex()
	require.NoError(t, err)
	_, exists := index.GetTask(task.ID)
	assert.False(t, exists)
}

func TestDeleteCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	deleteForce = true
	err := runDelete(nil, []string{"999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := runDelete(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestDeleteCommand_WithChildren_DeleteAll(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create parent task
	parent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Create child tasks
	time.Sleep(2 * time.Millisecond) // Ensure unique IDs
	child1 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task 1",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Parent:   &parent.ID,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child1, false))

	time.Sleep(2 * time.Millisecond)
	child2 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task 2",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Parent:   &parent.ID,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child2, false))

	// Delete parent with force flag (deletes children too)
	deleteForce = true
	deleteOrphanKids = false
	err = runDelete(nil, []string{fmt.Sprintf("%d", parent.ID)})
	require.NoError(t, err)

	// Verify parent was deleted
	_, err = store.LoadTask(parent.ID)
	assert.Error(t, err)

	// Verify children were deleted
	_, err = store.LoadTask(child1.ID)
	assert.Error(t, err)
	_, err = store.LoadTask(child2.ID)
	assert.Error(t, err)

	// Verify all are removed from index
	index, err := store.LoadIndex()
	require.NoError(t, err)
	_, exists := index.GetTask(parent.ID)
	assert.False(t, exists)
	_, exists = index.GetTask(child1.ID)
	assert.False(t, exists)
	_, exists = index.GetTask(child2.ID)
	assert.False(t, exists)
}

func TestDeleteCommand_WithChildren_Orphan(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create parent task
	parent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Create child task
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Parent:   &parent.ID,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child, false))

	// Delete parent with orphan flag
	deleteForce = true
	deleteOrphanKids = true
	err = runDelete(nil, []string{fmt.Sprintf("%d", parent.ID)})
	require.NoError(t, err)

	// Verify parent was deleted
	_, err = store.LoadTask(parent.ID)
	assert.Error(t, err)

	// Verify child still exists but is orphaned
	orphaned, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	assert.Nil(t, orphaned.Parent)
}

func TestDeleteCommand_NestedChildren(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create grandparent task
	grandparent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Grandparent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(grandparent, false))

	// Create parent task
	time.Sleep(2 * time.Millisecond)
	parent := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Parent:   &grandparent.ID,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Create child task
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Parent:   &parent.ID,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child, false))

	// Delete grandparent (should delete all descendants)
	deleteForce = true
	deleteOrphanKids = false
	err = runDelete(nil, []string{fmt.Sprintf("%d", grandparent.ID)})
	require.NoError(t, err)

	// Verify all were deleted
	_, err = store.LoadTask(grandparent.ID)
	assert.Error(t, err)
	_, err = store.LoadTask(parent.ID)
	assert.Error(t, err)
	_, err = store.LoadTask(child.ID)
	assert.Error(t, err)
}
