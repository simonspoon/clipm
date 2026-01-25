package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var (
	claimPretty bool
	claimForce  bool
)

var claimCmd = &cobra.Command{
	Use:   "claim <id> <agent-name>",
	Short: "Claim ownership of a task",
	Long:  `Set the owner of a task to the specified agent name.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runClaim,
}

func init() {
	claimCmd.Flags().BoolVar(&claimPretty, "pretty", false, "Pretty print output")
	claimCmd.Flags().BoolVar(&claimForce, "force", false, "Force claim even if already owned")
}

func runClaim(cmd *cobra.Command, args []string) error {
	var id int64
	if _, err := fmt.Sscanf(args[0], "%d", &id); err != nil {
		return fmt.Errorf("invalid task ID: %s", args[0])
	}

	agentName := args[1]
	if agentName == "" {
		return fmt.Errorf("agent name cannot be empty")
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

	// Check if already owned by different agent
	if task.Owner != nil && *task.Owner != agentName && !claimForce {
		return fmt.Errorf("task %d is already owned by %s (use --force to override)", id, *task.Owner)
	}

	task.Owner = &agentName
	task.Updated = time.Now()

	if err := store.SaveTask(task); err != nil {
		return err
	}

	if claimPretty {
		green := color.New(color.FgGreen)
		green.Printf("Task %d claimed by %s\n", id, agentName)
	} else {
		out, _ := json.Marshal(task)
		fmt.Println(string(out))
	}

	return nil
}
