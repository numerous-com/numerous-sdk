package list

import (
	"fmt"
	"os"

	"numerous/cli/auth"
	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all your apps (login required)",
	Run: func(cmd *cobra.Command, args []string) {
		if err := list(auth.NumerousTenantAuthenticator, gql.GetClient()); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func list(a auth.Authenticator, c *gqlclient.Client) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		fmt.Printf("Command requires login.\n Use \"numerous login\" to login or sign up.\n")
		return nil
	}
	apps, err := app.QueryList(c)
	if err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println(setupTable(apps))

	return nil
}
