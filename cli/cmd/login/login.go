package login

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"numerous/cli/auth"
	"numerous/cli/cmd/output"

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
			fmt.Println("✅ Great, you are already logged in!")
		}
	},
}

func Login(a auth.Authenticator, ctx context.Context) *auth.User {
	state, err := a.GetDeviceCode(ctx, http.DefaultClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf(`You are logging into Numerous.

When you press Enter, a browser window will automatically open.

Please verify that the code in the browser matches the code below to
complete authentication and click 'Confirm'.

Verification code: %s

Press Enter to continue...
`, state.UserCode)

	if _, err = fmt.Scanln(); err != nil {
		log.Fatal(err)
	}

	if err := a.OpenURL(state.VerificationURI); err != nil {
		fmt.Printf(
			"The browser could not be opened automatically, please go to this site to continue\n" +
				"the log-in: " + state.VerificationURI,
		)
	}

	result, err := a.WaitUntilUserLogsIn(ctx, http.DefaultClient, state)
	if err != nil {
		log.Fatal(err)
	}

	if err := a.StoreAccessToken(result.AccessToken); err != nil {
		output.PrintError(
			"Login failed.",
			"Error occurred storing access token in your keyring.\nError details: %s",
			err.Error(),
		)
		os.Exit(1)
	}
	if err := a.StoreRefreshToken(result.RefreshToken); err != nil {
		output.PrintError(
			"Error occurred storing refresh token in your keyring.",
			"When your access token expires, you will need to log in again.\n"+
				"Error details: "+err.Error(),
		)
	}

	fmt.Println("🎉 You are now logged in to Numerous!")

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
