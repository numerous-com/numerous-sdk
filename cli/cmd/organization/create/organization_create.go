package create

import (
	"errors"
	"fmt"
	"os"

	"numerous/cli/auth"
	"numerous/cli/cmd/organization/create/wizard"
	"numerous/cli/cmd/output"
	"numerous/cli/internal/gql"

	"numerous/cli/internal/gql/organization"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var OrganizationCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates an organization (login required)",
	Long:  "The organization feature is a structured space to keep the apps that you work on with team members. Organizations help to arrange your apps, manage who has access to them, and simplify the workflow with your team.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := organizationCreate(auth.NumerousTenantAuthenticator, gql.GetClient()); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
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
			fmt.Println(err)
			return err
		default:
			newOrganization = _organization
			break wizard
		}
	}

	fmt.Println("\nThe organization has been created!")
	fmt.Println(newOrganization.String())

	return nil
}
