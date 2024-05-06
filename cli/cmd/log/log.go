package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"numerous/cli/tool"

	"github.com/spf13/cobra"
)

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Display running application logs",
	Long:  `This command initiates the logging process, providing last hour of application logs for monitoring and troubleshooting purposes.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   log,
}

func log(cmd *cobra.Command, args []string) {
	msg := "Now streaming log entries from the last hour and all new entries..."
	fmt.Println(msg)
	userDir, err := os.Getwd()
	if err != nil {
		slog.Info("An error occurred when trying to get the current user path with log command.", slog.String("error", err.Error()))
		fmt.Println(err)

		return
	}
	if len(args) > 0 {
		userDir = args[0]
	}

	if err := os.Chdir(userDir); err != nil {
		fmt.Printf("Could not access \"%s\"", userDir)
		return
	}

	appID, err := os.ReadFile(filepath.Join(userDir, tool.ToolIDFileName))
	if err != nil {
		slog.Info("An error occurred when trying read tool id file.", slog.String("error", err.Error()))
		fmt.Println(tool.ErrAppIDNotFound)
		fmt.Println("Remember to be in the app directory or pass it as an argument to the numerous log command!")

		return
	}

	err = getLogs(string(appID))
	if err != nil {
		fmt.Println("Error listening for logs.", err)
	}
}
