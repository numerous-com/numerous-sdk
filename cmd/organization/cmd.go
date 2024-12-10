package organization

import (
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/organization/create"
	"numerous.com/cli/cmd/organization/list"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "organization",
	Short:   "Manage Numerous organizations",
	Args:    args.SubCommandRequired,
	GroupID: group.AdditionalCommandsGroupID,
}

func init() {
	Cmd.AddCommand(create.Cmd)
	Cmd.AddCommand(list.Cmd)
}
