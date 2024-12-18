package deletecmd

import (
	"fmt"
	"net/http"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/usage"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/output"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "delete [app directory]",
	RunE:    run,
	Short:   "Delete an app from an organization",
	GroupID: group.AppCommandsGroupID,
	Long:    long,
	Example: example,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
}

const longFormat = `Deletes the specified app from the organization.

%s

%s
`

var long string = fmt.Sprintf(longFormat, usage.AppIdentifier("to delete"), usage.AppDirectoryArgument)

const example = `To delete an app use the following form:

numerous delete --organization "organization-slug-a2ecf59b" --app "my-app"

Otherwise, assuming an app has been initialized in the directory
"my_project/my_app" and has a default deployment defined in its manifest:

numerous delete my_project/my_app
`

var cmdArgs struct {
	appIdent args.AppIdentifierArg
	appDir   string
}

func run(cmd *cobra.Command, _ []string) error {
	if exists, _ := dir.AppIDExists(cmdArgs.appDir); exists {
		output.NotifyCmdChanged("numerous delete", "numerous legacy delete")
		println()
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	err := deleteApp(cmd.Context(), service, cmdArgs.appDir, cmdArgs.appIdent.OrganizationSlug, cmdArgs.appIdent.AppSlug)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	f := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(f, "")
}
