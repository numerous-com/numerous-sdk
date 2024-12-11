package download

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var long string = `Download app sources of the specified app.

Downloads the sources of an app into the specified [app directory]. If
[app directory] is not specified the app slug will be used as the folder name.

If an app already exists in [app directory], and a default deployment is
configured in numerous.toml, then that will be used to identify the app to
download the source from. In this case the app source code will be downloaded
directly on top of the local source, so be careful!

A confirmation prompt will be shown if file overwrites are a possibility.
`

var Cmd = &cobra.Command{
	Use:     "download [app directory]",
	RunE:    run,
	Short:   "Download app sources",
	Long:    long,
	GroupID: group.AppCommandsGroupID,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
}

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := downloadInput{
		appDir:             cmdArgs.appDir,
		appSlug:            cmdArgs.appIdent.AppSlug,
		orgSlug:            cmdArgs.appIdent.OrganizationSlug,
		overwriteConfirmer: surveyConfirmOverwrite,
	}

	err := download(cmd.Context(), http.DefaultClient, service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, "to download")
}
