package stop

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

const longFormat string = `Stops a task instance.

Stops the specified task instance by its ID.

%s

%s
`

var (
	cmdActionText = "to stop a task instance for"
	long          = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)
)

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

var Cmd = &cobra.Command{
	Use:   "stop <task-instance-id> [app directory]",
	RunE:  run,
	Short: "Stop a task instance",
	Long:  long,
	Args:  cobra.MinimumNArgs(1),
	Example: `To stop a task instance for a specific app:

	numerous task stop ce5aba38-842d-4ee0-877b-4af9d426c848 --organization "my-org" --app "my-app"

Otherwise, assuming an app has been initialized in the current directory:

	numerous task stop ce5aba38-842d-4ee0-877b-4af9d426c848`,
}

func run(cmd *cobra.Command, args []string) error {
	taskInstanceID := args[0]

	if len(args) > 1 {
		cmdArgs.appDir = args[1]
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskStopInput{
		AppDir:           cmdArgs.appDir,
		OrganizationSlug: cmdArgs.appIdent.OrganizationSlug,
		AppSlug:          cmdArgs.appIdent.AppSlug,
		TaskInstanceID:   taskInstanceID,
	}
	err := stopTask(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
}
