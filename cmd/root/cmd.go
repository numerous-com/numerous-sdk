package root

import (
	"errors"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
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
	"numerous.com/cli/internal/logging"
)

var args = struct {
	logLevel logging.Level
}{
	logLevel: logging.LevelError,
}

var cmd = &cobra.Command{
	Use:               "numerous",
	Long:              logo,
	SilenceErrors:     true,
	SilenceUsage:      true,
	PersistentPreRunE: prerun,
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
