package share

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

const longFormat string = `Creates a shared URL for the specified app.

%s

%s
`

var cmdActionText = "to create a shared URL for"

var long string = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)

var Cmd = &cobra.Command{
	Use:   "share [app directory]",
	RunE:  run,
	Short: "Create app shared URL",
	Long:  long,
	Args:  args.OptionalAppDir(&cmdArgs.appDir),
}

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := Input{
		AppDir:  cmdArgs.appDir,
		AppSlug: cmdArgs.appIdent.AppSlug,
		OrgSlug: cmdArgs.appIdent.OrganizationSlug,
	}

	err := shareApp(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
}
