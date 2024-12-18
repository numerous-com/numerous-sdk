package create

import (
	"errors"
	"fmt"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/output"

	"numerous.com/cli/internal/gql/organization"

	"git.sr.ht/~emersion/gqlclient"
)

func organizationCreate(a auth.Authenticator, c *gqlclient.Client) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		output.PrintErrorLogin()
		return nil
	}

	newOrganization := organization.Organization{}

wizard:
	for {
		if err := runWizard(&newOrganization.Name, *user); err != nil {
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
