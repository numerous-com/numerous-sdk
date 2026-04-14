package stop

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var Cmd = &cobra.Command{
	Use:   "stop <task-instance-id>",
	RunE:  run,
	Short: "Stop a task instance",
	Long:  "Stops the specified task instance by its ID.",
	Args:  cobra.ExactArgs(1),
	Example: `To stop a task instance:

	numerous task instance stop ce5aba38-842d-4ee0-877b-4af9d426c848`,
}

func run(cmd *cobra.Command, args []string) error {
	taskInstanceID := args[0]

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskStopInput{
		TaskInstanceID: taskInstanceID,
	}
	err := stopTask(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}
