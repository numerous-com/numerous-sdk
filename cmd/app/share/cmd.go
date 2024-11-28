package share

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var long string = `Creates a shared URL for the specified app.

If <name> and <organization> flags are set, they define the app
to create a shared URL for.
If they are not, the default deployment section in the manifest is used,
if it is defined.

If [app directory] is specified, that directory will be used to read the
app manifest for the default deployment information.

If no [app directory] is specified, the current working directory is used.
`

var Cmd = &cobra.Command{
	Use:   "share [app directory]",
	RunE:  run,
	Short: "Create app shared URL",
	Long:  long,
	Args:  args.OptionalAppDir(&appDir),
}

var (
	orgSlug string
	appSlug string
	appDir  string
)

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}

	err := shareApp(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&orgSlug, "organization", "o", "", "The organization slug identifier of the app to create a shared URL for. List available organizations with 'numerous organization list'.")
	flags.StringVarP(&appSlug, "app", "a", "", "An app slug identifier of the app to create a shared URL for.")
}
