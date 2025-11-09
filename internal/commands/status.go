package commands

import (
	"fmt"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Update task status",
	Long:  `Update the status of a task. Valid statuses: todo, in-progress, done, blocked`,
	Args:  cobra.ExactArgs(2),
	RunE:  runStatus,
}

func runStatus(cmd *cobra.Command, args []string) error {
	// Parse task ID
	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Get new status
	newStatus := args[1]

	// Validate status
	if !models.IsValidStatus(newStatus) {
		return fmt.Errorf("invalid status %q. Must be: todo, in-progress, done, blocked", newStatus)
	}

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Load the task
	task, err := store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	// Check if task is in index to determine if archived
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	entry, exists := index.GetTask(id)
	if !exists {
		return fmt.Errorf("task %d not found in index", id)
	}

	// Update status and timestamp
	task.Status = newStatus
	task.Updated = time.Now()

	// Save the task
	if err := store.SaveTask(task, entry.Archived); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	green.Printf("✓ Updated task %d status: %s\n", task.ID, newStatus)

	return nil
}
