package commands

import (
	"fmt"
	"sort"
	"strings"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	listStatus   string
	listPriority string
	listTag      string
	listParent   int64
	listNoParent bool
	listAll      bool
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks",
	Long:  `List tasks with optional filtering by status, priority, tag, or parent.`,
	RunE:  runList,
}

func init() {
	listCmd.Flags().StringVarP(&listStatus, "status", "s", "", "Filter by status (todo|in-progress|done|blocked)")
	listCmd.Flags().StringVarP(&listPriority, "priority", "p", "", "Filter by priority (low|medium|high)")
	listCmd.Flags().StringVarP(&listTag, "tag", "t", "", "Filter by tag")
	listCmd.Flags().Int64Var(&listParent, "parent", 0, "Show children of parent task")
	listCmd.Flags().BoolVar(&listNoParent, "no-parent", false, "Show only top-level tasks")
	listCmd.Flags().BoolVarP(&listAll, "all", "a", false, "Include archived tasks")
}

func runList(cmd *cobra.Command, args []string) error {
	// Load storage
	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	// Load index
	index, err := store.LoadIndex()
	if err != nil {
		return err
	}

	// Filter tasks
	tasks := filterTasks(index, store)

	// Group tasks by status
	grouped := groupByStatus(tasks)

	// Print tasks
	printTasks(grouped)

	return nil
}

func filterTasks(index *models.Index, store *storage.Storage) []*taskWithTags {
	var filtered []*taskWithTags

	for _, entry := range index.Tasks {
		// Skip archived unless --all
		if entry.Archived && !listAll {
			continue
		}

		// Filter by status
		if listStatus != "" && entry.Status != listStatus {
			continue
		}

		// Filter by priority
		if listPriority != "" && entry.Priority != listPriority {
			continue
		}

		// Filter by parent
		if listParent != 0 {
			if entry.Parent == nil || *entry.Parent != listParent {
				continue
			}
		}

		// Filter by no-parent
		if listNoParent && entry.Parent != nil {
			continue
		}

		// Load full task for tag filtering
		var tags []string
		if listTag != "" {
			task, err := store.LoadTask(entry.ID)
			if err != nil {
				continue
			}
			tags = task.Tags

			// Check if tag matches
			found := false
			for _, t := range tags {
				if t == listTag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		filtered = append(filtered, &taskWithTags{
			entry: entry,
			tags:  tags,
		})
	}

	return filtered
}

type taskWithTags struct {
	entry *models.IndexEntry
	tags  []string
}

func groupByStatus(tasks []*taskWithTags) map[string][]*taskWithTags {
	grouped := make(map[string][]*taskWithTags)

	for _, task := range tasks {
		status := task.entry.Status
		grouped[status] = append(grouped[status], task)
	}

	// Sort within each group by ID (oldest first)
	for _, group := range grouped {
		sort.Slice(group, func(i, j int) bool {
			return group[i].entry.ID < group[j].entry.ID
		})
	}

	return grouped
}

func printTasks(grouped map[string][]*taskWithTags) {
	// Status order
	statuses := []string{
		models.StatusTodo,
		models.StatusInProgress,
		models.StatusBlocked,
		models.StatusDone,
	}

	// Colors
	todoColor := color.New(color.FgWhite)
	inProgressColor := color.New(color.FgYellow)
	blockedColor := color.New(color.FgRed)
	doneColor := color.New(color.FgGreen)

	statusColors := map[string]*color.Color{
		models.StatusTodo:       todoColor,
		models.StatusInProgress: inProgressColor,
		models.StatusBlocked:    blockedColor,
		models.StatusDone:       doneColor,
	}

	totalCount := 0
	for _, status := range statuses {
		tasks := grouped[status]
		if len(tasks) == 0 {
			continue
		}

		totalCount += len(tasks)

		// Print status header
		statusLabel := strings.ToUpper(strings.ReplaceAll(status, "-", " "))
		statusColor := statusColors[status]
		statusColor.Printf("\n%s (%d)\n", statusLabel, len(tasks))

		// Print tasks
		for _, task := range tasks {
			printTask(task)
		}
	}

	if totalCount == 0 {
		fmt.Println("No tasks found.")
	}
}

func printTask(task *taskWithTags) {
	// Format priority
	priorityLabel := ""
	switch task.entry.Priority {
	case models.PriorityLow:
		priorityLabel = "LOW"
	case models.PriorityMedium:
		priorityLabel = "MED"
	case models.PriorityHigh:
		priorityLabel = "HIGH"
	}

	// Load tags if not already loaded
	tags := task.tags
	if len(tags) == 0 && listTag == "" {
		// Try to load from task file for display
		// For now, skip tag display if not filtered by tag
		// (would require loading all tasks, which is inefficient)
	}

	// Format tags
	tagStr := ""
	if len(tags) > 0 {
		tagParts := make([]string, len(tags))
		for i, tag := range tags {
			tagParts[i] = "#" + tag
		}
		tagStr = "  " + strings.Join(tagParts, " ")
	}

	// Print task line
	fmt.Printf("  %d  [%s]  %s%s\n",
		task.entry.ID,
		priorityLabel,
		task.entry.Name,
		tagStr,
	)
}
