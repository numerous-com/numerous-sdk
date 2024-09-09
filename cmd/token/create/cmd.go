package create

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/token"
)

var (
	name string
	desc string
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a personal access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := Create(cmd.Context(), token.NewService(gql.NewClient()), CreateInput{Name: name, Description: desc})
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&name, "name", "n", "", "The name of the personal access token. Must be unique, and no longer than 40 characters.")
	flags.StringVarP(&desc, "description", "d", "", "The description of the personal access token.")
}
