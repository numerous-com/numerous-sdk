package logout

import (
	"net/http"

	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"

	"github.com/spf13/cobra"
)

var LogoutCmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout of the Numerous CLI",
	GroupID: group.AdditionalCommandsGroupID,
	RunE:    run,
}

func run(cmd *cobra.Command, args []string) error {
	return logout(auth.NumerousTenantAuthenticator)
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
