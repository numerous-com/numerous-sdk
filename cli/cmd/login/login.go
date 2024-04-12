package login

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"numerous/cli/auth"

	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login in to Numerous",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
		if user == nil {
			Login(auth.NumerousTenantAuthenticator, cmd.Context())
		} else {
			fmt.Println("Great, you are already logged in!")
		}
	},
}

func Login(a auth.Authenticator, ctx context.Context) *auth.User {
	state, err := a.GetDeviceCode(ctx, http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\n%s\n\n", "You are logging into Numerous. \nWhen you press Enter, a browser window will automatically open. \nPlease verify that the code in the browser matches the code below to complete authentication and click 'Confirm'.")
	fmt.Println("Verification code: " + state.UserCode)
	fmt.Println("Press Enter to continue...")

	if _, err = fmt.Scanln(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err := a.OpenURL(state.VerificationURI); err != nil {
		fmt.Printf("Whoops, we ran into an error opening the verification URL in your browser. \nPlease copy and paste the following URL into your browser: %s\n", state.VerificationURI)
	}

	result, err := a.WaitUntilUserLogsIn(ctx, http.DefaultClient, state)
	if err != nil {
		log.Fatal(err)
	}

	if err := a.StoreAccessToken(result.AccessToken); err != nil {
		fmt.Printf("An error occurred while storing your access token. We could not log you in. %s\n", err)
		os.Exit(1)
	}
	if err := a.StoreRefreshToken(result.RefreshToken); err != nil {
		fmt.Printf("An error occurred while storing your refresh token to the keyring. %s\n", err)
		fmt.Printf("You will need to login again when your access token expire.\n\n")
	}

	fmt.Println("You are now logged in to Numerous!")

	return a.GetLoggedInUserFromKeyring()
}

func RefreshAccessToken(user *auth.User, client *http.Client, a auth.Authenticator) error {
	if err := user.CheckAuthenticationStatus(); err != auth.ErrExpiredToken {
		if err != nil {
			return err
		}

		return nil
	}

	newAccessToken, err := a.RegenerateAccessToken(client, user.RefreshToken)
	if err != nil {
		return err
	}
	if err := a.StoreAccessToken(newAccessToken); err != nil {
		return err
	}
	user.AccessToken = newAccessToken

	return nil
}
