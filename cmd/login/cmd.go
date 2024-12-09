package login

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
)

var Cmd = &cobra.Command{
	Use:     "login",
	Short:   "Login to the Numerous CLI",
	Args:    cobra.NoArgs,
	GroupID: group.AdditionalCommandsGroupID,
	RunE: func(cmd *cobra.Command, args []string) error {
		user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
		if user != nil {
			output.PrintlnOK("Great, you are already logged in!")
			return nil
		}

		_, err := login(auth.NumerousTenantAuthenticator, cmd.Context())

		return errorhandling.ErrorAlreadyPrinted(err)
	},
}
