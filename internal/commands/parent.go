package commands

import (
	"fmt"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var parentCmd = &cobra.Command{
	Use:   "parent <id> <parent-id>",
	Short: "Set a task's parent",
	Long:  `Set the parent of a task to create a hierarchical relationship. Prevents circular dependencies.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runParent,
}

func runParent(cmd *cobra.Command, args []string) error {
	// Parse task IDs
	var childID, parentID int64
	if _, err := fmt.Sscanf(args[0], "%d", &childID); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}
	if _, err := fmt.Sscanf(args[1], "%d", &parentID); err != nil {
		return fmt.Errorf("invalid parent ID: %s", args[1])
	}

	// Can't parent to self
	if childID == parentID {
		return fmt.Errorf("cannot set task as its own parent")
	}

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Load index
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	// Check child task exists
	childEntry, exists := index.GetTask(childID)
	if !exists {
		return fmt.Errorf("task %d not found", childID)
	}

	// Check parent task exists
	parentEntry, exists := index.GetTask(parentID)
	if !exists {
		return fmt.Errorf("parent task %d not found", parentID)
	}

	// Check parent is not archived
	if parentEntry.Archived {
		return fmt.Errorf("cannot set archived task %d as parent", parentID)
	}

	// Check for circular dependencies
	if wouldCreateCycle(index, childID, parentID) {
		return fmt.Errorf("cannot set parent - would create circular dependency")
	}

	// Load the child task
	childTask, err := store.LoadTask(childID)
	if err != nil {
		return err
	}

	// Update parent and timestamp
	childTask.Parent = &parentID
	childTask.Updated = time.Now()

	// Save the task
	if err := store.SaveTask(childTask, childEntry.Archived); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	green.Printf("✓ Task %d is now a child of %d\n", childID, parentID)

	return nil
}

// wouldCreateCycle checks if setting parentID as the parent of childID would create a cycle
func wouldCreateCycle(index *models.Index, childID, parentID int64) bool {
	// Traverse up the parent chain from the proposed parent
	// If we encounter childID, we have a cycle
	currentID := parentID
	visited := make(map[int64]bool)

	for {
		// Detect loops in existing structure
		if visited[currentID] {
			return true
		}
		visited[currentID] = true

		// If we reached the child, we have a cycle
		if currentID == childID {
			return true
		}

		// Get the current task's parent
		entry, exists := index.GetTask(currentID)
		if !exists || entry.Parent == nil {
			// Reached a top-level task, no cycle
			return false
		}

		// Move up to parent
		currentID = *entry.Parent
	}
}
