package commands

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var (
	listStatus string
	listPretty bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks with optional filtering by status.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (todo|in-progress|done)")
	listCmd.Flags().BoolVar(&listPretty, "pretty", false, "Pretty print output")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Load all tasks
	tasks, err := store.LoadAll()
	if err != nil {
		return err
	}

	// Filter by status if specified
	if listStatus != "" {
		if !models.IsValidStatus(listStatus) {
			return fmt.Errorf("invalid status %q. Must be: todo, in-progress, done", listStatus)
		}
		var filtered []models.Task
		for _, t := range tasks {
			if t.Status == listStatus {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

	// Sort by created time (oldest first)
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Created.Before(tasks[j].Created)
	})

	if listPretty {
		printTasksPretty(tasks)
	} else {
		out, _ := json.Marshal(tasks)
		fmt.Println(string(out))
	}

	return nil
}

func printTasksPretty(tasks []models.Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	// Group by status
	grouped := make(map[string][]models.Task)
	for _, t := range tasks {
		grouped[t.Status] = append(grouped[t.Status], t)
	}

	// Status order
	statuses := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusDone,
	}

	// Colors
	statusColors := map[string]*color.Color{
		models.StatusTodo:       color.New(color.FgWhite),
		models.StatusInProgress: color.New(color.FgYellow),
		models.StatusDone:       color.New(color.FgGreen),
	}

	for _, status := range statuses {
		group := grouped[status]
		if len(group) == 0 {
			continue
		}

		statusColor := statusColors[status]
		statusColor.Printf("\n%s (%d)\n", status, len(group))

		for _, t := range group {
			fmt.Printf("  %d  %s\n", t.ID, t.Name)
		}
	}
}
