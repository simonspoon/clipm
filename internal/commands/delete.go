package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	deleteForce      bool
	deleteOrphanKids bool
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a task",
	Long:  `Delete a task. If the task has children, you'll be prompted to delete them or orphan them.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runDelete,
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")
	deleteCmd.Flags().BoolVar(&deleteOrphanKids, "orphan", false, "Orphan child tasks instead of deleting them")
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

	// Load the task
	task, err := store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	// Load index to check for children
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	// Find children
	children := index.GetChildren(id)

	// Prompt for confirmation if not forced
	if !deleteForce {
		reader := bufio.NewReader(os.Stdin)

		if len(children) > 0 {
			fmt.Printf("Delete task %d: %q?\n", id, task.Name)
			fmt.Printf("This task has %d child task(s). Delete children too? [y/N/orphan]: ", len(children))

			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))

			if response == "orphan" {
				deleteOrphanKids = true
			} else if response != "y" && response != "yes" {
				return fmt.Errorf("deletion cancelled")
			}
		} else {
			fmt.Printf("Delete task %d: %q? [y/N]: ", id, task.Name)

			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))

			if response != "y" && response != "yes" {
				return fmt.Errorf("deletion cancelled")
			}
		}
	}

	// Handle children
	if len(children) > 0 {
		if deleteOrphanKids {
			// Orphan children by setting their parent to nil
			for _, childID := range children {
				child, err := store.LoadTask(childID)
				if err != nil {
					continue
				}

				entry, exists := index.GetTask(childID)
				if !exists {
					continue
				}

				child.Parent = nil
				if err := store.SaveTask(child, entry.Archived); err != nil {
					return fmt.Errorf("failed to orphan child task %d: %w", childID, err)
				}
			}
		} else {
			// Delete all children recursively
			for _, childID := range children {
				if err := deleteTaskRecursive(store, childID); err != nil {
					return fmt.Errorf("failed to delete child task %d: %w", childID, err)
				}
			}
		}
	}

	// Delete the task
	if err := store.DeleteTask(id); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	if len(children) > 0 && !deleteOrphanKids {
		green.Printf("✓ Deleted task %d and %d child task(s)\n", id, len(children))
	} else if len(children) > 0 && deleteOrphanKids {
		green.Printf("✓ Deleted task %d and orphaned %d child task(s)\n", id, len(children))
	} else {
		green.Printf("✓ Deleted task %d\n", id)
	}

	return nil
}

// deleteTaskRecursive recursively deletes a task and all its children
func deleteTaskRecursive(store *storage.Storage, id int64) error {
	// Load index to find children
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	// Find and delete children first
	children := index.GetChildren(id)
	for _, childID := range children {
		if err := deleteTaskRecursive(store, childID); err != nil {
			return err
		}
	}

	// Delete the task itself
	return store.DeleteTask(id)
}
