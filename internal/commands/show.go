package commands

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var showPretty bool

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Long:  `Display detailed information about a task.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showPretty, "pretty", false, "Pretty print output")
}

func runShow(cmd *cobra.Command, args []string) error {
	// Normalize and validate task ID
	id := models.NormalizeTaskID(args[0])
	if !models.IsValidTaskID(id) {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Load task
	task, err := store.LoadTask(id)
	if err != nil {
		return err
	}

	if showPretty {
		printTaskDetails(task)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}

func printTaskDetails(task *models.Task) {
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite)
	gray := color.New(color.FgHiBlack)
	yellow := color.New(color.FgYellow)

	separator := strings.Repeat("-", 60)
	cyan.Println(separator)
	cyan.Printf("Task: %s\n", task.ID)
	cyan.Println(separator)
	fmt.Println()

	white.Printf("Name:        %s\n", task.Name)

	if task.Description != "" {
		white.Printf("Description: %s\n", task.Description)
	}

	white.Printf("Status:      %s\n", task.Status)

	if task.Parent != nil {
		white.Printf("Parent:      %s\n", *task.Parent)
	} else {
		white.Println("Parent:      none")
	}

	if task.Owner != nil {
		white.Printf("Owner:       %s\n", *task.Owner)
	}

	if len(task.BlockedBy) > 0 {
		white.Printf("Blocked by:  %v\n", task.BlockedBy)
	}

	gray.Printf("Created:     %s\n", task.Created.Format("2006-01-02 15:04:05"))
	gray.Printf("Updated:     %s\n", task.Updated.Format("2006-01-02 15:04:05"))

	if len(task.Notes) > 0 {
		fmt.Println()
		yellow.Println("Notes:")
		for _, note := range task.Notes {
			gray.Printf("  [%s] ", note.Timestamp.Format("2006-01-02 15:04"))
			white.Printf("%s\n", note.Content)
		}
	}
}
