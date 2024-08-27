package create

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/token"
)

var (
	name string
	desc string
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a user access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return Create(cmd.Context(), token.New(gql.NewClient()), CreateInput{Name: name, Description: desc})
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&name, "name", "n", "", "The name of the user access token. Must be unique, and no longer than 40 characters.")
	flags.StringVarP(&desc, "description", "d", "", "The description of the user access token.")
}
