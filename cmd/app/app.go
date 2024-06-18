package app

import (
	"numerous.com/cli/cmd/app/deletecmd"
	"numerous.com/cli/cmd/app/deploy"
	"numerous.com/cli/cmd/app/logs"
	"numerous.com/cli/cmd/args"

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
