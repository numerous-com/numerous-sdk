package list

import (
	"fmt"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/user"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var OrganizationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your organizations (login required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := list(auth.NumerousTenantAuthenticator, gql.GetClient())

		return err
	},
}

func list(a auth.Authenticator, g *gqlclient.Client) error {
	u := a.GetLoggedInUserFromKeyring()
	if u == nil {
		output.PrintErrorLogin()
		return nil
	}

	userResp, err := user.QueryUser(g)
	if err != nil {
		if err != nil {
			output.PrintErrorDetails("Error occurred querying user organization memberships", err)
		}

		return err
	}

	fmt.Println(setupTable(userResp.Memberships))

	return nil
}
