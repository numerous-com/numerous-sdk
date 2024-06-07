package deploy

import (
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

	numerous deploy --organization "organization-slug-a3ecfh2b" --name "my-app"
	`,
}

var (
	slug    string
	appName string
	verbose bool
)

func run(cmd *cobra.Command, args []string) {
	service := app.New(gql.NewClient(), gql.NewSubscriptionClient(), http.DefaultClient)
	err := Deploy(cmd.Context(), ".", slug, verbose, appName, service)

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

	if err := DeployCmd.MarkFlagRequired("organization"); err != nil {
		panic(err.Error())
	}

	if err := DeployCmd.MarkFlagRequired("name"); err != nil {
		panic(err.Error())
	}
}
