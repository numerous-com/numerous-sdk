package deploy

import (
	"net/http"
	"os"

	"numerous/cli/internal/app"
	"numerous/cli/internal/gql"

	"github.com/spf13/cobra"
)

var DeployCmd = &cobra.Command{
	Use:   "deploy",
	Run:   run,
	Short: "Deploy an app to an organization.",
}

var (
	slug    string
	appName string
)

func run(cmd *cobra.Command, args []string) {
	service := app.New(gql.NewClient(), nil, http.DefaultClient)
	err := Deploy(cmd.Context(), ".", slug, appName, service)

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

	if err := DeployCmd.MarkFlagRequired("organization"); err != nil {
		panic(err.Error())
	}

	if err := DeployCmd.MarkFlagRequired("name"); err != nil {
		panic(err.Error())
	}
}
