package list

import (
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		if err := list(auth.NumerousTenantAuthenticator, gql.GetClient()); err != nil {
			output.PrintUnknownError(err)
			os.Exit(1)
		}
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
