package list

import (
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/gql/user"

	"git.sr.ht/~emersion/gqlclient"
)

func list(a auth.Authenticator, g *gqlclient.Client, displayMode DisplayMode) error {
	u := a.GetLoggedInUserFromKeyring()
	if u == nil {
		output.PrintErrorLogin()
		return nil
	}

	userResp, err := user.QueryUser(g)
	if err != nil {
		output.PrintErrorDetails("Error occurred querying user organization memberships", err)

		return err
	}

	configuredOrganization := config.OrganizationSlug()

	switch displayMode {
	case DisplayModeList:
		displayList(userResp.Memberships, configuredOrganization)
	case DisplayModeTable:
		displayTable(userResp.Memberships, configuredOrganization)
	default:
		output.PrintError("Unexpected display mode %q", "", displayMode)
		return errInvalidDisplayMode
	}

	return nil
}
