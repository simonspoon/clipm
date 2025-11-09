package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDoneCommand(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	// Create a test task
	store, err := storage.NewStorage()
	require.NoError(t, err)

	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Test Task",
		Status:   models.StatusInProgress,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, false))

	// Verify task is in active directory
	activePath := filepath.Join(tmpDir, storage.ClipmDir, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, task.ID, storage.TaskFileExt))
	_, err = os.Stat(activePath)
	require.NoError(t, err)

	// Run done command
	err = runDone(nil, []string{fmt.Sprintf("%d", task.ID)})
	require.NoError(t, err)

	// Verify task is no longer in active directory
	_, err = os.Stat(activePath)
	assert.True(t, os.IsNotExist(err))

	// Verify task is in archive directory
	archivePath := filepath.Join(tmpDir, storage.ClipmDir, storage.ArchiveDir, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, task.ID, storage.TaskFileExt))
	_, err = os.Stat(archivePath)
	require.NoError(t, err)

	// Verify task status is done
	archived, err := store.LoadTask(task.ID)
	require.NoError(t, err)
	assert.Equal(t, models.StatusDone, archived.Status)

	// Verify index marks task as archived
	index, err := store.LoadIndex()
	require.NoError(t, err)
	entry, exists := index.GetTask(task.ID)
	require.True(t, exists)
	assert.True(t, entry.Archived)
}

func TestDoneCommand_TaskNotFound(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test non-existent task
	err := runDone(nil, []string{"999999999999"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDoneCommand_InvalidID(t *testing.T) {
	_, cleanup := setupTestEnv(t)
	defer cleanup()

	// Test invalid ID format
	err := runDone(nil, []string{"not-a-number"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid task ID")
}

func TestDoneCommand_AlreadyArchived(t *testing.T) {
	tmpDir, cleanup := setupTestEnv(t)
	defer cleanup()

	store, err := storage.NewStorage()
	require.NoError(t, err)

	// Create a task that's already archived
	now := time.Now()
	task := &models.Task{
		ID:       now.UnixMilli(),
		Name:     "Archived Task",
		Status:   models.StatusDone,
		Priority: models.PriorityMedium,
		Created:  now,
		Updated:  now,
	}
	require.NoError(t, store.SaveTask(task, true))

	// Verify task is in archive
	archivePath := filepath.Join(tmpDir, storage.ClipmDir, storage.ArchiveDir, fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, task.ID, storage.TaskFileExt))
	_, err = os.Stat(archivePath)
	require.NoError(t, err)

	// Should not be able to archive again (task is already in archive, not in active)
	err = runDone(nil, []string{fmt.Sprintf("%d", task.ID)})
	assert.Error(t, err)
}
