package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var unclaimPretty bool

var unclaimCmd = &cobra.Command{
	Use:   "unclaim <id>",
	Short: "Remove ownership from a task",
	Long:  `Clear the owner of a task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runUnclaim,
}

func init() {
	unclaimCmd.Flags().BoolVar(&unclaimPretty, "pretty", false, "Pretty print output")
}

func runUnclaim(cmd *cobra.Command, args []string) error {
	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	task, err := store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	if task.Owner == nil {
		return fmt.Errorf("task %d has no owner", id)
	}

	task.Owner = nil
	task.Updated = time.Now()

	if err := store.SaveTask(task); err != nil {
		return err
	}

	if unclaimPretty {
		green := color.New(color.FgGreen)
		green.Printf("Task %d ownership cleared\n", id)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
