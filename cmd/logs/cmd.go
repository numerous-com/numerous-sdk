package logs

import (
	"net/http"
	"os"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:     "logs [app directory]",
	Run:     run,
	Short:   "Display running application logs",
	GroupID: group.AppCommandsGroupID,
	Long: `Read the logs of an application deployed to an organization on the
Numerous platform.

If <name> and <organization> flags are set, they define the app to read logs
from. If they are not, the default deployment section in the manifest is used,
if it is defined.

If [app directory] is specified, that directory will be used to read the
app manifest for the default deployment information.

If no [app directory] is specified, the current working directory is used.`,
	Example: `To read the logs from a specific app deployment, use the following form:

    numerous logs --organization "organization-slug-a2ecf59b" --app "my-app"

Otherwise, assuming an app has been initialized in the directory
"my_project/my_app" and has a default deployment defined in its manifest:

    numerous logs my_project/my_app
`,
	Args: args.OptionalAppDir(&appDir),
}

var (
	orgSlug    string
	appSlug    string
	timestamps bool
	appDir     string = "."
)

func run(cmd *cobra.Command, args []string) {
	// TODO: this is just here for users who expect the "old" log command in
	// this location, which will primarily be for apps initialized with an App
	// ID file
	if exists, _ := dir.AppIDExists(appDir); exists {
		output.NotifyCmdMoved("numerous log", "numerous legacy log")
		println()
	}

	var printer func(app.AppDeployLogEntry)
	if timestamps {
		printer = TimestampPrinter
	} else {
		printer = TextPrinter
	}
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)

	if err := Logs(cmd.Context(), service, appDir, orgSlug, appSlug, printer); err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func init() {
	flags := LogsCmd.Flags()
	flags.StringVarP(&orgSlug, "organization", "o", "", "The organization slug identifier of the app to read logs from.")
	flags.StringVarP(&appSlug, "app", "a", "", "The app slug identifier of the app to read logs from.")
	flags.BoolVarP(&timestamps, "timestamps", "t", false, "Print a timestamp for each log entry.")
}
