package commands

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	addDescription string
	addPriority    string
	addTags        string
	addParent      int64
	addBody        string
)

var addCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add a new task",
	Long:  `Add a new task with the specified name and optional metadata.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

func init() {
	addCmd.Flags().StringVarP(&addDescription, "description", "d", "", "Task description")
	addCmd.Flags().StringVarP(&addPriority, "priority", "p", models.PriorityMedium, "Priority level (low|medium|high)")
	addCmd.Flags().StringVarP(&addTags, "tags", "t", "", "Comma-separated tags")
	addCmd.Flags().Int64Var(&addParent, "parent", 0, "Parent task ID")
	addCmd.Flags().StringVarP(&addBody, "body", "b", "", "Task body content (markdown)")
}

func runAdd(cmd *cobra.Command, args []string) error {
	// Get task name
	name := args[0]

	// Validate priority
	if !models.IsValidPriority(addPriority) {
		return fmt.Errorf("invalid priority %q. Must be: low, medium, high", addPriority)
	}

	// Parse tags
	var tags []string
	if addTags != "" {
		for _, tag := range strings.Split(addTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				tags = append(tags, tag)
			}
		}
	}

	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Validate parent if specified
	var parent *int64
	if addParent != 0 {
		index, err := store.LoadIndex()
		if err != nil {
			return err
		}

		entry, exists := index.GetTask(addParent)
		if !exists {
			return fmt.Errorf("parent task %d not found", addParent)
		}
		if entry.Archived {
			return fmt.Errorf("cannot set archived task as parent")
		}

		parent = &addParent
	}

	// Determine body content (--body flag takes precedence over stdin)
	body := addBody
	if body == "" {
		// Check if stdin is piped
		fileInfo, err := os.Stdin.Stat()
		if err == nil && (fileInfo.Mode()&os.ModeCharDevice) == 0 {
			// stdin is piped, read it
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read stdin: %w", err)
			}
			body = strings.TrimSpace(string(data))
		}
	}

	// Create task
	now := time.Now()
	task := &models.Task{
		ID:          now.UnixMilli(),
		Name:        name,
		Description: addDescription,
		Parent:      parent,
		Status:      models.StatusTodo,
		Priority:    addPriority,
		Created:     now,
		Updated:     now,
		Tags:        tags,
		Body:        body,
	}

	// Save task
	if err := store.SaveTask(task, false); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	green.Printf("✓ Created task %d: %s\n", task.ID, task.Name)

	return nil
}
