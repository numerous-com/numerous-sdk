package create

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Creates an organization (login required)",
	Long:  "The organization feature is a structured space to keep the apps that you work on with team members. Organizations help to arrange your apps, manage who has access to them, and simplify the workflow with your team.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := organizationCreate(auth.NumerousTenantAuthenticator, gql.GetClient())
		if err != nil {
			output.PrintErrorDetails("Error creating organization", err)
		}

		return errorhandling.ErrorAlreadyPrinted(err)
	},
}
