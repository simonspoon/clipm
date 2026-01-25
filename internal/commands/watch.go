package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"time"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/models"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var (
	watchInterval time.Duration
	watchPretty   bool
	watchStatus   string
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch tasks for changes",
	Long:  `Continuously monitor tasks and display updates. Use Ctrl+C to exit.`,
	RunE:  runWatch,
}

func init() {
	watchCmd.Flags().DurationVar(&watchInterval, "interval", 500*time.Millisecond, "Polling interval")
	watchCmd.Flags().BoolVar(&watchPretty, "pretty", false, "Human-readable output (clear & redraw)")
	watchCmd.Flags().StringVar(&watchStatus, "status", "", "Filter by status (todo|in-progress|done)")
}

// WatchEvent represents a change event for JSON output
type WatchEvent struct {
	Type      string        `json:"type"`
	Task      *models.Task  `json:"task,omitempty"`
	Tasks     []models.Task `json:"tasks,omitempty"`
	TaskID    int64         `json:"taskId,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
}

func runWatch(cmd *cobra.Command, args []string) error {
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Validate status filter
	if watchStatus != "" && !models.IsValidStatus(watchStatus) {
		return fmt.Errorf("invalid status %q. Must be: todo, in-progress, done", watchStatus)
	}

	// Setup signal handling for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	ticker := time.NewTicker(watchInterval)
	defer ticker.Stop()

	var prevTasks map[int64]models.Task
	first := true

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			tasks, err := store.LoadAll()
			if err != nil {
				continue
			}

			// Filter by status if specified
			if watchStatus != "" {
				tasks = filterByStatus(tasks, watchStatus)
			}

			// Sort by created time
			sort.Slice(tasks, func(i, j int) bool {
				return tasks[i].Created.Before(tasks[j].Created)
			})

			currTasks := toTaskMap(tasks)

			if watchPretty {
				clearAndRender(tasks)
			} else {
				if first {
					outputSnapshot(tasks)
					first = false
				} else {
					outputChanges(prevTasks, currTasks)
				}
			}

			prevTasks = currTasks
		}
	}
}

func filterByStatus(tasks []models.Task, status string) []models.Task {
	var filtered []models.Task
	for i := range tasks {
		if tasks[i].Status == status {
			filtered = append(filtered, tasks[i])
		}
	}
	return filtered
}

func toTaskMap(tasks []models.Task) map[int64]models.Task {
	m := make(map[int64]models.Task)
	for i := range tasks {
		m[tasks[i].ID] = tasks[i]
	}
	return m
}

func detectChanges(prev, curr map[int64]models.Task) (added, updated, deleted []int64) {
	for id := range curr {
		task := curr[id]
		if _, exists := prev[id]; !exists {
			added = append(added, id)
		} else if !prev[id].Updated.Equal(task.Updated) {
			updated = append(updated, id)
		}
	}
	for id := range prev {
		if _, exists := curr[id]; !exists {
			deleted = append(deleted, id)
		}
	}
	return
}

func outputSnapshot(tasks []models.Task) {
	event := WatchEvent{
		Type:      "snapshot",
		Tasks:     tasks,
		Timestamp: time.Now(),
	}
	out, _ := json.Marshal(event)
	fmt.Println(string(out))
}

func outputChanges(prev, curr map[int64]models.Task) {
	added, updated, deleted := detectChanges(prev, curr)

	now := time.Now()

	for _, id := range added {
		task := curr[id]
		event := WatchEvent{
			Type:      "added",
			Task:      &task,
			Timestamp: now,
		}
		out, _ := json.Marshal(event)
		fmt.Println(string(out))
	}

	for _, id := range updated {
		task := curr[id]
		event := WatchEvent{
			Type:      "updated",
			Task:      &task,
			Timestamp: now,
		}
		out, _ := json.Marshal(event)
		fmt.Println(string(out))
	}

	for _, id := range deleted {
		event := WatchEvent{
			Type:      "deleted",
			TaskID:    id,
			Timestamp: now,
		}
		out, _ := json.Marshal(event)
		fmt.Println(string(out))
	}
}

func clearAndRender(tasks []models.Task) {
	// Clear screen using ANSI escape codes
	fmt.Print("\033[H\033[2J")

	// Header
	fmt.Printf("clipm watch - %s\n", time.Now().Format("15:04:05"))
	fmt.Printf("Tasks: %d todo, %d in-progress, %d done\n\n",
		countByStatus(tasks, models.StatusTodo),
		countByStatus(tasks, models.StatusInProgress),
		countByStatus(tasks, models.StatusDone))

	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}

	// Group by status
	grouped := make(map[string][]models.Task)
	for i := range tasks {
		grouped[tasks[i].Status] = append(grouped[tasks[i].Status], tasks[i])
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

		for i := range group {
			fmt.Printf("  %d  %s\n", group[i].ID, group[i].Name)
		}
	}
}

func countByStatus(tasks []models.Task, status string) int {
	count := 0
	for i := range tasks {
		if tasks[i].Status == status {
			count++
		}
	}
	return count
}
