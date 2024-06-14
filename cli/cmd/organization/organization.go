package organization

import (
	"os"

	"numerous/cli/cmd/organization/create"
	"numerous/cli/cmd/organization/list"

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

func init() {
	OrganizationRootCmd.AddCommand(create.OrganizationCreateCmd)
	OrganizationRootCmd.AddCommand(list.OrganizationListCmd)
}
