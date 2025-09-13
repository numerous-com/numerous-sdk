package list

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

const longFormat string = `Lists tasks defined for an application.

Shows all tasks related to the current app version.

%s

%s
`

var (
	cmdActionText = "to list tasks for"
	long          = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)
)

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

var Cmd = &cobra.Command{
	Use:   "list [app directory]",
	RunE:  run,
	Short: "List tasks defined for an app",
	Long:  long,
	Args:  args.OptionalAppDir(&cmdArgs.appDir),
	Example: `To list all tasks for a specific app:

	numerous task list --organization "my-org" --app "my-app"

Otherwise, assuming an app has been initialized in the current directory:

	numerous task list`,
}

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskListInput{
		AppDir:           cmdArgs.appDir,
		OrganizationSlug: cmdArgs.appIdent.OrganizationSlug,
		AppSlug:          cmdArgs.appIdent.AppSlug,
	}
	err := list(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
}
