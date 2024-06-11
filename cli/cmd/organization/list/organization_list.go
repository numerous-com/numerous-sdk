package list

import (
	"fmt"
	"os"

	"numerous/cli/auth"
	"numerous/cli/cmd/output"
	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/user"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var OrganizationListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your organizations (login required)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := list(auth.NumerousTenantAuthenticator, gql.GetClient()); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
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
		fmt.Println(err)
		return err
	}

	fmt.Println(setupTable(userResp.Memberships))

	return nil
}
