package config

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/config/set"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/config"
)

var Cmd = &cobra.Command{
	GroupID: group.AdditionalCommandsGroupID,
	Use:     "config",
	Short:   "Configure the Numerous CLI",
	Long:    "Set configuration values, or print the entire configuration.",
	RunE:    run,
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		// do nothing if subcommand is running
		return nil
	}

	cfg := config.Config{}
	if err := cfg.Load(); err != nil {
		output.PrintErrorDetails("Error loading configuration", err)
		return err
	}

	cfg.Print()

	return nil
}

func init() {
	Cmd.AddCommand(set.Cmd)
}
