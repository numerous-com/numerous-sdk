package login

import (
	"context"
	"fmt"
	"net/http"

	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"

	"github.com/spf13/cobra"
)

var LoginCmd = &cobra.Command{
	Use:     "login",
	Short:   "Login in to Numerous",
	Args:    cobra.NoArgs,
	GroupID: group.AdditionalCommandsGroupID,
	RunE: func(cmd *cobra.Command, args []string) error {
		user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
		if user != nil {
			output.PrintlnOK("Great, you are already logged in!")
			return nil
		}

		_, err := Login(auth.NumerousTenantAuthenticator, cmd.Context())

		return err
	},
}

const loggingInMessage = `You are logging into Numerous.

When you press Enter, a browser window will automatically open.

Please verify that the code in the browser matches the code below to
complete authentication and click 'Confirm'.

Verification code: %s

Press Enter to continue...
`

func Login(a auth.Authenticator, ctx context.Context) (*auth.User, error) {
	state, err := a.GetDeviceCode(ctx, http.DefaultClient)
	if err != nil {
		output.PrintErrorDetails("Error getting device code", err)
		return nil, err
	}

	fmt.Printf(loggingInMessage, state.UserCode)

	if _, err = fmt.Scanln(); err != nil {
		output.PrintErrorDetails("Error getting keystroke", err)
		return nil, err
	}

	if err := a.OpenURL(state.VerificationURI); err != nil {
		fmt.Printf(
			"The browser could not be opened automatically, please go to this site to continue\n" +
				"the log-in: " + state.VerificationURI,
		)
	}

	result, err := a.WaitUntilUserLogsIn(ctx, http.DefaultClient, state)
	if err != nil {
		output.PrintErrorDetails("Error waiting for login", err)
		return nil, err
	}

	if err := a.StoreAccessToken(result.AccessToken); err != nil {
		output.PrintErrorDetails("Login failed. Could not store credentials in keyring.", err)
		return nil, err
	}

	if err := a.StoreRefreshToken(result.RefreshToken); err != nil {
		output.PrintError(
			"Error occurred storing refresh token in your keyring.",
			"When your access token expires, you will need to log in again.\n"+
				"Error details: "+err.Error(),
		)
	}

	output.PrintlnOK("You are now logged in to Numerous!")

	return a.GetLoggedInUserFromKeyring(), nil
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
