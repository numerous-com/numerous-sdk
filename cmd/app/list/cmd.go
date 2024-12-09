package list

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var cmdArgs struct{ organizationSlug string }

var Cmd = &cobra.Command{
	Use:   "list",
	Short: "List all your apps (login required)",
	RunE:  func(cmd *cobra.Command, args []string) error { return run(cmd) },
}

func run(cmd *cobra.Command) error {
	if cmdArgs.organizationSlug == "" {
		output.PrintError("Missing organization argument.", "")
		cmd.Usage() // nolint:errcheck

		return errorhandling.ErrAlreadyPrinted
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	err := list(cmd.Context(), service, AppListInput{OrganizationSlug: cmdArgs.organizationSlug})

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	args.AddOrganizationSlugFlag(flags, "to list apps from", &cmdArgs.organizationSlug)
}
