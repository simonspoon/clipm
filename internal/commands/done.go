package commands

import (
	"fmt"

	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done <id>",
	Short: "Mark task as done and archive it",
	Long:  `Mark a task as done and move it to the archive directory.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDone,
}

func runDone(cmd *cobra.Command, args []string) error {
	// Parse task ID
	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Check if task exists
	task, err := store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	// Archive the task (this also sets status to done)
	if err := store.ArchiveTask(id); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	green.Printf("✓ Completed and archived task %d: %s\n", task.ID, task.Name)

	return nil
}
