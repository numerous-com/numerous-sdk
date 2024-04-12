package organization

import (
	"os"

	"github.com/spf13/cobra"
)

var OrganizationRootCmd = &cobra.Command{
	Use: "organization",
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
