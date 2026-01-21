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

var (
	addDescription string
	addParent      int64
	addPretty      bool
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new task",
	Long:  `Add a new task with the specified name and optional description.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Task description")
	addCmd.Flags().Int64Var(&addParent, "parent", 0, "Parent task ID")
	addCmd.Flags().BoolVar(&addPretty, "pretty", false, "Pretty print output")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Get task name
	name := args[0]

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Validate parent if specified
	var parent *int64
	if addParent != 0 {
		parentTask, err := store.LoadTask(addParent)
		if err != nil {
			return fmt.Errorf("parent task %d not found", addParent)
		}
		if parentTask.Status == models.StatusDone {
			return fmt.Errorf("cannot add child to done task")
		}
		parent = &addParent
	}

	// Create task
	now := time.Now()
	task := &models.Task{
		ID:          now.UnixMilli(),
		Name:        name,
		Description: addDescription,
		Parent:      parent,
		Status:      models.StatusTodo,
		Created:     now,
		Updated:     now,
	}

	// Save task
	if err := store.SaveTask(task); err != nil {
		return err
	}

	if addPretty {
		green := color.New(color.FgGreen)
		green.Printf("Created task %d: %s\n", task.ID, task.Name)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
