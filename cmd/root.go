package cmd

import (
	"errors"
	"log/slog"
	"net/http"
	"os"

	"numerous.com/cli/cmd/app"
	"numerous.com/cli/cmd/config"
	"numerous.com/cli/cmd/deletecmd"
	"numerous.com/cli/cmd/deploy"
	"numerous.com/cli/cmd/download"
	"numerous.com/cli/cmd/errorhandling"
	cmdinit "numerous.com/cli/cmd/init"
	"numerous.com/cli/cmd/legacy"
	"numerous.com/cli/cmd/login"
	"numerous.com/cli/cmd/logout"
	"numerous.com/cli/cmd/logs"
	"numerous.com/cli/cmd/organization"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/cmd/token"
	cmdversion "numerous.com/cli/cmd/version"
	"numerous.com/cli/internal/auth"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/logging"
	"numerous.com/cli/internal/version"

	"github.com/spf13/cobra"
)

var (
	ErrNotAuthorized       = errors.New("not authorized")
	ErrIncompatibleVersion = errors.New("incompatible version")
)

var args = struct {
	logLevel logging.Level
}{
	logLevel: logging.LevelError,
}

var cmd = &cobra.Command{
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
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
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
	},
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

func Execute() {
	executedCmd, err := cmd.ExecuteC()
	if err != nil {
		if !errors.Is(err, errorhandling.ErrAlreadyPrinted) {
			output.PrintError("Error: %s", "", err.Error())
			println()
			executedCmd.Usage() // nolint: errcheck
		}
		os.Exit(1)
	}
}

func init() {
	cmd.PersistentFlags().VarP(&args.logLevel, "log-level", "l", "The log level, one of \"debug\", \"info\", \"warning\", or \"error\". Defaults to \"error\".")

	cmd.AddGroup(&cobra.Group{
		Title: "Numerous App Commands:",
		ID:    "app-cmds",
	})
	cmd.AddGroup(&cobra.Group{
		Title: "Additional Numerous Commands:",
		ID:    "additional-cmds",
	})

	cmd.AddCommand(
		cmdinit.Cmd,
		login.Cmd,
		logout.Cmd,
		organization.Cmd,
		legacy.Cmd,
		deletecmd.Cmd,
		deploy.Cmd,
		logs.Cmd,
		download.Cmd,
		token.Cmd,
		cmdversion.Cmd,
		app.Cmd,
		config.Cmd,

		// dummy commands to display helpful messages for legacy commands
		dummyLegacyCmd("push"),
		dummyLegacyCmd("publish"),
		dummyLegacyCmd("unpublish"),
		dummyLegacyCmd("list"),
		dummyLegacyCmd("log"),
	)

	cobra.OnInitialize(func() {
		logOpts := &slog.HandlerOptions{Level: args.logLevel.ToSlogLevel()}
		logHandler := slog.NewTextHandler(os.Stderr, logOpts)
		logger := slog.New(logHandler)
		slog.SetDefault(logger)
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
