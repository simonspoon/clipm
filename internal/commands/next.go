package commands

import (
	"encoding/json"
	"fmt"

	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var nextPretty bool

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Get the next task to work on",
	Long: `Returns the next task using depth-first traversal.

When in-progress tasks exist: returns todo children (then siblings) of the deepest in-progress task, walking up the hierarchy as needed.
When no in-progress tasks: returns a list of root-level todo candidates.`,
	RunE: runNext,
}

func init() {
	nextCmd.Flags().BoolVar(&nextPretty, "pretty", false, "Pretty print output")
}

func runNext(cmd *cobra.Command, args []string) error {
	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Get next task
	result, err := store.GetNextTask()
	if err != nil {
		return err
	}

	// Handle single task result
	if result.Task != nil {
		if nextPretty {
			cyan := color.New(color.FgCyan)
			cyan.Printf("Next task: %d - %s\n", result.Task.ID, result.Task.Name)
			if result.Task.Description != "" {
				fmt.Printf("Description: %s\n", result.Task.Description)
			}
		} else {
			out, _ := json.Marshal(result)
			fmt.Println(string(out))
		}
		return nil
	}

	// Handle candidates list
	if len(result.Candidates) > 0 {
		if nextPretty {
			yellow := color.New(color.FgYellow)
			yellow.Println("No task in progress. Available candidates:")
			for i, t := range result.Candidates {
				fmt.Printf("  %d. %d - %s\n", i+1, t.ID, t.Name)
			}
		} else {
			out, _ := json.Marshal(result)
			fmt.Println(string(out))
		}
		return nil
	}

	// No tasks at all
	if nextPretty {
		fmt.Println("No tasks in queue")
	} else {
		out, _ := json.Marshal(result)
		fmt.Println(string(out))
	}
	return nil
}
