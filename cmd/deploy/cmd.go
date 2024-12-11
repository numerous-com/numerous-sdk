package deploy

import (
	"fmt"
	"net/http"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/usage"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

const longFormat string = `Deploys an application to an organization on the Numerous platform.

After deployment the deployed version of the app is available in the
organization's apps page.

%s

%s
`

var (
	cmdActionText        = "to deploy"
	long          string = fmt.Sprintf(longFormat, usage.AppIdentifier(cmdActionText), usage.AppDirectoryArgument)
)

var Cmd = &cobra.Command{
	Use:     "deploy [app directory]",
	RunE:    run,
	GroupID: group.AppCommandsGroupID,
	Short:   "Deploy an app to an organization",
	Long:    long,
	Example: `
If an app has been initialized in the current working directory, and it should
be pushed to the organization "organization-slug-a2ecf59b", and the app slug
"my-app", the following command can be used:

	numerous deploy --organization "organization-slug-a2ecf59b" --app "my-app"
	`,
	Args: args.OptionalAppDir(&cmdArgs.appDir),
}

var cmdArgs struct {
	appIdent   args.AppIdentifierArg
	verbose    bool
	appDir     string
	projectDir string
	message    string
	version    string
	follow     bool
}

func run(cmd *cobra.Command, args []string) error {
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)
	input := deployInput{
		appDir:     cmdArgs.appDir,
		projectDir: cmdArgs.projectDir,
		orgSlug:    cmdArgs.appIdent.OrganizationSlug,
		appSlug:    cmdArgs.appIdent.AppSlug,
		message:    cmdArgs.message,
		version:    cmdArgs.version,
		verbose:    cmdArgs.verbose,
		follow:     cmdArgs.follow,
	}
	err := deploy(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := Cmd.Flags()
	cmdArgs.appIdent.AddAppIdentifierFlags(flags, cmdActionText)
	flags.BoolVarP(&cmdArgs.verbose, "verbose", "v", false, "Display detailed information about the app deployment.")
	flags.BoolVarP(&cmdArgs.follow, "follow", "f", false, "Follow app deployment logs after deployment has succeeded.")
	flags.StringVarP(&cmdArgs.projectDir, "project-dir", "p", "", "The project directory, which is the build context if using a custom Dockerfile.")
}
