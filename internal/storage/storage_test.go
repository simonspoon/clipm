package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"clipm/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)

	// Test initialization
	err = store.Init()
	require.NoError(t, err)

	// Verify .clipm directory exists
	clipmPath := filepath.Join(tmpDir, ClipmDir)
	info, err := os.Stat(clipmPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Verify tasks.json exists
	tasksPath := filepath.Join(clipmPath, TasksFile)
	_, err = os.Stat(tasksPath)
	require.NoError(t, err)

	// Test duplicate init fails
	err = store.Init()
	assert.Error(t, err)
}

func TestSaveAndLoadTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create a test task
	now := time.Now()
	task := &models.Task{
		ID:          now.UnixMilli(),
		Name:        "Test Task",
		Description: "Test Description",
		Status:      models.StatusTodo,
		Created:     now,
		Updated:     now,
	}

	// Save the task
	err = store.SaveTask(task)
	require.NoError(t, err)

	// Load the task
	loaded, err := store.LoadTask(task.ID)
	require.NoError(t, err)

	// Verify task fields
	assert.Equal(t, task.ID, loaded.ID)
	assert.Equal(t, task.Name, loaded.Name)
	assert.Equal(t, task.Description, loaded.Description)
	assert.Equal(t, task.Status, loaded.Status)
}

func TestLoadAll(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create multiple tasks
	now := time.Now()
	for i := 0; i < 3; i++ {
		task := &models.Task{
			ID:      now.UnixMilli() + int64(i),
			Name:    "Test Task",
			Status:  models.StatusTodo,
			Created: now,
			Updated: now,
		}
		require.NoError(t, store.SaveTask(task))
	}

	// Load all tasks
	tasks, err := store.LoadAll()
	require.NoError(t, err)
	assert.Len(t, tasks, 3)
}

func TestDeleteTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create and save a test task
	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Task to Delete",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	// Delete the task
	err = store.DeleteTask(task.ID)
	require.NoError(t, err)

	// Verify task is gone
	_, err = store.LoadTask(task.ID)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestDeleteTasks(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create multiple tasks
	now := time.Now()
	var ids []int64
	for i := 0; i < 3; i++ {
		id := now.UnixMilli() + int64(i)
		ids = append(ids, id)
		task := &models.Task{
			ID:      id,
			Name:    "Task",
			Status:  models.StatusTodo,
			Created: now,
			Updated: now,
		}
		require.NoError(t, store.SaveTask(task))
	}

	// Delete first two tasks
	err = store.DeleteTasks(ids[:2])
	require.NoError(t, err)

	// Verify only one task remains
	tasks, err := store.LoadAll()
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, ids[2], tasks[0].ID)
}

func TestTaskWithParent(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create parent task
	now := time.Now()
	parentID := now.UnixMilli()
	parent := &models.Task{
		ID:      parentID,
		Name:    "Parent Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child task
	child := &models.Task{
		ID:      parentID + 1,
		Name:    "Child Task",
		Parent:  &parentID,
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(child))

	// Load and verify child has parent
	loaded, err := store.LoadTask(child.ID)
	require.NoError(t, err)
	require.NotNil(t, loaded.Parent)
	assert.Equal(t, parentID, *loaded.Parent)
}

func TestGetChildren(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create parent task
	now := time.Now()
	parentID := now.UnixMilli()
	parent := &models.Task{
		ID:      parentID,
		Name:    "Parent",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child tasks
	for i := 1; i <= 3; i++ {
		child := &models.Task{
			ID:      parentID + int64(i),
			Name:    "Child",
			Parent:  &parentID,
			Status:  models.StatusTodo,
			Created: now,
			Updated: now,
		}
		require.NoError(t, store.SaveTask(child))
	}

	// Get children
	children, err := store.GetChildren(parentID)
	require.NoError(t, err)
	assert.Len(t, children, 3)
}

func TestGetNextTask(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// No tasks - should return nil
	next, err := store.GetNextTask()
	require.NoError(t, err)
	assert.Nil(t, next)

	// Create tasks with different creation times
	baseTime := time.Now()
	task1 := &models.Task{
		ID:      baseTime.UnixMilli(),
		Name:    "First Task",
		Status:  models.StatusTodo,
		Created: baseTime,
		Updated: baseTime,
	}
	require.NoError(t, store.SaveTask(task1))

	time.Sleep(10 * time.Millisecond)
	task2 := &models.Task{
		ID:      baseTime.UnixMilli() + 10,
		Name:    "Second Task",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(task2))

	// Should get oldest task
	next, err = store.GetNextTask()
	require.NoError(t, err)
	require.NotNil(t, next)
	assert.Equal(t, task1.ID, next.ID)

	// Mark first task as in-progress
	task1.Status = models.StatusInProgress
	require.NoError(t, store.SaveTask(task1))

	// Should get second task
	next, err = store.GetNextTask()
	require.NoError(t, err)
	require.NotNil(t, next)
	assert.Equal(t, task2.ID, next.ID)
}

func TestHasUndoneChildren(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create parent task
	now := time.Now()
	parentID := now.UnixMilli()
	parent := &models.Task{
		ID:      parentID,
		Name:    "Parent",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// No children - should return false
	hasUndone, err := store.HasUndoneChildren(parentID)
	require.NoError(t, err)
	assert.False(t, hasUndone)

	// Add undone child
	child := &models.Task{
		ID:      parentID + 1,
		Name:    "Child",
		Parent:  &parentID,
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(child))

	// Should return true
	hasUndone, err = store.HasUndoneChildren(parentID)
	require.NoError(t, err)
	assert.True(t, hasUndone)

	// Mark child as done
	child.Status = models.StatusDone
	require.NoError(t, store.SaveTask(child))

	// Should return false
	hasUndone, err = store.HasUndoneChildren(parentID)
	require.NoError(t, err)
	assert.False(t, hasUndone)
}

func TestHasUndoneChildrenRecursive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	now := time.Now()
	grandparentID := now.UnixMilli()

	// Create grandparent
	grandparent := &models.Task{
		ID:      grandparentID,
		Name:    "Grandparent",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(grandparent))

	// Create parent (done)
	time.Sleep(2 * time.Millisecond)
	parentID := time.Now().UnixMilli()
	parent := &models.Task{
		ID:      parentID,
		Name:    "Parent",
		Status:  models.StatusDone,
		Parent:  &grandparentID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(parent))

	// Create child (undone)
	time.Sleep(2 * time.Millisecond)
	child := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child",
		Status:  models.StatusTodo,
		Parent:  &parentID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child))

	// Grandparent should have undone descendants
	hasUndone, err := store.HasUndoneChildren(grandparentID)
	require.NoError(t, err)
	assert.True(t, hasUndone)
}

