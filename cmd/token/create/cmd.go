package create

import (
	"github.com/spf13/cobra"
)

var (
	name string
	desc string
)

var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Create a user access token.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&name, "name", "n", "", "The name of the user access token. Must be unique, and no longer than 40 characters.")
	flags.StringVarP(&desc, "description", "d", "", "The description of the user access token.")
}
