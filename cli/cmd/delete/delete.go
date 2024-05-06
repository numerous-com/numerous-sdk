package deleteapp

import (
	"errors"
	"fmt"
	"os"

	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"
	"numerous/cli/tool"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [app ID]",
	Short: "Deletes the app and removes its associated resources",
	Long: `Removes the app from the server and deletes any associated resources, such as docker images or containers.
This action cannot be undone.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := deleteApp(gql.GetClient(), args)
		if err != nil {
			os.Exit(1)
		}
	},
}

func deleteApp(client *gqlclient.Client, args []string) error {
	var appID string
	var err error

	if len(args) == 1 {
		appID = args[0]
	} else {
		appID, err = tool.ReadAppID(".")
		if err == tool.ErrAppIDNotFound {
			fmt.Println("Sorry, we could not recognize your app in the specified directory.",
				"\nRun \"numerous init\" to initialize the app in Numerous.")

			return err
		} else if err != nil {
			fmt.Println("Whoops! An error occurred when reading the app ID. \n Please make sure you are in the correct directory and try again.")
			fmt.Println("Error: ", err)

			return err
		}
	}

	if _, err := app.Query(appID, client); err != nil {
		fmt.Println(
			"Sorry, we could not find the app in our database. \nPlease, make sure that the App ID in the .tool_id.txt file is correct and try again.")

		return err
	}

	if result, err := app.Delete(appID, client); err != nil {
		fmt.Println("An error occurred while removing the app from Numerous. Please try again.")
		fmt.Println("Error: ", err)

		return err
	} else {
		if result.ToolDelete.Typename == "ToolDeleteSuccess" {
			fmt.Println("The app has been successfully removed from Numerous")
		} else if result.ToolDelete.Typename == "ToolDeleteFailure" {
			fmt.Println("An error occurred while removing the app from Numerous. Please try again.")
			fmt.Println("Error: ", result.ToolDelete.Result)

			return errors.New(result.ToolDelete.Result)
		}

		return nil
	}
}