func TestOrphanChildren(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	now := time.Now()
	parentID := now.UnixMilli()

	// Create parent
	parent := &models.Task{
		ID:      parentID,
		Name:    "Parent",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(parent))

	// Create two children
	time.Sleep(2 * time.Millisecond)
	child1 := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child 1",
		Status:  models.StatusDone,
		Parent:  &parentID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child1))

	time.Sleep(2 * time.Millisecond)
	child2 := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Child 2",
		Status:  models.StatusDone,
		Parent:  &parentID,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(child2))

	// Verify children have parent
	children, err := store.GetChildren(parentID)
	require.NoError(t, err)
	assert.Len(t, children, 2)

	// Orphan children
	err = store.OrphanChildren(parentID)
	require.NoError(t, err)

	// Verify children are orphaned
	children, err = store.GetChildren(parentID)
	require.NoError(t, err)
	assert.Len(t, children, 0)

	// Verify children still exist but have no parent
	loadedChild1, err := store.LoadTask(child1.ID)
	require.NoError(t, err)
	assert.Nil(t, loadedChild1.Parent)

	loadedChild2, err := store.LoadTask(child2.ID)
	require.NoError(t, err)
	assert.Nil(t, loadedChild2.Parent)
}

func TestGetRootDir(t *testing.T) {
	store := NewStorageAt("/some/path")
	assert.Equal(t, "/some/path", store.GetRootDir())
}

func TestNewStorage(t *testing.T) {
	// Create temp directory with .clipm initialized
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Save original directory and change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	// NewStorage should find the project
	newStore, err := NewStorage()
	require.NoError(t, err)
	assert.Equal(t, tmpDir, newStore.GetRootDir())
}

func TestNewStorageNotInProject(t *testing.T) {
	// Create temp directory without .clipm
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Save original directory and change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer os.Chdir(origDir)

	// NewStorage should fail
	_, err = NewStorage()
	assert.Equal(t, ErrNotInProject, err)
}

func TestFindProjectRootInParent(t *testing.T) {
	// Create temp directory structure: parent/.clipm and parent/child
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Resolve symlinks (macOS /var -> /private/var)
	tmpDir, err = filepath.EvalSymlinks(tmpDir)
	require.NoError(t, err)

	// Initialize .clipm in parent
	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create nested child directory
	childDir := filepath.Join(tmpDir, "child", "grandchild")
	require.NoError(t, os.MkdirAll(childDir, 0755))

	// Save original directory and change to child dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(childDir))
	defer os.Chdir(origDir)

	// NewStorage should find the project in parent
	newStore, err := NewStorage()
	require.NoError(t, err)
	assert.Equal(t, tmpDir, newStore.GetRootDir())
}

func TestLoadStoreCorruptedJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Corrupt the tasks.json file
	tasksPath := filepath.Join(tmpDir, ClipmDir, TasksFile)
	err = os.WriteFile(tasksPath, []byte("not valid json{"), 0644)
	require.NoError(t, err)

	// LoadAll should fail
	_, err = store.LoadAll()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse")
}

func TestDeleteTaskNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Delete non-existent task
	err = store.DeleteTask(999999999)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestDeleteTaskWithMultipleTasks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Create multiple tasks
	now := time.Now()
	for i := 0; i < 5; i++ {
		task := &models.Task{
			ID:      now.UnixMilli() + int64(i),
			Name:    "Task",
			Status:  models.StatusTodo,
			Created: now,
			Updated: now,
		}
		require.NoError(t, store.SaveTask(task))
	}

	// Delete middle task
	err = store.DeleteTask(now.UnixMilli() + 2)
	require.NoError(t, err)

	// Verify 4 tasks remain
	tasks, err := store.LoadAll()
	require.NoError(t, err)
	assert.Len(t, tasks, 4)
}

func TestLoadTaskNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Load non-existent task
	_, err = store.LoadTask(999999999)
	assert.Equal(t, ErrTaskNotFound, err)
}

func TestGetChildrenEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "clipm-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	store := NewStorageAt(tmpDir)
	require.NoError(t, store.Init())

	// Get children of non-existent task
	children, err := store.GetChildren(999999999)
	require.NoError(t, err)
	assert.Len(t, children, 0)
}
