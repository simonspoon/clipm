package commands

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a task in your editor",
	Long:  `Opens the task markdown file in your $EDITOR. Falls back to vim, vi, or nano if $EDITOR is not set.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEdit,
}

func runEdit(cmd *cobra.Command, args []string) error {
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

	// Check if task exists
	_, err = store.LoadTask(id)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", id)
		}
		return err
	}

	// Check if task is archived
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	entry, exists := index.GetTask(id)
	if !exists {
		return fmt.Errorf("task %d not found in index", id)
	}

	// Get the task file path
	taskPath := getTaskFilePath(store, id, entry.Archived)

	// Get editor from environment or use default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		// Try common editors in order of preference
		editors := []string{"vim", "vi", "nano"}
		for _, ed := range editors {
			if _, err := exec.LookPath(ed); err == nil {
				editor = ed
				break
			}
		}
	}

	if editor == "" {
		return fmt.Errorf("no editor found. Set $EDITOR environment variable or install vim, vi, or nano")
	}

	// Get file modification time before editing
	info, err := os.Stat(taskPath)
	if err != nil {
		return fmt.Errorf("failed to stat task file: %w", err)
	}
	beforeModTime := info.ModTime()

	// Open editor
	editCmd := exec.Command(editor, taskPath)
	editCmd.Stdin = os.Stdin
	editCmd.Stdout = os.Stdout
	editCmd.Stderr = os.Stderr

	if err := editCmd.Run(); err != nil {
		return fmt.Errorf("editor exited with error: %w", err)
	}

	// Check if file was modified
	info, err = os.Stat(taskPath)
	if err != nil {
		return fmt.Errorf("failed to stat task file: %w", err)
	}
	afterModTime := info.ModTime()

	if afterModTime.Equal(beforeModTime) {
		// File not modified, no need to reload
		yellow := color.New(color.FgYellow)
		yellow.Println("No changes made")
		return nil
	}

	// Reload the task to validate YAML
	reloaded, err := store.LoadTask(id)
	if err != nil {
		red := color.New(color.FgRed)
		red.Printf("Error: Failed to parse task file. YAML frontmatter may be invalid.\n")
		red.Printf("Fix the file at: %s\n", taskPath)
		return fmt.Errorf("invalid YAML frontmatter: %w", err)
	}

	// Update the timestamp
	reloaded.Updated = time.Now()

	// Save the task to update index
	if err := store.SaveTask(reloaded, entry.Archived); err != nil {
		return err
	}

	// Print success message
	green := color.New(color.FgGreen)
	green.Printf("✓ Updated task %d\n", id)

	return nil
}

// getTaskFilePath returns the file path for a task
func getTaskFilePath(store *storage.Storage, id int64, archived bool) string {
	return store.GetRootDir() + "/" + storage.ClipmDir + "/" + getTaskFileName(id, archived)
}

// getTaskFileName returns the filename for a task
func getTaskFileName(id int64, archived bool) string {
	filename := fmt.Sprintf("%s%d%s", storage.TaskFilePrefix, id, storage.TaskFileExt)
	if archived {
		return storage.ArchiveDir + "/" + filename
	}
	return filename
}
