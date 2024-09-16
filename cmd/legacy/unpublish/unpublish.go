package unpublish

import (
	"fmt"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var UnpublishCmd = &cobra.Command{
	Use:   "unpublish",
	Short: "Removes a published app from the public app gallery",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := unpublish(gql.GetClient())
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func unpublish(client *gqlclient.Client) error {
	appDir := "."
	appID, err := dir.ReadAppID(appDir)
	if err != nil {
		dir.PrintReadAppIDErrors(err, appDir)
		return err
	}

	if a, err := app.Query(appID, client); err != nil {
		output.PrintError("The app could not be found in the database.", "")
		return err
	} else if a.PublicURL == "" {
		output.PrintError("The app is currently not published to the public app gallery.", "")
		return nil
	}

	if _, err := app.Unpublish(appID, client); err != nil {
		output.PrintErrorDetails("An error occurred when unpublishing the app", err)
		return err
	} else {
		fmt.Println("The app has been removed from the public app gallery!")
		return nil
	}
}
