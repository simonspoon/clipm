package commands

import (
	"testing"
	"time"

	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNextCommand_NoTasks(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	nextPretty = false

	err := runNext(nil, nil)
	require.NoError(t, err)
}

func TestNextCommand_SingleTask(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Only Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	nextPretty = false

	err = runNext(nil, nil)
	require.NoError(t, err)
}

func TestNextCommand_ReturnsFIFO(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create older task
	older := time.Now()
	task1 := &models.Task{
		ID:      older.UnixMilli(),
		Name:    "Older Task",
		Status:  models.StatusTodo,
		Created: older,
		Updated: older,
	}
	require.NoError(t, store.SaveTask(task1))

	// Create newer task
	time.Sleep(5 * time.Millisecond)
	newer := time.Now()
	task2 := &models.Task{
		ID:      newer.UnixMilli(),
		Name:    "Newer Task",
		Status:  models.StatusTodo,
		Created: newer,
		Updated: newer,
	}
	require.NoError(t, store.SaveTask(task2))

	// No in-progress tasks - returns candidates (older task first)
	next, err := store.GetNextTask()
	require.NoError(t, err)
	require.NotNil(t, next)
	require.NotEmpty(t, next.Candidates)
	assert.Equal(t, "Older Task", next.Candidates[0].Name)
}

func TestNextCommand_SkipsNonTodoTasks(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()

	// Create done task (oldest)
	doneTask := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Done Task",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(doneTask))

	// Create in-progress task
	time.Sleep(2 * time.Millisecond)
	inProgress := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "In Progress Task",
		Status:  models.StatusInProgress,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(inProgress))

	// Create todo task (newest)
	time.Sleep(2 * time.Millisecond)
	todoTask := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Todo Task",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(todoTask))

	// Should return the todo task (sibling of in-progress task)
	next, err := store.GetNextTask()
	require.NoError(t, err)
	require.NotNil(t, next)
	require.NotNil(t, next.Task)
	assert.Equal(t, "Todo Task", next.Task.Name)
}

func TestNextCommand_Pretty(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:          now.UnixMilli(),
		Name:        "Test Task",
		Description: "A description",
		Status:      models.StatusTodo,
		Created:     now,
		Updated:     now,
	}
	require.NoError(t, store.SaveTask(task))

	nextPretty = true

	err = runNext(nil, nil)
	require.NoError(t, err)
}
