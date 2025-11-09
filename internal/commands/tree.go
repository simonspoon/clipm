package commands

import (
	"fmt"
	"sort"

	"clipm/internal/models"
	"clipm/internal/storage"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	treeAll bool
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display tasks in a hierarchical tree view",
	Long:  `Display all tasks in a hierarchical tree structure showing parent-child relationships.`,
	RunE:  runTree,
}

func init() {
	treeCmd.Flags().BoolVarP(&treeAll, "all", "a", false, "Include archived tasks")
}

func runTree(cmd *cobra.Command, args []string) error {
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
	var tasks []*models.IndexEntry
	for _, entry := range index.Tasks {
		if !treeAll && entry.Archived {
			continue
		}
		tasks = append(tasks, entry)
	}

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	// Sort tasks by creation time
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Created.Before(tasks[j].Created)
	})

	// Build task map for easy lookup
	taskMap := make(map[int64]*models.IndexEntry)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	// Find root tasks (tasks with no parent)
	var roots []*models.IndexEntry
	for _, task := range tasks {
		if task.Parent == nil {
			roots = append(roots, task)
		}
	}

	// Print tree for each root
	for i, root := range roots {
		isLast := i == len(roots)-1
		printTaskTree(root, taskMap, "", isLast)
	}

	return nil
}

func printTaskTree(task *models.IndexEntry, taskMap map[int64]*models.IndexEntry, prefix string, isLast bool) {
	// Color setup
	boldWhite := color.New(color.Bold, color.FgWhite)
	gray := color.New(color.FgHiBlack)
	statusColor := getStatusColor(task.Status)
	priorityColor := getPriorityColor(task.Priority)

	// Print current task
	var marker string
	if prefix == "" {
		marker = ""
	} else if isLast {
		marker = "└─ "
	} else {
		marker = "├─ "
	}

	// Format: ID  Name  [STATUS]  [PRIORITY]
	fmt.Print(prefix + marker)
	gray.Printf("%d  ", task.ID)
	boldWhite.Print(task.Name)
	fmt.Print("  ")
	statusColor.Printf("[%s]", formatStatus(task.Status))
	fmt.Print("  ")
	priorityColor.Printf("[%s]", formatPriority(task.Priority))
	if task.Archived {
		gray.Print(" (archived)")
	}
	fmt.Println()

	// Find children
	var children []*models.IndexEntry
	for _, t := range taskMap {
		if t.Parent != nil && *t.Parent == task.ID {
			children = append(children, t)
		}
	}

	// Sort children by creation time
	sort.Slice(children, func(i, j int) bool {
		return children[i].Created.Before(children[j].Created)
	})

	// Print children recursively
	for i, child := range children {
		childIsLast := i == len(children)-1
		var childPrefix string
		if prefix == "" {
			childPrefix = "  "
		} else if isLast {
			childPrefix = prefix + "   "
		} else {
			childPrefix = prefix + "│  "
		}
		printTaskTree(child, taskMap, childPrefix, childIsLast)
	}
}

func getStatusColor(status string) *color.Color {
	switch status {
	case models.StatusTodo:
		return color.New(color.FgCyan)
	case models.StatusInProgress:
		return color.New(color.FgYellow)
	case models.StatusDone:
		return color.New(color.FgGreen)
	case models.StatusBlocked:
		return color.New(color.FgRed)
	default:
		return color.New(color.FgWhite)
	}
}

func getPriorityColor(priority string) *color.Color {
	switch priority {
	case models.PriorityHigh:
		return color.New(color.FgRed)
	case models.PriorityMedium:
		return color.New(color.FgYellow)
	case models.PriorityLow:
		return color.New(color.FgBlue)
	default:
		return color.New(color.FgWhite)
	}
}

func formatStatus(status string) string {
	switch status {
	case models.StatusInProgress:
		return "IN-PROG"
	case models.StatusBlocked:
		return "BLOCKED"
	case models.StatusDone:
		return "DONE"
	case models.StatusTodo:
		return "TODO"
	default:
		return status
	}
}

func formatPriority(priority string) string {
	switch priority {
	case models.PriorityHigh:
		return "HIGH"
	case models.PriorityMedium:
		return "MED"
	case models.PriorityLow:
		return "LOW"
	default:
		return priority
	}
}
