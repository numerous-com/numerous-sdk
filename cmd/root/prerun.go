package root

import (
	"errors"
	"net/http"
	"os"

	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/output"
	cmdversion "numerous.com/cli/cmd/version"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/version"

	"github.com/spf13/cobra"
)

var (
	ErrNotAuthorized       = errors.New("not authorized")
	ErrIncompatibleVersion = errors.New("incompatible version")
)

func prerun(cmd *cobra.Command, args []string) error {
	output.NotifyFeedbackMaybe()

	if !cmdversion.Check(version.NewService(gql.NewClient())) {
		return errorhandling.ErrorAlreadyPrinted(ErrIncompatibleVersion)
	}

	if !commandRequiresAuthentication(cmd.CommandPath()) {
		return nil
	}

	if os.Getenv("NUMEROUS_ACCESS_TOKEN") != "" {
		return nil
	}

	user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
	if user.CheckAuthenticationStatus() == auth.ErrUserNotLoggedIn {
		output.PrintErrorLoginForCommand(cmd)
		return ErrNotAuthorized
	}

	if err := user.RefreshAccessToken(http.DefaultClient, auth.NumerousTenantAuthenticator); err != nil {
		return err
	}

	return nil
}

func commandRequiresAuthentication(invokedCommandName string) bool {
	commandsWithAuthRequired := []string{
		"numerous legacy list",
		"numerous legacy push",
		"numerous legacy log",
		"numerous organization create",
		"numerous organization list",
		"numerous deploy",
		"numerous delete",
		"numerous download",
		"numerous logs",
		"numerous token create",
		"numerous token list",
		"numerous token revoke",
		"numerous app list",
		"numerous app share",
		"numerous app unshare",
	}

	for _, cmd := range commandsWithAuthRequired {
		if cmd == invokedCommandName {
			return true
		}
	}

	return false
}

func dummyLegacyCmd(cmd string) *cobra.Command {
	return &cobra.Command{
		Hidden: true,
		Use:    cmd,
		Run: func(*cobra.Command, []string) {
			output.NotifyCmdMoved("numerous "+cmd, "numerous legacy "+cmd)
		},
	}
}
