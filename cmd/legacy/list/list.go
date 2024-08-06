package list

import (
	"fmt"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your apps (login required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := list(auth.NumerousTenantAuthenticator, gql.GetClient())
		if err != nil {
			output.PrintUnknownError(err)
		}

		return err
	},
}

func list(a auth.Authenticator, c *gqlclient.Client) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		output.PrintErrorLogin()
		return nil
	}

	apps, err := app.QueryList(c)
	if err != nil {
		output.PrintUnknownError(err)
		return err
	}

	fmt.Println(setupTable(apps))

	return nil
}
