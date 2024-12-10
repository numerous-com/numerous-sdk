package logs

import (
	"fmt"
	"net/http"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/cmd/usage"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

const longFormat string = `Read the logs of an application deployed to an organization on the
Numerous platform.

%s

%s
`
const cmdActionText string = "to read logs from"

var long string = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)

const example string = `To read the logs from a specific app deployment, use the following form:

    numerous logs --organization "organization-slug-a2ecf59b" --app "my-app"

Otherwise, assuming an app has been initialized in the directory
"my_project/my_app" and has a default deployment defined in its manifest:

    numerous logs my_project/my_app
`

var Cmd = &cobra.Command{
	Use:     "logs [app directory]",
	RunE:    run,
	Short:   "Display running application logs",
	GroupID: group.AppCommandsGroupID,
	Long:    long,
	Example: example,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
}

var cmdArgs struct {
	appIdent   args.AppIdentifierArg
	timestamps bool
	appDir     string
}

func run(cmd *cobra.Command, args []string) error {
	// TODO: this is just here for users who expect the "old" log command in
	// this location, which will primarily be for apps initialized with an App
	// ID file
	if exists, _ := dir.AppIDExists(cmdArgs.appDir); exists {
		output.NotifyCmdMoved("numerous log", "numerous legacy log")
		println()
	}

	var printer func(app.AppDeployLogEntry)
	if cmdArgs.timestamps {
		printer = TimestampPrinter
	} else {
		printer = TextPrinter
	}
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)
	err := Logs(cmd.Context(), service, cmdArgs.appDir, cmdArgs.appIdent.OrganizationSlug, cmdArgs.appIdent.AppSlug, printer)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
	flags.BoolVarP(&cmdArgs.timestamps, "timestamps", "t", false, "Print a timestamp for each log entry.")
}
