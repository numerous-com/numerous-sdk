package unshare

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/usage"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

const longFormat string = `Removes a shared URL for the specified app.

%s

%s
`

var long string = fmt.Sprintf(longFormat, usage.AppIdentifier("to remove a shared URL for"), usage.AppDirectoryArgument)

var Cmd = &cobra.Command{
	Use:   "unshare [app directory]",
	RunE:  run,
	Short: "Remove app shared URL",
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

	err := unshareApp(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	flags.StringVarP(&orgSlug, "organization", "o", "", "The organization slug identifier of the app to remove a shared URL for. List available organizations with 'numerous organization list'.")
	flags.StringVarP(&appSlug, "app", "a", "", "An app slug identifier of the app to remove a shared URL for.")
}
