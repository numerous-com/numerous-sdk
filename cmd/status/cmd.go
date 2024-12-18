package status

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
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
	appIdent     args.AppIdentifierArg
	appDir       string
	metricsSince Since
}

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "Get the status of an app's workloads",
	Long:    long,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
	RunE:    run,
	GroupID: group.AppCommandsGroupID,
}

func run(cmd *cobra.Command, args []string) error {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)

	input := statusInput{
		appDir:       cmdArgs.appDir,
		appSlug:      cmdArgs.appIdent.AppSlug,
		orgSlug:      cmdArgs.appIdent.OrganizationSlug,
		metricsSince: cmdArgs.metricsSince.Time(),
	}
	err := status(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
	f := flags.VarPF(&cmdArgs.metricsSince, "metrics-since", "", `Read metrics since this time. Can be an RFC3339 timestamp (e.g. "2024-01-01T12:00:00Z"), a date (e.g. "2024-06-06"), or a duration of seconds, minutes, hours or days (e.g. "1s", "10m", "5h", "2d").`)
	f.DefValue = `"1h"` // Hack to display correct default value in the help text
}
