package publish

import (
	"fmt"
	"os"

	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"
	"numerous/cli/tool"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var PublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publishes an app to the public app gallery",
	Run: func(cmd *cobra.Command, args []string) {
		err := publish(gql.GetClient())
		if err != nil {
			os.Exit(1)
		}
	},
}

func publish(client *gqlclient.Client) error {
	appID, err := tool.ReadToolID(".")
	if err == tool.ErrToolIDNotFound {
		fmt.Println("The current directory is not a numerous app",
			"\nrun \"numerous init\" to initialize a numerous app in the current directory")

		return err
	} else if err != nil {
		fmt.Println("An error occurred reading the app ID")
		fmt.Println("Error: ", err)

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
		fmt.Println("An error occurred when publishing the app")
		fmt.Println("Error: ", err)

		return err
	} else {
		fmt.Println("The app has been published to the open app gallery!")
		fmt.Printf("Access it here: %s\n", t.PublicURL)

		return nil
	}
}
