package commands

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/simonspoon/clipm/internal/storage"
	"github.com/spf13/cobra"
)

var unblockPretty bool

var unblockCmd = &cobra.Command{
	Use:   "unblock <blocker-id> <blocked-id>",
	Short: "Remove a dependency between tasks",
	Long:  `Remove blocker-id from blocked-id's dependencies.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runUnblock,
}

func init() {
	unblockCmd.Flags().BoolVar(&unblockPretty, "pretty", false, "Pretty print output")
}

func runUnblock(cmd *cobra.Command, args []string) error {
	var blockerID, blockedID int64
	if _, err := fmt.Sscanf(args[0], "%d", &blockerID); err != nil {
		return fmt.Errorf("invalid blocker ID: %s", args[0])
	}
	if _, err := fmt.Sscanf(args[1], "%d", &blockedID); err != nil {
		return fmt.Errorf("invalid blocked ID: %s", args[1])
	}

	store, err := storage.NewStorage()
	if err != nil {
		return err
	}

	blocked, err := store.LoadTask(blockedID)
	if err != nil {
		if err == storage.ErrTaskNotFound {
			return fmt.Errorf("task %d not found", blockedID)
		}
		return err
	}

	// Find and remove blocker
	found := false
	newBlockedBy := make([]int64, 0, len(blocked.BlockedBy))
	for _, id := range blocked.BlockedBy {
		if id == blockerID {
			found = true
			continue
		}
		newBlockedBy = append(newBlockedBy, id)
	}

	if !found {
		return fmt.Errorf("task %d is not blocked by %d", blockedID, blockerID)
	}

	blocked.BlockedBy = newBlockedBy
	blocked.Updated = time.Now()

	if err := store.SaveTask(blocked); err != nil {
		return err
	}

	if unblockPretty {
		green := color.New(color.FgGreen)
		green.Printf("Task %d is no longer blocked by %d\n", blockedID, blockerID)
	} else {
		out, _ := json.Marshal(blocked)
		fmt.Println(string(out))
	}

	return nil
}
