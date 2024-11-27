package app

import (
	"numerous.com/cli/cmd/app/list"
	"numerous.com/cli/cmd/app/share"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "app",
	Short:   "Manage Numerous applications",
	Args:    args.SubCommandRequired,
	GroupID: group.AdditionalCommandsGroupID,
}

func init() {
	Cmd.AddCommand(list.Cmd)
	Cmd.AddCommand(share.Cmd)
}
