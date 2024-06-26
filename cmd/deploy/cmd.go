package deploy

import (
	"net/http"
	"os"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [app directory]",
	Run:   run,
	Short: "Deploy an app to an organization",
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
be pushed to the organization "organization-slug-a2ecf59b", and the app name
"my-app", the following command can be used:

	numerous app deploy --organization "organization-slug-a2ecf59b" --name "my-app"
	`,
	Args: args.OptionalAppDir(&appDir),
}

var (
	slug       string
	appName    string
	verbose    bool
	appDir     string = "."
	projectDir string = "."
	message    string
	version    string
)

func run(cmd *cobra.Command, args []string) {
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)
	input := DeployInput{
		AppDir:     appDir,
		ProjectDir: projectDir,
		Slug:       slug,
		AppName:    appName,
		Message:    message,
		Version:    version,
		Verbose:    verbose,
	}
	err := Deploy(cmd.Context(), service, input)

	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func init() {
	flags := DeployCmd.Flags()
	flags.StringVarP(&slug, "organization", "o", "", "The organization slug identifier. List available organizations with 'numerous organization list'.")
	flags.StringVarP(&appName, "name", "n", "", "A unique name for the application to deploy.")
	flags.BoolVarP(&verbose, "verbose", "v", false, "Display detailed information about the app deployment.")
	flags.StringVarP(&projectDir, "project-dir", "p", "", "The project directory, which is the build context if using a custom Dockerfile.")
}
