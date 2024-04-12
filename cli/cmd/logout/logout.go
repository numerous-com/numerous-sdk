package logout

import (
	"fmt"
	"net/http"
	"os"

	"numerous/cli/auth"

	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of the Numerous CLI",
	Run: func(cmd *cobra.Command, args []string) {
		if err := logout(auth.NumerousTenantAuthenticator); err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func logout(a auth.Authenticator) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		fmt.Println("You are not logged in.")
		return nil
	}

	_ = a.RevokeRefreshToken(http.DefaultClient, user.RefreshToken)
	if err := a.RemoveLoggedInUserFromKeyring(); err != nil {
		fmt.Println("A problem occurred when signing out")
		return err
	}
	fmt.Println("Successfully logged out!")

	return nil
}
