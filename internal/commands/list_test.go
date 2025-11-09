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

func createTestTask(t *testing.T, store *storage.Storage, name, status, priority string, tags []string, parent *int64) int64 {
	now := time.Now()
	taskIDCounter++
	task := &models.Task{
		ID:       taskIDCounter,
		Name:     name,
		Status:   status,
		Priority: priority,
		Parent:   parent,
		Created:  now,
		Updated:  now,
		Tags:     tags,
	}
	require.NoError(t, store.SaveTask(task, false))
	return task.ID
}

func TestListFilterByStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different statuses
	createTestTask(t, store, "Todo Task", models.StatusTodo, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "In Progress Task", models.StatusInProgress, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "Done Task", models.StatusDone, models.PriorityMedium, nil, nil)

	// Test filter by status
	listStatus = models.StatusTodo
	listPriority = ""
	listTag = ""
	listParent = 0
	listNoParent = false
	listAll = false

	// Reload index after creating tasks
	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "Todo Task", filtered[0].entry.Name)

	// Test filter in-progress
	listStatus = models.StatusInProgress
	// Reload index
	index, err = store.LoadIndex()
	require.NoError(t, err)
	filtered = filterTasks(index, store)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "In Progress Task", filtered[0].entry.Name)
}

func TestListFilterByPriority(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different priorities
	createTestTask(t, store, "High Task", models.StatusTodo, models.PriorityHigh, nil, nil)
	createTestTask(t, store, "Medium Task", models.StatusTodo, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "Low Task", models.StatusTodo, models.PriorityLow, nil, nil)

	// Test filter by priority
	listStatus = ""
	listPriority = models.PriorityHigh
	listTag = ""
	listParent = 0
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "High Task", filtered[0].entry.Name)
}

func TestListFilterByTag(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different tags
	createTestTask(t, store, "Backend Task", models.StatusTodo, models.PriorityMedium, []string{"backend"}, nil)
	createTestTask(t, store, "Frontend Task", models.StatusTodo, models.PriorityMedium, []string{"frontend"}, nil)
	createTestTask(t, store, "Full Stack Task", models.StatusTodo, models.PriorityMedium, []string{"backend", "frontend"}, nil)

	// Test filter by tag
	listStatus = ""
	listPriority = ""
	listTag = "backend"
	listParent = 0
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 2)

	names := []string{filtered[0].entry.Name, filtered[1].entry.Name}
	assert.Contains(t, names, "Backend Task")
	assert.Contains(t, names, "Full Stack Task")
}

func TestListFilterByParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent and child tasks
	parentID := createTestTask(t, store, "Parent Task", models.StatusTodo, models.PriorityHigh, nil, nil)
	createTestTask(t, store, "Child Task 1", models.StatusTodo, models.PriorityMedium, nil, &parentID)
	createTestTask(t, store, "Child Task 2", models.StatusTodo, models.PriorityMedium, nil, &parentID)
	createTestTask(t, store, "Independent Task", models.StatusTodo, models.PriorityMedium, nil, nil)

	// Test filter by parent
	listStatus = ""
	listPriority = ""
	listTag = ""
	listParent = parentID
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 2)

	names := []string{filtered[0].entry.Name, filtered[1].entry.Name}
	assert.Contains(t, names, "Child Task 1")
	assert.Contains(t, names, "Child Task 2")
}

func TestListFilterNoParent(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create parent and child tasks
	parentID := createTestTask(t, store, "Parent Task", models.StatusTodo, models.PriorityHigh, nil, nil)
	createTestTask(t, store, "Child Task", models.StatusTodo, models.PriorityMedium, nil, &parentID)
	createTestTask(t, store, "Independent Task", models.StatusTodo, models.PriorityMedium, nil, nil)

	// Test filter no parent
	listStatus = ""
	listPriority = ""
	listTag = ""
	listParent = 0
	listNoParent = true
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 2)

	names := []string{filtered[0].entry.Name, filtered[1].entry.Name}
	assert.Contains(t, names, "Parent Task")
	assert.Contains(t, names, "Independent Task")
}

func TestListIncludeArchived(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create active and archived tasks
	activeID := createTestTask(t, store, "Active Task", models.StatusTodo, models.PriorityMedium, nil, nil)
	archivedID := createTestTask(t, store, "Archived Task", models.StatusDone, models.PriorityMedium, nil, nil)

	// Archive one task
	require.NoError(t, store.ArchiveTask(archivedID))

	// Test without archived
	listStatus = ""
	listPriority = ""
	listTag = ""
	listParent = 0
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 1)
	assert.Equal(t, activeID, filtered[0].entry.ID)

	// Test with archived
	listAll = true
	filtered = filterTasks(index, store)
	assert.Len(t, filtered, 2)
}

func TestListMultipleFilters(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create various tasks
	createTestTask(t, store, "High Priority Backend", models.StatusTodo, models.PriorityHigh, []string{"backend"}, nil)
	createTestTask(t, store, "High Priority Frontend", models.StatusTodo, models.PriorityHigh, []string{"frontend"}, nil)
	createTestTask(t, store, "Low Priority Backend", models.StatusTodo, models.PriorityLow, []string{"backend"}, nil)
	createTestTask(t, store, "Done Backend Task", models.StatusDone, models.PriorityHigh, []string{"backend"}, nil)

	// Test multiple filters: status=todo, priority=high, tag=backend
	listStatus = models.StatusTodo
	listPriority = models.PriorityHigh
	listTag = "backend"
	listParent = 0
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	assert.Len(t, filtered, 1)
	assert.Equal(t, "High Priority Backend", filtered[0].entry.Name)
}

func TestGroupByStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create tasks with different statuses
	createTestTask(t, store, "Todo 1", models.StatusTodo, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "Todo 2", models.StatusTodo, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "In Progress 1", models.StatusInProgress, models.PriorityMedium, nil, nil)
	createTestTask(t, store, "Done 1", models.StatusDone, models.PriorityMedium, nil, nil)

	listStatus = ""
	listPriority = ""
	listTag = ""
	listParent = 0
	listNoParent = false
	listAll = false

	index, err := store.LoadIndex()
	require.NoError(t, err)

	filtered := filterTasks(index, store)
	grouped := groupByStatus(filtered)

	assert.Len(t, grouped[models.StatusTodo], 2)
	assert.Len(t, grouped[models.StatusInProgress], 1)
	assert.Len(t, grouped[models.StatusDone], 1)
}
