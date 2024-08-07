package download

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var long string = `Download app sources of the specified app.

Downloads the sources of an app into the specified [app directory], or the
directory named after the app slug.

If an app already exists in the [app directory] and it has a default deployment
specified organization and app slugs are not required. In this case the app
source code will be downloaded directly on top of the local source, so be
careful!

Prompts are displayed for all downloaded files that would override an existing
file.
`

var Cmd = &cobra.Command{
	Use:     "download [app directory]",
	RunE:    run,
	Short:   "Download app sources",
	Long:    long,
	GroupID: group.AppCommandsGroupID,
	Args:    args.OptionalAppDir(&appDir),
}

var (
	orgSlug string
	appSlug string
	appDir  string
)

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := Input{
		AppDir:  appDir,
		AppSlug: appSlug,
		OrgSlug: orgSlug,
	}

	return Download(cmd.Context(), http.DefaultClient, service, input, surveyConfirmOverwrite)
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&orgSlug, "organization", "o", "", "The organization slug identifier of the app to download. List available organizations with 'numerous organization list'.")
	flags.StringVarP(&appSlug, "app", "a", "", "A app slug identifier of the app to download.")
}
