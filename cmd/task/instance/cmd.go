package instance

import (
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/task/instance/create"
	"numerous.com/cli/cmd/task/instance/logs"
	"numerous.com/cli/cmd/task/instance/stop"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "instance",
	Short: "Manage task instances",
	Args:  args.SubCommandRequired,
}

func init() {
	Cmd.AddCommand(create.Cmd)
	Cmd.AddCommand(logs.Cmd)
	Cmd.AddCommand(stop.Cmd)
}
