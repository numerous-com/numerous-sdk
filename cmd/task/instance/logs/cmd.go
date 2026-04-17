package logs

import (
	"net/http"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

const longFormat string = `Read the logs of a specific task instance.

The task instance ID can be obtained from the "numerous task instances" command.
`

const example string = `To read the logs from a specific task instance:

    numerous task instance logs ce5aba38-842d-4ee0-877b-4af9d426c848

To tail the last 100 lines and follow new logs:

    numerous task instance logs ce5aba38-842d-4ee0-877b-4af9d426c848 --tail 100 --follow

To get logs without following (one-time read):

    numerous task instance logs ce5aba38-842d-4ee0-877b-4af9d426c848 --follow=false`

var Cmd = &cobra.Command{
	Use:     "logs <instance-id>",
	RunE:    run,
	Short:   "Display task instance logs",
	Long:    longFormat,
	Example: example,
	Args:    cobra.ExactArgs(1),
}

var cmdArgs struct {
	timestamps bool
	tail       int
	follow     bool
}

func run(cmd *cobra.Command, args []string) error {
	instanceID := args[0]

	var printer func(app.WorkloadLogEntry)
	if cmdArgs.timestamps {
		printer = TimestampPrinter
	} else {
		printer = TextPrinter
	}

	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)

	input := taskLogsInput{
		instanceID: instanceID,
		tail:       cmdArgs.tail,
		follow:     cmdArgs.follow,
		printer:    printer,
	}
	err := taskLogs(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	flags.BoolVarP(&cmdArgs.timestamps, "timestamps", "t", false, "Print a timestamp for each log entry.")
	flags.IntVarP(&cmdArgs.tail, "tail", "n", 0, "Number of lines to show from the end")
	flags.BoolVarP(&cmdArgs.follow, "follow", "f", true, "Continue streaming new log entries (default: true)")
}
