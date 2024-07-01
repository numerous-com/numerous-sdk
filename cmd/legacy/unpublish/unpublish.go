package unpublish

import (
	"fmt"
	"os"

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
	Run: func(cmd *cobra.Command, args []string) {
		err := unpublish(gql.GetClient())
		if err != nil {
			os.Exit(1)
		}
	},
}

func unpublish(client *gqlclient.Client) error {
	appDir := "."
	appID, err := dir.ReadAppID(appDir)
	if err != nil {
		output.PrintReadAppIDErrors(err, appDir)
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
