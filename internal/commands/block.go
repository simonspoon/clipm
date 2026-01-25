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

var blockPretty bool

var blockCmd = &cobra.Command{
	Use:   "block <blocker-id> <blocked-id>",
	Short: "Add a dependency between tasks",
	Long:  `Make blocked-id depend on blocker-id. The blocked task cannot be worked on until the blocker is done.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runBlock,
}

func init() {
	blockCmd.Flags().BoolVar(&blockPretty, "pretty", false, "Pretty print output")
}

func runBlock(cmd *cobra.Command, args []string) error {
	blockerID, blockedID, err := parseBlockArgs(args)
	if err != nil {
		return err
	}

	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	blocker, blocked, err := loadBlockTasks(store, blockerID, blockedID)
	if err != nil {
		return err
	}

	if err := validateBlock(store, blocker, blocked, blockerID, blockedID); err != nil {
		return err
	}

	blocked.BlockedBy = append(blocked.BlockedBy, blockerID)
	blocked.Updated = time.Now()

	if err := store.SaveTask(blocked); err != nil {
		return err
	}

	if blockPretty {
		green := color.New(color.FgGreen)
		green.Printf("Task %d is now blocked by %d\n", blockedID, blockerID)
	} else {
		out, _ := json.Marshal(blocked)
		fmt.Println(string(out))
	}

	return nil
}

func parseBlockArgs(args []string) (blockerID, blockedID int64, err error) {
	if _, err := fmt.Sscanf(args[0], "%d", &blockerID); err != nil {
		return 0, 0, fmt.Errorf("invalid blocker ID: %s", args[0])
	}
	if _, err := fmt.Sscanf(args[1], "%d", &blockedID); err != nil {
		return 0, 0, fmt.Errorf("invalid blocked ID: %s", args[1])
	}
	if blockerID == blockedID {
		return 0, 0, fmt.Errorf("a task cannot block itself")
	}
	return blockerID, blockedID, nil
}

func loadBlockTasks(store *storage.Storage, blockerID, blockedID int64) (*models.Task, *models.Task, error) {
	blocker, err := store.LoadTask(blockerID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return nil, nil, fmt.Errorf("blocker task %d not found", blockerID)
		}
		return nil, nil, err
	}

	blocked, err := store.LoadTask(blockedID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return nil, nil, fmt.Errorf("blocked task %d not found", blockedID)
		}
		return nil, nil, err
	}
	return blocker, blocked, nil
}

func validateBlock(store *storage.Storage, blocker, blocked *models.Task, blockerID, blockedID int64) error {
	if blocker.Status == models.StatusDone {
		return fmt.Errorf("cannot block on completed task %d", blockerID)
	}

	hasCycle, err := store.WouldCreateCycle(blockerID, blockedID)
	if err != nil {
		return err
	}
	if hasCycle {
		return fmt.Errorf("cannot add dependency: would create a cycle")
	}

	for _, id := range blocked.BlockedBy {
		if id == blockerID {
			return fmt.Errorf("task %d is already blocked by %d", blockedID, blockerID)
		}
	}
	return nil
}
