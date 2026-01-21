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

func TestUnparentCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent and child with relationship
	now := time.Now()
	parent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	time.Sleep(2 * time.Millisecond)
	childNow := time.Now()
	child := &models.Task{
		ID:      childNow.UnixMilli(),
		Name:    "Child Task",
		Parent:  &parent.ID,
		Status:  models.StatusTodo,
		Created: childNow,
		Updated: childNow,
	}
	require.NoError(t, store.SaveTask(child))

	// Reset flag
	unparentPretty = false

	// Unparent the child
	err = runUnparent(nil, []string{fmt.Sprintf("%d", child.ID)})
	require.NoError(t, err)

	// Verify parent was removed
	updated, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Parent)
}

func TestUnparentCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	unparentPretty = false

	err := runUnparent(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestUnparentCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	unparentPretty = false

	err := runUnparent(nil, []string{"999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestUnparentCommand_AlreadyTopLevel(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create task without parent
	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Top Level Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	unparentPretty = false

	// Unparent already top-level task (should not error, just inform)
	err = runUnparent(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)

	// Verify still no parent
	updated, err := store.LoadTask(task.ID)
	require.NoError(t, err)
	assert.Nil(t, updated.Parent)
}

func TestUnparentCommand_PrettyOutput(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent and child
	now := time.Now()
	parent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child Task",
		Parent:  &parent.ID,
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child))

	// Set pretty flag
	unparentPretty = true

	err = runUnparent(nil, []string{fmt.Sprintf("%d", child.ID)})
	require.NoError(t, err)
}

func TestUnparentCommand_AlreadyTopLevelPretty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Top Level Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Set pretty flag
	unparentPretty = true

	err = runUnparent(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)
}
