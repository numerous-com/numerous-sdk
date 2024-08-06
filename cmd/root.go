package cmd

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"numerous.com/cli/cmd/deletecmd"
	"numerous.com/cli/cmd/deploy"
	"numerous.com/cli/cmd/dev"
	"numerous.com/cli/cmd/initialize"
	"numerous.com/cli/cmd/legacy"
	"numerous.com/cli/cmd/login"
	"numerous.com/cli/cmd/logout"
	"numerous.com/cli/cmd/logs"
	"numerous.com/cli/cmd/organization"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/logging"

	"github.com/spf13/cobra"
)

var ErrNotAuthorized = errors.New("not authorized")

var (
	logLevel logging.Level = logging.LevelError
	rootCmd                = &cobra.Command{
		Use: "numerous",
		Long: "\n                      ~~~        \n" +
			"            ---       ~~~~~~~      \n" +
			"     °      -------   ~~~~~~~~~~  \n" +
			"     °°°°   ----------- ~~~~~~~~~\n" +
			"     °°°°°°° ----------- ~~~~~~~~       _   _                                              \n" +
			"     °°°°°°°°°°  ------- ~~~~~~~~      | \\ | |                                         \n" +
			"     °°°°°°°°°°°°° -----  ~~~~~~~      |  \\| |_   _ _ __ ___   ___ _ __ ___  _   _ ___\n" +
			"     °°°°°°°°°°°°° -----  ~~~~~~~      | . ` | | | | '_ ` _ \\ / _ \\ '__/ _ \\| | | / __|\n" +
			"     °°°°°°°°°°°°° -----     ~~~~      | |\\  | |_| | | | | | |  __/ | | (_) | |_| \\__ \\\n" +
			"     °°°°°°°°°°°°°  ----       ~~      |_| \\_|\\__,_|_| |_| |_|\\___|_|  \\___/ \\__,_|___/\n" +
			"        °°°°°°°°°°    --\n" +
			"          °°°°°°°°    \n" +
			"             °°°°°   \n" +
			"                °°     \n" +
			"",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			output.NotifyFeedbackMaybe()

			if !commandRequiresAuthentication(cmd.CommandPath()) {
				return nil
			}

			user := auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring()
			if user.CheckAuthenticationStatus() == auth.ErrUserNotLoggedIn {
				output.PrintErrorLoginForCommand(cmd)
				return ErrNotAuthorized
			}

			if err := login.RefreshAccessToken(user, http.DefaultClient, auth.NumerousTenantAuthenticator); err != nil {
				return err
			}

			return nil
		},
	}
)

func commandRequiresAuthentication(invokedCommandName string) bool {
	commandsWithAuthRequired := []string{
		"numerous legacy list",
		"numerous legacy push",
		"numerous legacy log",
		"numerous organization create",
		"numerous organization list",
		"numerous deploy",
		"numerous delete",
		"numerous logs",
	}

	for _, cmd := range commandsWithAuthRequired {
		if cmd == invokedCommandName {
			return true
		}
	}

	return false
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().VarP(&logLevel, "log-level", "l", "The log level, one of \"debug\", \"info\", \"warning\", or \"error\". Defaults to \"error\".")

	rootCmd.AddGroup(&cobra.Group{
		Title: "Numerous App Commands:",
		ID:    "app-cmds",
	})
	rootCmd.AddGroup(&cobra.Group{
		Title: "Additional Numerous Commands:",
		ID:    "additional-cmds",
	})

	rootCmd.AddCommand(initialize.InitCmd,
		login.LoginCmd,
		logout.LogoutCmd,
		dev.DevCmd,
		organization.OrganizationRootCmd,
		legacy.LegacyRootCmd,
		deletecmd.DeleteCmd,
		deploy.DeployCmd,
		logs.LogsCmd,

		// dummy commands to display helpful messages for legacy commands
		dummyLegacyCmd("push"),
		dummyLegacyCmd("publish"),
		dummyLegacyCmd("unpublish"),
		dummyLegacyCmd("list"),
		dummyLegacyCmd("log"),
	)

	cobra.OnInitialize(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel.ToSlogLevel()})))
	})
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
