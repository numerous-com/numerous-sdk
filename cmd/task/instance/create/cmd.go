package create

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

const longFormat string = `Creates and starts a new instance of a specific task.

Creates and starts a new task instance in an app deployment.

%s

%s
`

var (
	cmdActionText = "to start a task for"
	long          = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)
)

var cmdArgs struct {
	appIdent  args.AppIdentifierArg
	appDir    string
	input     string
	inputFile string
}

var Cmd = &cobra.Command{
	Use:   "create <task-name> [app directory]",
	RunE:  run,
	Short: "Create and start a new instance of a specific task",
	Long:  long,
	Args:  cobra.MinimumNArgs(1),
	Example: `To create a new instance of task named "worker" for a specific app:

	numerous task instance create worker --organization "my-org" --app "my-app"

Otherwise, assuming an app has been initialized in the current directory:

	numerous task instance create worker

With input data:

	numerous task instance create worker --input "user123"
	numerous task instance create worker --input '{"user_id": 123, "action": "process"}'

With input from a file:

	numerous task instance create worker --input-file config.json`,
}

func run(cmd *cobra.Command, args []string) error {
	taskName := args[0]

	if len(args) > 1 {
		cmdArgs.appDir = args[1]
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskStartInput{
		AppDir:           cmdArgs.appDir,
		OrganizationSlug: cmdArgs.appIdent.OrganizationSlug,
		AppSlug:          cmdArgs.appIdent.AppSlug,
		TaskName:         taskName,
		Input:            cmdArgs.input,
		InputFile:        cmdArgs.inputFile,
	}
	err := startTask(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
	flags.StringVar(&cmdArgs.input, "input", "", "Input data to pass to the task")
	flags.StringVar(&cmdArgs.inputFile, "input-file", "", "Path to file containing input data to pass to the task")
}
