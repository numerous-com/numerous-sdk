package task

import (
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/task/instances"
	"numerous.com/cli/cmd/task/list"
	"numerous.com/cli/cmd/task/logs"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "task",
	Short:   "Manage application tasks",
	Args:    args.SubCommandRequired,
	GroupID: group.AppCommandsGroupID,
}

func init() {
	Cmd.AddCommand(list.Cmd)
	Cmd.AddCommand(instances.Cmd)
	Cmd.AddCommand(logs.Cmd)
}
