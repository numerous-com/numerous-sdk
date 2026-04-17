package instances

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

const longFormat string = `Lists task instances for a specific task in an app deployment.

Shows instances of the specified task,
including their status and resource usage.

%s

%s
`

var (
	cmdActionText = "to list task instances for"
	long          = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)
)

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

var Cmd = &cobra.Command{
	Use:   "instances <task-id> [app directory]",
	RunE:  run,
	Short: "List instances of a specific task",
	Long:  long,
	Args:  cobra.MinimumNArgs(1),
	Example: `To list instances of task named "worker" for a specific app:

	numerous task instances worker --organization "my-org" --app "my-app"

Otherwise, assuming an app has been initialized in the current directory:

	numerous task instances worker`,
}

func run(cmd *cobra.Command, args []string) error {
	taskID := args[0]

	if len(args) > 1 {
		cmdArgs.appDir = args[1]
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskInstancesInput{
		AppDir:           cmdArgs.appDir,
		OrganizationSlug: cmdArgs.appIdent.OrganizationSlug,
		AppSlug:          cmdArgs.appIdent.AppSlug,
		TaskID:           taskID,
	}
	err := listInstances(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
}
