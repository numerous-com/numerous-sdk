package deploy

import (
	"net/http"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:     "deploy [app directory]",
	RunE:    run,
	GroupID: group.AppCommandsGroupID,
	Short:   "Deploy an app to an organization",
	Long: `Deploys an application to an organization on the Numerous platform.

An app's deployment is identified with the <name> and <organization> identifier.
Deploying an app to a given <name> and <organization> combination, will override
the existing version.

The <name> must contain only lower-case alphanumeric characters and dashes.

After deployment the deployed version of the app is available in the
organization's apps page.

If no [app directory] is specified, the current working directory is used.`,
	Example: `
If an app has been initialized in the current working directory, and it should
be pushed to the organization "organization-slug-a2ecf59b", and the app slug
"my-app", the following command can be used:

	numerous deploy --organization "organization-slug-a2ecf59b" --app "my-app"
	`,
	Args: args.OptionalAppDir(&appDir),
}

var (
	orgSlug    string
	appSlug    string
	verbose    bool
	appDir     string = "."
	projectDir string = "."
	message    string
	version    string
)

func run(cmd *cobra.Command, args []string) error {
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)
	input := DeployInput{
		AppDir:     appDir,
		ProjectDir: projectDir,
		OrgSlug:    orgSlug,
		AppSlug:    appSlug,
		Message:    message,
		Version:    version,
		Verbose:    verbose,
	}
	err := Deploy(cmd.Context(), service, input)

	return errorhandling.ErrorAlreadyPrinted(err)
}

func init() {
	flags := DeployCmd.Flags()
	flags.StringVarP(&orgSlug, "organization", "o", "", "The organization slug identifier of the app to deploy to. List available organizations with 'numerous organization list'.")
	flags.StringVarP(&appSlug, "app", "a", "", "A app slug identifier of the app to deploy to.")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Display detailed information about the app deployment.")
	flags.StringVarP(&projectDir, "project-dir", "p", "", "The project directory, which is the build context if using a custom Dockerfile.")
}
