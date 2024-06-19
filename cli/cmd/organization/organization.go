package organization

import (
	"numerous/cli/cmd/args"
	"numerous/cli/cmd/organization/create"
	"numerous/cli/cmd/organization/list"

	"github.com/spf13/cobra"
)

var OrganizationRootCmd = &cobra.Command{
	Use:  "organization",
	Args: args.SubCommandRequired,
}

func init() {
	OrganizationRootCmd.AddCommand(create.OrganizationCreateCmd)
	OrganizationRootCmd.AddCommand(list.OrganizationListCmd)
}
