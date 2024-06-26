package organization

import (
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/organization/create"
	"numerous.com/cli/cmd/organization/list"

	"github.com/spf13/cobra"
)

var OrganizationRootCmd = &cobra.Command{
	Use:     "organization",
	Short:   "Manage Numerous organizations",
	Args:    args.SubCommandRequired,
	GroupID: "additional-cmds",
}

func init() {
	OrganizationRootCmd.AddCommand(create.OrganizationCreateCmd)
	OrganizationRootCmd.AddCommand(list.OrganizationListCmd)
}
