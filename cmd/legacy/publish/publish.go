package publish

import (
	"fmt"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/output"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publishes an app to the public app gallery",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := publish(gql.GetClient())
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func publish(client *gqlclient.Client) error {
	appDir := "."
	appID, err := dir.ReadAppID(appDir)
	if err != nil {
		dir.PrintReadAppIDErrors(err, appDir)
		return err
	}

	if a, err := app.Query(appID, client); err != nil {
		fmt.Println("The app could not be found in the database.")
		return err
	} else if a.PublicURL != "" {
		fmt.Println("The app has already been published to the open app gallery!")
		fmt.Printf("Access it here: %s\n", a.PublicURL)

		return nil
	}

	if t, err := app.Publish(appID, client); err != nil {
		output.PrintErrorDetails("An error occurred when publishing the app", err)

		return err
	} else {
		fmt.Println("The app has been published to the open app gallery!")
		fmt.Printf("Access it here: %s\n", t.PublicURL)

		return nil
	}
}
