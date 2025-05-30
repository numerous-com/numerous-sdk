package login

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/output"
)

const loggingInMessage = `You are logging into Numerous.

When you press Enter, a browser window will automatically open.

Please verify that the code in the browser matches the code below to
complete authentication and click 'Confirm'.

Verification code: %s

Press Enter to continue...
`

const emailNotVerifiedErrorMessage = `
Email verification is required to log in.
Please log in to the website to receive a verification email:
  https://www.numerous.com/app/verify
`

func login(a auth.Authenticator, ctx context.Context) (*auth.User, error) {
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
		fmt.Println(
			"The browser could not be opened automatically, please go to this site to continue\n" +
				"the log-in: " + state.VerificationURI,
		)
	}

	result, err := a.WaitUntilUserLogsIn(ctx, http.DefaultClient, state)
	if errors.Is(err, auth.ErrEmailNotVerified) {
		output.PrintError("Email not verified", emailNotVerifiedErrorMessage)
		return nil, err
	} else if err != nil {
		output.PrintErrorDetails("Error waiting for login", err)
		return nil, err
	}

	// Store access token (with automatic fallback)
	if err := a.StoreAccessToken(result.AccessToken); err != nil {
		output.PrintErrorDetails("Failed to store credentials. Both keyring and file-based storage failed.", err)
		return nil, err
	}

	// Store refresh token (with automatic fallback)
	if err := a.StoreRefreshToken(result.RefreshToken); err != nil {
		output.PrintError(
			"Warning: Failed to store refresh token",
			"Your access token was saved, but the refresh token could not be stored.\n"+
				"You may need to log in again when your access token expires.\n"+
				"Error details: "+err.Error(),
		)
		// Don't return error here - access token was stored successfully
	}

	output.PrintlnOK("You are now logged in to Numerous!")

	return a.GetLoggedInUserFromKeyring(), nil
}
