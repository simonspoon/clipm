package commands

import (
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var deletePretty bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Long:  `Delete a task. Cannot delete tasks that have undone children.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVar(&deletePretty, "pretty", false, "Pretty print output")
}

type deleteResult struct {
	Success bool  `json:"success"`
	ID      int64 `json:"id"`
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Load the task to verify it exists
	_, err = store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	// Check for undone children (recursive)
	hasUndone, err := store.HasUndoneChildren(id)
	if err != nil {
		return err
	}
	if hasUndone {
		return fmt.Errorf("cannot delete task: has undone children")
	}

	// Orphan any children before deleting
	if err := store.OrphanChildren(id); err != nil {
		return err
	}

	// Delete the task
	if err := store.DeleteTask(id); err != nil {
		return err
	}

	result := deleteResult{
		Success: true,
		ID:      id,
	}

	if deletePretty {
		green := color.New(color.FgGreen)
		green.Printf("Deleted task %d\n", id)
	} else {
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	}

	return nil
}
