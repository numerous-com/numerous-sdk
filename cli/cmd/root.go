package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"numerous/cli/auth"
	deleteapp "numerous/cli/cmd/delete"
	"numerous/cli/cmd/deploy"
	"numerous/cli/cmd/dev"
	"numerous/cli/cmd/initialize"
	"numerous/cli/cmd/list"
	"numerous/cli/cmd/log"
	"numerous/cli/cmd/login"
	"numerous/cli/cmd/logout"
	"numerous/cli/cmd/organization"
	createorganization "numerous/cli/cmd/organization/create"
	listorganization "numerous/cli/cmd/organization/list"
	"numerous/cli/cmd/publish"
	"numerous/cli/cmd/push"
	"numerous/cli/cmd/report"
	"numerous/cli/cmd/unpublish"
	"numerous/cli/logging"

	"github.com/spf13/cobra"
)

// TODO: add lipgloss lib here instead of using bash and unicode hardcoded! Check cli/appdev/output
const (
	resetPrompt       = "\033[0m"
	cyanBold          = "\033[1;36m"
	raiseHandEmoji    = "\U0000270B"
	shootingStarEmoji = "\U0001F320"
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
				if runtime.GOOS == "windows" {
					fmt.Printf("\"%s\" can only be used when logged in.\n", cmd.CommandPath())
					fmt.Println("Use \"numerous login\" to enable this command.")
				} else {
					fmt.Printf("The use of %s%s%s command can only be done when logged in %s\n", cyanBold, cmd.CommandPath(), resetPrompt, raiseHandEmoji)
					fmt.Printf("To enable it, please first proceed with %snumerous login%s %s\n", cyanBold, resetPrompt, shootingStarEmoji)
				}

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
		"numerous deploy",
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
	rootCmd.AddCommand(deploy.DeployCmd)
	organization.OrganizationRootCmd.AddCommand(createorganization.OrganizationCreateCmd)
	organization.OrganizationRootCmd.AddCommand(listorganization.OrganizationListCmd)
}

func Execute() {
	err := rootCmd.Execute()
	// fmt.Println(auth.NumerousTenantAuthenticator.GetLoggedInUserFromKeyring().AccessToken)
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
