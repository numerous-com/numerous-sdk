package task

import (
	"github.com/spf13/cobra"
)

// TaskCmd represents the task command
var TaskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage and run task collections",
	Long: `Task collection management commands for Numerous.

This command group allows you to run, validate, and manage task collections locally
before deploying them to the Numerous platform.`,
}

func init() {
	// Add subcommands
	TaskCmd.AddCommand(runCmd)
	TaskCmd.AddCommand(validateCmd)
	TaskCmd.AddCommand(listCmd)
}
