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
	Long:  `Returns the oldest task with status 'todo' (FIFO queue behavior).`,
	RunE:  runNext,
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
	task, err := store.GetNextTask()
	if err != nil {
		return err
	}

	if task == nil {
		if nextPretty {
			fmt.Println("No tasks in queue")
		} else {
			fmt.Println("null")
		}
		return nil
	}

	if nextPretty {
		cyan := color.New(color.FgCyan)
		cyan.Printf("Next task: %d - %s\n", task.ID, task.Name)
		if task.Description != "" {
			fmt.Printf("Description: %s\n", task.Description)
		}
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
