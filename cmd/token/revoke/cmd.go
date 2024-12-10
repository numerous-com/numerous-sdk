package revoke

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/token"
)

var cmdArgs struct{ id string }

var Cmd = &cobra.Command{
	Use:   "revoke",
	Short: "Revoke a personal access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Revoke(cmd.Context(), token.NewService(gql.NewClient()), cmdArgs.id)
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&cmdArgs.id, "id", "", "", "The id of the personal access token.")
}
