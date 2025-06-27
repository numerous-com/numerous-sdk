package login

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/output"
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to the Numerous CLI",
	Args:    cobra.NoArgs,
	GroupID: group.AdditionalCommandsGroupID,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use the standard tenant authenticator which now handles storage fallback
		authenticator := auth.NumerousTenantAuthenticator

		user := authenticator.GetLoggedInUserFromKeyring()
		if user != nil {
			output.PrintlnOK("Great, you are already logged in!")
			return nil
		}

		_, err := login(authenticator, cmd.Context())

		return errorhandling.ErrorAlreadyPrinted(err)
	},
}
