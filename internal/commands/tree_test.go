package commands

import (
	"testing"
	"time"

	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestTreeCommand_Empty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Reset flag
	treePretty = true

	// Should not error on empty project
	err := runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_SingleTask(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a single task
	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Single Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Reset flag
	treePretty = true

	// Should display without error
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_SimpleHierarchy(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent
	now := time.Now()
	parent := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Parent Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child Task",
		Parent:  &parent.ID,
		Status:  models.StatusInProgress,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child))

	// Reset flag
	treePretty = true

	// Should display hierarchy
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_ComplexHierarchy(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	root1 := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Root 1",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(root1))

	time.Sleep(2 * time.Millisecond)
	child1 := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child 1",
		Parent:  &root1.ID,
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child1))

	time.Sleep(2 * time.Millisecond)
	child2ID := time.Now().UnixMilli()
	child2 := &models.Task{
		ID:      child2ID,
		Name:    "Child 2",
		Parent:  &root1.ID,
		Status:  models.StatusInProgress,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child2))

	time.Sleep(2 * time.Millisecond)
	grandchild1 := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Grandchild 1",
		Parent:  &child2ID,
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(grandchild1))

	// Reset flag
	treePretty = true

	// Should display full hierarchy
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_MultipleRoots(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create multiple root tasks
	now := time.Now()
	for i := 0; i < 3; i++ {
		task := &models.Task{
			ID:      now.Add(time.Duration(i) * 2 * time.Millisecond).UnixMilli(),
			Name:    "Root Task",
			Status:  models.StatusTodo,
			Created: now.Add(time.Duration(i) * 2 * time.Millisecond),
			Updated: now.Add(time.Duration(i) * 2 * time.Millisecond),
		}
		time.Sleep(2 * time.Millisecond)
		require.NoError(t, store.SaveTask(task))
	}

	// Reset flag
	treePretty = true

	// Should display all roots
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_EmptyJSON(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test empty with JSON output
	treePretty = false

	err := runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_JSONOutput(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a task
	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Test Task",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Test JSON output
	treePretty = false

	err = runTree(nil, []string{})
	require.NoError(t, err)
}
