package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var statusPretty bool

var statusCmd = &cobra.Command{
	Use:   "status <id> <status>",
	Short: "Update task status",
	Long:  `Update the status of a task. Valid statuses: todo, in-progress, done`,
	Args:  cobra.ExactArgs(2),
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().BoolVar(&statusPretty, "pretty", false, "Pretty print output")
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
		return fmt.Errorf("invalid status %q. Must be: todo, in-progress, done", newStatus)
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

	// If marking as done, check that all children are done
	if newStatus == models.StatusDone {
		hasUndone, err := store.HasUndoneChildren(id)
		if err != nil {
			return err
		}
		if hasUndone {
			return fmt.Errorf("cannot mark task as done: has undone children")
		}
	}

	// Update status and timestamp
	task.Status = newStatus
	task.Updated = time.Now()

	// Save the task
	if err := store.SaveTask(task); err != nil {
		return err
	}

	if statusPretty {
		green := color.New(color.FgGreen)
		green.Printf("Updated task %d status: %s\n", task.ID, newStatus)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
