package list

import (
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/user"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var displayMode DisplayMode = DisplayModeList

var OrganizationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your organizations (login required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := list(auth.NumerousTenantAuthenticator, gql.GetClient())
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func init() {
	OrganizationListCmd.Flags().VarP(&displayMode, "display-mode", "d", "Display mode. Display organizations as a list or as a table. (\"list\", \"table\")")
}

func list(a auth.Authenticator, g *gqlclient.Client) error {
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
