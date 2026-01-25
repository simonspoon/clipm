package commands

import (
	"testing"
	"time"

	"github.com/simonspoon/clipm/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDetectChanges(t *testing.T) {
	now := time.Now()

	prev := map[int64]models.Task{
		1: {ID: 1, Name: "Task 1", Updated: now},
		2: {ID: 2, Name: "Task 2", Updated: now},
		3: {ID: 3, Name: "Task 3", Updated: now},
	}

	updatedTime := now.Add(time.Second)
	curr := map[int64]models.Task{
		1: {ID: 1, Name: "Task 1", Updated: now},                 // unchanged
		2: {ID: 2, Name: "Task 2 Updated", Updated: updatedTime}, // updated
		4: {ID: 4, Name: "Task 4", Updated: now},                 // added
	}

	added, updated, deleted := detectChanges(prev, curr)

	assert.ElementsMatch(t, []int64{4}, added)
	assert.ElementsMatch(t, []int64{2}, updated)
	assert.ElementsMatch(t, []int64{3}, deleted)
}

func TestDetectChangesEmpty(t *testing.T) {
	prev := map[int64]models.Task{}
	curr := map[int64]models.Task{}

	added, updated, deleted := detectChanges(prev, curr)

	assert.Empty(t, added)
	assert.Empty(t, updated)
	assert.Empty(t, deleted)
}

func TestDetectChangesAllNew(t *testing.T) {
	now := time.Now()
	prev := map[int64]models.Task{}
	curr := map[int64]models.Task{
		1: {ID: 1, Name: "Task 1", Updated: now},
		2: {ID: 2, Name: "Task 2", Updated: now},
	}

	added, updated, deleted := detectChanges(prev, curr)

	assert.ElementsMatch(t, []int64{1, 2}, added)
	assert.Empty(t, updated)
	assert.Empty(t, deleted)
}

func TestDetectChangesAllDeleted(t *testing.T) {
	now := time.Now()
	prev := map[int64]models.Task{
		1: {ID: 1, Name: "Task 1", Updated: now},
		2: {ID: 2, Name: "Task 2", Updated: now},
	}
	curr := map[int64]models.Task{}

	added, updated, deleted := detectChanges(prev, curr)

	assert.Empty(t, added)
	assert.Empty(t, updated)
	assert.ElementsMatch(t, []int64{1, 2}, deleted)
}

func TestFilterByStatus(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Name: "Task 1", Status: models.StatusTodo},
		{ID: 2, Name: "Task 2", Status: models.StatusInProgress},
		{ID: 3, Name: "Task 3", Status: models.StatusDone},
		{ID: 4, Name: "Task 4", Status: models.StatusTodo},
	}

	filtered := filterByStatus(tasks, models.StatusTodo)
	assert.Len(t, filtered, 2)
	for _, task := range filtered {
		assert.Equal(t, models.StatusTodo, task.Status)
	}

	filtered = filterByStatus(tasks, models.StatusInProgress)
	assert.Len(t, filtered, 1)
	assert.Equal(t, int64(2), filtered[0].ID)

	filtered = filterByStatus(tasks, models.StatusDone)
	assert.Len(t, filtered, 1)
	assert.Equal(t, int64(3), filtered[0].ID)
}

func TestFilterByStatusEmpty(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Name: "Task 1", Status: models.StatusTodo},
	}

	filtered := filterByStatus(tasks, models.StatusDone)
	assert.Empty(t, filtered)
}

func TestToTaskMap(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Name: "Task 1"},
		{ID: 2, Name: "Task 2"},
		{ID: 3, Name: "Task 3"},
	}

	m := toTaskMap(tasks)

	assert.Len(t, m, 3)
	assert.Equal(t, "Task 1", m[1].Name)
	assert.Equal(t, "Task 2", m[2].Name)
	assert.Equal(t, "Task 3", m[3].Name)
}

func TestCountByStatus(t *testing.T) {
	tasks := []models.Task{
		{ID: 1, Status: models.StatusTodo},
		{ID: 2, Status: models.StatusTodo},
		{ID: 3, Status: models.StatusInProgress},
		{ID: 4, Status: models.StatusDone},
		{ID: 5, Status: models.StatusDone},
		{ID: 6, Status: models.StatusDone},
	}

	assert.Equal(t, 2, countByStatus(tasks, models.StatusTodo))
	assert.Equal(t, 1, countByStatus(tasks, models.StatusInProgress))
	assert.Equal(t, 3, countByStatus(tasks, models.StatusDone))
}

func TestWatchInvalidStatus(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Set invalid status
	watchStatus = "invalid"
	watchPretty = false
	watchInterval = 100 * time.Millisecond

	err := runWatch(nil, []string{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
}
