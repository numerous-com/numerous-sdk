package deletecmd

import (
	"net/http"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:     "delete [app directory]",
	RunE:    run,
	Short:   "Delete an app from an organization",
	GroupID: "app-cmds",
	Long: `Deletes the specified app from the organization.

If <name> and <organization> flags are set, they define the app to delete. If
they are not the default deployment section in the manifest is used if it is
defined.

If [app directory] is specified, that directory will be used to read the
app manifest for the default deployment information.

If no [app directory] is specified, the current working directory is used.`,
	Example: `To delete an app use the following form:

    numerous app delete --organization "organization-slug-a2ecf59b" --name "my-app"

Otherwise, assuming an app has been initialized in the directory
"my_project/my_app" and has a default deployment defined in its manifest:

    numerous app delete my_project/my_app
`,
	Args: args.OptionalAppDir(&appDir),
}

var (
	slug    string
	appName string
	appDir  string = "."
)

func run(cmd *cobra.Command, args []string) error {
	if exists, _ := dir.AppIDExists(appDir); exists {
		output.NotifyCmdChanged("numerous delete", "numerous legacy delete")
		println()
	}

	service := app.New(gql.NewClient(), nil, http.DefaultClient)

	return Delete(cmd.Context(), service, appDir, slug, appName)
}

func init() {
	flags := DeleteCmd.Flags()
	flags.StringVarP(&slug, "organization", "o", "", "The organization slug identifier of the app to read logs from.")
	flags.StringVarP(&appName, "name", "n", "", "The name of the app to read logs from.")
}
