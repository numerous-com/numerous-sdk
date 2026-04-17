package get

import (
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"
)

var Cmd = &cobra.Command{
	Use:   "get <task-instance-id>",
	RunE:  run,
	Short: "Get detailed information about a specific task instance",
	Long:  "Displays detailed information about a task instance, including resources usage and input data.",
	Args:  cobra.ExactArgs(1),
	Example: `Get details for a specific task instance:

	numerous task instance get ce5aba38-842d-4ee0-877b-4af9d426c848`,
}

func run(cmd *cobra.Command, args []string) error {
	taskInstanceID := args[0]

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	input := TaskGetInput{
		TaskInstanceID: taskInstanceID,
	}
	err := getInstance(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}
