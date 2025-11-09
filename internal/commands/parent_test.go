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

func TestParentCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent and child tasks
	now := time.Now()
	parent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	time.Sleep(2 * time.Millisecond) // Ensure different ID
	child := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child, false))

	// Set parent
	err = runParent(nil, []string{fmt.Sprintf("%d", child.ID), fmt.Sprintf("%d", parent.ID)})
	require.NoError(t, err)

	// Verify parent was set
	updated, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	assert.NotNil(t, updated.Parent)
	assert.Equal(t, parent.ID, *updated.Parent)
	assert.True(t, updated.Updated.After(child.Updated))
}

func TestParentCommand_InvalidChildID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := runParent(nil, []string{"not-a-number", "123456"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestParentCommand_InvalidParentID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	err := runParent(nil, []string{"123456", "not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid parent ID")
}

func TestParentCommand_SelfParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a task
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

	// Try to set task as its own parent
	err = runParent(nil, []string{fmt.Sprintf("%d", task.ID), fmt.Sprintf("%d", task.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set task as its own parent")
}

func TestParentCommand_ChildNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent task
	now := time.Now()
	parent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Try to set parent for non-existent child
	err = runParent(nil, []string{"999999999999", fmt.Sprintf("%d", parent.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestParentCommand_ParentNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create child task
	now := time.Now()
	child := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Child Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(child, false))

	// Try to set non-existent parent
	err = runParent(nil, []string{fmt.Sprintf("%d", child.ID), "999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "parent task")
	assert.Contains(t, err.Error(), "not found")
}

func TestParentCommand_ArchivedParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create archived parent
	now := time.Now()
	parent := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Archived Parent",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, true))

	// Create child
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child, false))

	// Try to set archived task as parent
	err = runParent(nil, []string{fmt.Sprintf("%d", child.ID), fmt.Sprintf("%d", parent.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot set archived task")
}

func TestParentCommand_CircularDependency(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a chain: A -> B -> C
	now := time.Now()
	taskA := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Task A",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(taskA, false))

	time.Sleep(2 * time.Millisecond)
	taskBID := time.Now().UnixMilli()
	taskB := &models.Task{
		ID:       taskBID,
		Name:     "Task B",
		Parent:   &taskA.ID,
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(taskB, false))

	time.Sleep(2 * time.Millisecond)
	taskC := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Task C",
		Parent:   &taskBID,
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(taskC, false))

	// Try to create a cycle: A -> B -> C -> A
	err = runParent(nil, []string{fmt.Sprintf("%d", taskA.ID), fmt.Sprintf("%d", taskC.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestParentCommand_DirectCycle(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create A -> B
	now := time.Now()
	taskA := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Task A",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(taskA, false))

	time.Sleep(2 * time.Millisecond)
	taskB := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Task B",
		Parent:   &taskA.ID,
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(taskB, false))

	// Try to create B -> A (direct cycle)
	err = runParent(nil, []string{fmt.Sprintf("%d", taskA.ID), fmt.Sprintf("%d", taskB.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}
