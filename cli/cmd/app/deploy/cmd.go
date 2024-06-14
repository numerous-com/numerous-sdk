package deploy

import (
	"fmt"
	"net/http"
	"os"

	"numerous/cli/internal/app"
	"numerous/cli/internal/gql"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy [app directory]",
	Run:   run,
	Short: "Deploy an app to an organization.",
	Long: `Deploys an application to an organization on the Numerous platform.

An apps deployment is identified with the <name> and <organization> identifier.
Deploying an app to a given <name> and <organization> combination, will override
the existing version.

The <name> must contain only lower-case alphanumeric characters and dashes.

After deployment the deployed version of the app is available in the
organization's apps page.

If no [app directory] is specified, the current working directory is used.`,
	Example: `
If an app has been initialized in the current working directory, and it should
be pushed to the organization "organization-slug-a3ecfh2b", and the app name
"my-app", the following command can be used:

	numerous app deploy --organization "organization-slug-a3ecfh2b" --name "my-app"
	`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			fn := cmd.HelpFunc()
			fn(cmd, args)

			return fmt.Errorf("accepts only an optional [app directory] as a positional argument, you provided %d arguments", len(args))
		}

		if len(args) == 1 {
			appDir = args[0]
		}

		return nil
	},
}

var (
	slug       string
	appName    string
	verbose    bool
	appDir     string = "."
	projectDir string = "."
)

func run(cmd *cobra.Command, args []string) {
	sc := gql.NewSubscriptionClient().WithSyncMode(true)
	service := app.New(gql.NewClient(), sc, http.DefaultClient)
	err := Deploy(cmd.Context(), service, appDir, projectDir, slug, appName, verbose)

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

	if err := DeployCmd.MarkFlagRequired("organization"); err != nil {
		panic(err.Error())
	}

	if err := DeployCmd.MarkFlagRequired("name"); err != nil {
		panic(err.Error())
	}
}
