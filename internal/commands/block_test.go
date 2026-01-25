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

func TestBlockCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	blocker := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Blocker Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(blocker))

	time.Sleep(2 * time.Millisecond)
	blocked := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Blocked Task",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(blocked))

	blockPretty = false
	err = runBlock(nil, []string{fmt.Sprintf("%d", blocker.ID), fmt.Sprintf("%d", blocked.ID)})
	require.NoError(t, err)

	// Verify blocked task has blocker in BlockedBy
	updated, err := store.LoadTask(blocked.ID)
	require.NoError(t, err)
	assert.Contains(t, updated.BlockedBy, blocker.ID)
}

func TestBlockCommand_SelfBlock(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task))

	blockPretty = false
	err = runBlock(nil, []string{fmt.Sprintf("%d", task.ID), fmt.Sprintf("%d", task.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot block itself")
}

func TestBlockCommand_CycleDetection(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	taskA := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Task A",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(taskA))

	time.Sleep(2 * time.Millisecond)
	taskB := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Task B",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(taskB))

	blockPretty = false

	// A blocks B (B is blocked by A)
	err = runBlock(nil, []string{fmt.Sprintf("%d", taskA.ID), fmt.Sprintf("%d", taskB.ID)})
	require.NoError(t, err)

	// B blocks A should fail (would create cycle)
	err = runBlock(nil, []string{fmt.Sprintf("%d", taskB.ID), fmt.Sprintf("%d", taskA.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cycle")
}

func TestBlockCommand_CannotBlockOnDone(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	doneTask := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Done Task",
		Status:  models.StatusDone,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(doneTask))

	time.Sleep(2 * time.Millisecond)
	todoTask := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Todo Task",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(todoTask))

	blockPretty = false
	err = runBlock(nil, []string{fmt.Sprintf("%d", doneTask.ID), fmt.Sprintf("%d", todoTask.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "completed task")
}

func TestBlockCommand_AlreadyBlocked(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	blocker := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Blocker",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(blocker))

	time.Sleep(2 * time.Millisecond)
	blocked := &models.Task{
		ID:        time.Now().UnixMilli(),
		Name:      "Blocked",
		Status:    models.StatusTodo,
		BlockedBy: []int64{blocker.ID},
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	require.NoError(t, store.SaveTask(blocked))

	blockPretty = false
	err = runBlock(nil, []string{fmt.Sprintf("%d", blocker.ID), fmt.Sprintf("%d", blocked.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already blocked")
}

func TestUnblockCommand(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	blocker := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Blocker",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(blocker))

	time.Sleep(2 * time.Millisecond)
	blocked := &models.Task{
		ID:        time.Now().UnixMilli(),
		Name:      "Blocked",
		Status:    models.StatusTodo,
		BlockedBy: []int64{blocker.ID},
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	require.NoError(t, store.SaveTask(blocked))

	unblockPretty = false
	err = runUnblock(nil, []string{fmt.Sprintf("%d", blocker.ID), fmt.Sprintf("%d", blocked.ID)})
	require.NoError(t, err)

	updated, err := store.LoadTask(blocked.ID)
	require.NoError(t, err)
	assert.NotContains(t, updated.BlockedBy, blocker.ID)
}

func TestUnblockCommand_NotBlocked(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task1 := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Task 1",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(task1))

	time.Sleep(2 * time.Millisecond)
	task2 := &models.Task{
		ID:      time.Now().UnixMilli(),
		Name:    "Task 2",
		Status:  models.StatusTodo,
		Created: time.Now(),
		Updated: time.Now(),
	}
	require.NoError(t, store.SaveTask(task2))

	unblockPretty = false
	err = runUnblock(nil, []string{fmt.Sprintf("%d", task1.ID), fmt.Sprintf("%d", task2.ID)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not blocked")
}

func TestNextCommand_SkipsBlockedTasks(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	blocker := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Blocker Task",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(blocker))

	time.Sleep(2 * time.Millisecond)
	blocked := &models.Task{
		ID:        time.Now().UnixMilli(),
		Name:      "Blocked Task",
		Status:    models.StatusTodo,
		BlockedBy: []int64{blocker.ID},
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	require.NoError(t, store.SaveTask(blocked))

	// next should return blocker, not blocked
	result, err := store.GetNextTask()
	require.NoError(t, err)
	require.NotEmpty(t, result.Candidates)
	assert.Equal(t, "Blocker Task", result.Candidates[0].Name)
	assert.Len(t, result.Candidates, 1) // blocked task should be filtered out
}

func TestStatusCommand_AutoRemovesBlockedBy(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	blocker := &models.Task{
		ID:      now.UnixMilli(),
		Name:    "Blocker",
		Status:  models.StatusTodo,
		Created: now,
		Updated: now,
	}
	require.NoError(t, store.SaveTask(blocker))

	time.Sleep(2 * time.Millisecond)
	blocked := &models.Task{
		ID:        time.Now().UnixMilli(),
		Name:      "Blocked",
		Status:    models.StatusTodo,
		BlockedBy: []int64{blocker.ID},
		Created:   time.Now(),
		Updated:   time.Now(),
	}
	require.NoError(t, store.SaveTask(blocked))

	// Mark blocker as done
	statusPretty = false
	err = runStatus(nil, []string{fmt.Sprintf("%d", blocker.ID), models.StatusDone})
	require.NoError(t, err)

	// Blocked task should no longer be blocked
	updated, err := store.LoadTask(blocked.ID)
	require.NoError(t, err)
	assert.Empty(t, updated.BlockedBy)
}
