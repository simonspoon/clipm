package commands

import (
	"testing"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestTreeCommand_Empty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

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
		ID:       now.UnixMilli(),
		Name:     "Single Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

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
		ID:       now.UnixMilli(),
		Name:     "Parent Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityHigh,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(parent, false))

	// Create child
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child Task",
		Parent:   &parent.ID,
		Status:   models.StatusInProgress,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child, false))

	// Reset flag
	treeAll = false

	// Should display hierarchy
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_ComplexHierarchy(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a multi-level hierarchy
	// Root1
	//   ├─ Child1
	//   └─ Child2
	//       └─ Grandchild1
	// Root2
	//   └─ Child3

	now := time.Now()
	root1 := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Root 1",
		Status:   models.StatusTodo,
		Priority: models.PriorityHigh,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(root1, false))

	time.Sleep(2 * time.Millisecond)
	child1 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child 1",
		Parent:   &root1.ID,
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child1, false))

	time.Sleep(2 * time.Millisecond)
	child2ID := time.Now().UnixMilli()
	child2 := &models.Task{
		ID:       child2ID,
		Name:     "Child 2",
		Parent:   &root1.ID,
		Status:   models.StatusInProgress,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child2, false))

	time.Sleep(2 * time.Millisecond)
	grandchild1 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Grandchild 1",
		Parent:   &child2ID,
		Status:   models.StatusBlocked,
		Priority: models.PriorityLow,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(grandchild1, false))

	time.Sleep(2 * time.Millisecond)
	root2 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Root 2",
		Status:   models.StatusTodo,
		Priority: models.PriorityHigh,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(root2, false))

	time.Sleep(2 * time.Millisecond)
	child3 := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Child 3",
		Parent:   &root2.ID,
		Status:   models.StatusDone,
		Priority: models.PriorityLow,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(child3, false))

	// Reset flag
	treeAll = false

	// Should display full hierarchy
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_ExcludeArchived(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create active task
	now := time.Now()
	active := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Active Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(active, false))

	// Create archived task
	time.Sleep(2 * time.Millisecond)
	archived := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Archived Task",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(archived, true))

	// Reset flag
	treeAll = false

	// Should only show active task
	err = runTree(nil, []string{})
	require.NoError(t, err)
}

func TestTreeCommand_IncludeArchived(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create active task
	now := time.Now()
	active := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Active Task",
		Status:   models.StatusTodo,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(active, false))

	// Create archived task
	time.Sleep(2 * time.Millisecond)
	archived := &models.Task{
		ID:       time.Now().UnixMilli(),
		Name:     "Archived Task",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  time.Now(),
		Updated:  time.Now(),
	}
	require.NoError(t, store.SaveTask(archived, true))

	// Set flag to include archived
	treeAll = true

	// Should show both tasks
	err = runTree(nil, []string{})
	require.NoError(t, err)

	// Reset flag
	treeAll = false
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
			ID:       now.Add(time.Duration(i) * 2 * time.Millisecond).UnixMilli(),
			Name:     "Root Task",
			Status:   models.StatusTodo,
			Priority: models.PriorityMedium,
			Created:  now.Add(time.Duration(i) * 2 * time.Millisecond),
			Updated:  now.Add(time.Duration(i) * 2 * time.Millisecond),
		}
		time.Sleep(2 * time.Millisecond)
		require.NoError(t, store.SaveTask(task, false))
	}

	// Reset flag
	treeAll = false

	// Should display all roots
	err = runTree(nil, []string{})
	require.NoError(t, err)
}
