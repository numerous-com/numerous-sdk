package list

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/gql"
)

var cmdArgs struct{ organizationSlug string }

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List all your apps (login required)",
	RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd) },
}

func run(cmd *cobra.Command) error {
	orgSlug := cmdArgs.organizationSlug
	if orgSlug == "" {
		orgSlug = config.OrganizationSlug()
	}

	if orgSlug == "" {
		output.PrintError(
			"No organization provided or configured",
			"Specify an organization with the --organization flag, or configure one with \"numerous config\".",
		)
		cmd.Usage() // nolint:errcheck

		return errorhandling.ErrAlreadyPrinted
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	err := list(cmd.Context(), service, AppListInput{OrganizationSlug: cmdArgs.organizationSlug})

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&cmdArgs.organizationSlug, "organization", "o", "", "The organization slug identifier to list apps from. List available organizations with 'numerous organization list'.")
}
