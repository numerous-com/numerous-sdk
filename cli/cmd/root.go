package cmd

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"numerous/cli/auth"
	"numerous/cli/cmd/app"
	deleteapp "numerous/cli/cmd/delete"
	"numerous/cli/cmd/dev"
	"numerous/cli/cmd/initialize"
	"numerous/cli/cmd/list"
	"numerous/cli/cmd/log"
	"numerous/cli/cmd/login"
	"numerous/cli/cmd/logout"
	"numerous/cli/cmd/organization"
	"numerous/cli/cmd/output"
	"numerous/cli/cmd/publish"
	"numerous/cli/cmd/push"
	"numerous/cli/cmd/report"
	"numerous/cli/cmd/unpublish"
	"numerous/cli/logging"

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
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
		"numerous list",
		"numerous push",
		"numerous log",
		"numerous organization create",
		"numerous organization list",
		"numerous app deploy",
	}

	for _, cmd := range commandsWithAuthRequired {
		if cmd == invokedCommandName {
			return true
		}
	}

	return false
}

func bindCommands() {
	rootCmd.AddCommand(initialize.InitCmd)
	rootCmd.AddCommand(push.PushCmd)
	rootCmd.AddCommand(log.LogCmd)
	rootCmd.AddCommand(deleteapp.DeleteCmd)
	rootCmd.AddCommand(login.LoginCmd)
	rootCmd.AddCommand(logout.LogoutCmd)
	rootCmd.AddCommand(dev.DevCmd)
	rootCmd.AddCommand(publish.PublishCmd)
	rootCmd.AddCommand(unpublish.UnpublishCmd)
	rootCmd.AddCommand(list.ListCmd)
	rootCmd.AddCommand(report.ReportCmd)
	rootCmd.AddCommand(organization.OrganizationRootCmd)
	rootCmd.AddCommand(app.AppRootCmd)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().VarP(&logLevel, "log-level", "l", "The log level, one of \"debug\", \"info\", \"warning\", or \"error\". Defaults to \"error\".")
	bindCommands()
	cobra.OnInitialize(func() {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel.ToSlogLevel()})))
	})
}
