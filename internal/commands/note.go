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

var notePretty bool

var noteCmd = &cobra.Command{
	Use:   "note <id> <message>",
	Short: "Add a note to a task",
	Long:  `Append an observation or progress update to a task.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runNote,
}

func init() {
	noteCmd.Flags().BoolVar(&notePretty, "pretty", false, "Pretty print output")
}

func runNote(cmd *cobra.Command, args []string) error {
	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	message := args[1]
	if message == "" {
		return fmt.Errorf("note message cannot be empty")
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

	note := models.Note{
		Content:   message,
		Timestamp: time.Now(),
	}

	task.Notes = append(task.Notes, note)
	task.Updated = time.Now()

	if err := store.SaveTask(task); err != nil {
		return err
	}

	if notePretty {
		green := color.New(color.FgGreen)
		green.Printf("Added note to task %d\n", id)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
