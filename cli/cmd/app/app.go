package app

import (
	"numerous/cli/cmd/app/deletecmd"
	"numerous/cli/cmd/app/deploy"
	"numerous/cli/cmd/app/logs"
	"numerous/cli/cmd/args"

	"github.com/spf13/cobra"
)

var AppRootCmd = &cobra.Command{
	Use:  "app",
	Args: args.SubCommandRequired,
}

func init() {
	AppRootCmd.AddCommand(deletecmd.DeleteCmd)
	AppRootCmd.AddCommand(deploy.DeployCmd)
	AppRootCmd.AddCommand(logs.LogsCmd)
}
