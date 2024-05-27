package deploy

import (
	"os"

	"numerous/cli/app"
	"numerous/cli/cmd/validate"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Run:   deploy,
	Short: "Deploy an app to an organization.",
}

var (
	slug    string
	appName string
)

func deploy(cmd *cobra.Command, args []string) {
	if !validate.IsValidIdentifier(slug) {
		println("Error: Invalid organization '" + slug + "'. Must contain only lower-case alphanumerical characters and dashes.")
		os.Exit(1)
	}

	if !validate.IsValidIdentifier(appName) {
		println("Error: Invalid app name '" + appName + "'. Must contain only lower-case alphanumerical characters and dashes.")
		os.Exit(1)
	}

	err := app.Deploy(slug, appName)
	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func init() {
	flags := DeployCmd.Flags()
	flags.StringVarP(&slug, "organization", "o", "", "Find the organization slug in the browser URL, https://numerous.com/app/organization/<ORGANIZATION_SLUG>")
	flags.StringVarP(&appName, "name", "n", "", "A unique name for the application to deploy.")

	if err := DeployCmd.MarkFlagRequired("organization"); err != nil {
		panic(err.Error())
	}

	if err := DeployCmd.MarkFlagRequired("name"); err != nil {
		panic(err.Error())
	}
}
