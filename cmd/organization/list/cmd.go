package list

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
)

var cmdArgs = struct {
	displayMode DisplayMode
}{
	displayMode: DisplayModeList,
}

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List all your organizations (login required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := list(auth.NumerousTenantAuthenticator, gql.GetClient(), cmdArgs.displayMode)
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.VarP(&cmdArgs.displayMode, "display-mode", "d", "Display mode. Display organizations as a list or as a table. (\"list\", \"table\")")
}
