package list

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var argOrganizationSlug string

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List all your apps (login required)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if argOrganizationSlug == "" {
			output.PrintError("Missing organization argument.", "")
			cmd.Usage() // nolint:errcheck

			return errorhandling.ErrAlreadyPrinted
		}

		service := app.New(gql.NewClient(), nil, http.DefaultClient)
		err := list(cmd.Context(), service, AppListInput{OrganizationSlug: argOrganizationSlug})

		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func init() {
	Cmd.Flags().StringVarP(&argOrganizationSlug, "organization", "o", "", "The organization slug identifier to list app from.")
}
