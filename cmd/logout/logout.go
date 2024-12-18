package logout

import (
	"net/http"

	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/output"
)

func logout(a auth.Authenticator) error {
	user := a.GetLoggedInUserFromKeyring()
	if user == nil {
		output.PrintError("You are not logged in.", "")
		return nil
	}

	_ = a.RevokeRefreshToken(http.DefaultClient, user.RefreshToken)
	if err := a.RemoveLoggedInUserFromKeyring(); err != nil {
		output.PrintUnknownError(err)
		return err
	}

	output.PrintlnOK("You are now logged out.")

	return nil
}
