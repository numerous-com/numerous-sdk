package app

import (
	"os"

	"numerous/cli/cmd/app/deploy"

	"github.com/spf13/cobra"
)

var AppRootCmd = &cobra.Command{
	Use: "app",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			if err := cmd.Help(); err != nil {
				return err
			}
			os.Exit(0)
		}

		return nil
	},
}

func init() {
	AppRootCmd.AddCommand(deploy.DeployCmd)
}
