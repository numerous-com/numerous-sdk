package deleteapp

import (
	"errors"
	"fmt"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/output"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
	Use:   "delete [app ID]",
	Short: "Deletes the app and removes its associated resources",
	Long: `Removes the app from the server and deletes any associated resources, such as docker images or containers.
This action cannot be undone.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := deleteApp(gql.GetClient(), args)
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

func deleteApp(client *gqlclient.Client, args []string) error {
	var appID string
	appDir := "."
	if len(args) == 1 {
		appID = args[0]
	} else if readAppID, err := dir.ReadAppID(appDir); err != nil {
		dir.PrintReadAppIDErrors(err, appDir)
		return err
	} else {
		appID = readAppID
	}

	if _, err := app.Query(appID, client); err != nil {
		output.PrintError(
			"Sorry, we could not find the app in our database.",
			"Please, make sure that the App ID in the \"%s\" file is correct and try again.",
			dir.AppIDFileName,
		)

		return err
	}

	if result, err := app.Delete(appID, client); err != nil {
		output.PrintUnknownError(err)

		return err
	} else {
		switch result.ToolDelete.Typename {
		case "ToolDeleteSuccess":
			fmt.Println("The app has been successfully removed from Numerous")
		case "ToolDeleteFailure":
			err := errors.New(result.ToolDelete.Result)
			output.PrintUnknownError(err)

			return err
		}

		return nil
	}
}
