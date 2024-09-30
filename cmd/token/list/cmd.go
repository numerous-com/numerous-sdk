package list

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/token"
)

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List personal access tokens.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := List(cmd.Context(), token.NewService(gql.NewClient()))
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}
