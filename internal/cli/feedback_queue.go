package cli

import (
	"fmt"
	"os"

	"github.com/dinacomputer/cli/internal/feedback"
	"github.com/spf13/cobra"
)

var feedbackQueueCmd = &cobra.Command{
	Use:   "queue",
	Short: "Manage the local feedback retry queue",
	Long: `Manage the local queue of feedback submissions that couldn't be sent.

Items are stored under ~/.config/dina/feedback-queue/. Use ` + "`dina doctor`" + `
to inspect the queue and ` + "`dina doctor --fix`" + ` to retry pending items.`,
}

var feedbackQueueClearForce bool

var feedbackQueueClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Discard every pending feedback submission",
	Long: `Delete every item in the local feedback retry queue. This is irreversible —
items removed here are gone, even if they were never delivered to the server.

Prompts for confirmation by default. Pass --force to skip the prompt.`,
	Example: `  # interactive
  dina feedback queue clear

  # non-interactive (CI / scripts)
  dina feedback queue clear --force`,
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := feedback.List()
		if err != nil {
			return err
		}
		if len(items) == 0 {
			Infoln("Feedback queue is already empty.")
			return nil
		}

		prompt := fmt.Sprintf("Discard %d pending feedback submission(s)?", len(items))
		if err := confirmYesNo(prompt, feedbackQueueClearForce); err != nil {
			return err
		}

		removed, err := feedback.Clear()
		if err != nil {
			return fmt.Errorf("removed %d item(s) before failure: %w", removed, err)
		}
		fmt.Fprintf(os.Stderr, "Cleared %d item(s) from the feedback queue.\n", removed)
		return nil
	},
}

func init() {
	feedbackQueueClearCmd.Flags().BoolVarP(&feedbackQueueClearForce, "force", "f", false, "Skip confirmation prompt")
	feedbackQueueCmd.AddCommand(feedbackQueueClearCmd)
	feedbackCmd.AddCommand(feedbackQueueCmd)
}
