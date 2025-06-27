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

	// Store both tokens using the unified storage interface
	if err := a.StoreBothTokens(result.AccessToken, result.RefreshToken); err != nil {
		if errors.Is(err, auth.ErrUserDeclinedConsent) {
			output.PrintError("Login cancelled", "File storage consent was declined.")

			return nil, err
		}
		output.PrintErrorDetails("Login failed. Could not store credentials.", err)

		return nil, err
	}

	output.PrintlnOK("You are now logged in to Numerous!")

	return a.GetLoggedInUserFromKeyring(), nil
}
