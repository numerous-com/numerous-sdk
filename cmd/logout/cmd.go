package logout

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/internal/auth"
)

var Cmd = &cobra.Command{
	Use:     "logout",
	Short:   "Logout of the Numerous CLI",
	GroupID: group.AdditionalCommandsGroupID,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Use the fallback authenticator for logout
		err := logout(auth.NumerousTenantAuthenticator)
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}
