package logout

import (
	"net/http"
	"os"

	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"

	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout of the Numerous CLI",
	GroupID: group.AdditionalCommandsGroupID,
	Run: func(cmd *cobra.Command, args []string) {
		if err := logout(auth.NumerousTenantAuthenticator); err != nil {
			output.PrintUnknownError(err)
			os.Exit(1)
		}
	},
}

func logout(a auth.Authenticator) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		output.PrintError("You are not logged in.", "")
		return nil
	}

	_ = a.RevokeRefreshToken(http.DefaultClient, user.RefreshToken)
	if err := a.RemoveLoggedInUserFromKeyring(); err != nil {
		output.PrintUnknownError(err)
		return err
	}
	output.PrintlnOK("You are now logged out.")

	return nil
}
