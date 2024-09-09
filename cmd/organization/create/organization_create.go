package create

import (
	"errors"
	"fmt"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/organization/create/wizard"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"

	"numerous.com/cli/internal/gql/organization"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var OrganizationCreateCmd = &cobra.Command{
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

func organizationCreate(a auth.Authenticator, c *gqlclient.Client) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		output.PrintErrorLogin()
		return nil
	}

	newOrganization := organization.Organization{}

wizard:
	for {
		if err := wizard.RunOrganizationCreateWizard(&newOrganization.Name, *user); err != nil {
			return err
		}

		_organization, err := organization.Create(newOrganization.Name, c)

		switch {
		case errors.Is(err, organization.ErrOrganizationNameInvalidCharacter):
			fmt.Println("The input name contains invalid characters. Please choose another name or press ctrl + c to quit.")
		case err != nil:
			return err
		default:
			newOrganization = _organization
			break wizard
		}
	}

	output.PrintlnOK("The organization has been created:")
	fmt.Println(newOrganization.String())

	return nil
}
