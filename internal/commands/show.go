package commands

import (
	"fmt"
	"strconv"
	"strings"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show task details",
	Long:  `Display detailed information about a task including metadata and body content.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

func runShow(cmd *cobra.Command, args []string) error {
	// Parse task ID
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
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

	// Print task details
	printTaskDetails(task)

	return nil
}

func printTaskDetails(task *models.Task) {
	// Colors
	cyan := color.New(color.FgCyan, color.Bold)
	white := color.New(color.FgWhite)
	gray := color.New(color.FgHiBlack)

	// Print header
	separator := strings.Repeat("━", 60)
	cyan.Println(separator)
	cyan.Printf("Task: %d\n", task.ID)
	cyan.Println(separator)
	fmt.Println()

	// Print metadata
	white.Printf("Name:        %s\n", task.Name)

	if task.Description != "" {
		white.Printf("Description: %s\n", task.Description)
	}

	white.Printf("Status:      %s\n", task.Status)
	white.Printf("Priority:    %s\n", task.Priority)

	if task.Parent != nil {
		white.Printf("Parent:      %d\n", *task.Parent)
	} else {
		white.Println("Parent:      none")
	}

	if len(task.Tags) > 0 {
		white.Printf("Tags:        %s\n", strings.Join(task.Tags, ", "))
	}

	gray.Printf("Created:     %s\n", task.Created.Format("2006-01-02 15:04:05"))
	gray.Printf("Updated:     %s\n", task.Updated.Format("2006-01-02 15:04:05"))

	// Print body if present
	if task.Body != "" {
		fmt.Println()
		cyan.Println(separator)
		fmt.Println()
		fmt.Println(task.Body)
	}
}
