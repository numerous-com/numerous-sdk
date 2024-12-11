package status

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

const cmdActionText = "to see the status of"

const longFormat = `Get an overview of the status of all workloads related to an app.

%s

%s
`

var long = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)

var cmdArgs struct {
	orgSlug string
	appSlug string
	appDir  string
}

var Cmd = &cobra.Command{
	Use:   "status",
	Short: "Get the status of an apps workloads",
	Long:  long,
	Args:  args.OptionalAppDir(&cmdArgs.appDir),
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)

	input := statusInput{
		appDir:  cmdArgs.appDir,
		appSlug: cmdArgs.appSlug,
		orgSlug: cmdArgs.orgSlug,
	}
	err := status(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}
