package log

import (
	"os"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"

	"github.com/spf13/cobra"
)

var timestamps = false

var LogCmd = &cobra.Command{
	Use:   "log",
	Short: "Display running application logs",
	Long:  `This command initiates the logging process, providing last hour of application logs for monitoring and troubleshooting purposes.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   log,
}

func log(cmd *cobra.Command, args []string) {
	appDir, err := os.Getwd()
	if err != nil {
		output.PrintUnknownError(err)
		return
	}

	if len(args) > 0 {
		appDir = args[0]
	}

	if err := os.Chdir(appDir); err != nil {
		output.PrintUnknownError(err)
		return
	}

	appID, err := dir.ReadAppID(appDir)
	if err != nil {
		output.PrintReadAppIDErrors(err, appID)
		return
	}

	err = getLogs(appID, timestamps)
	if err != nil {
		output.PrintUnknownError(err)
		return
	}
}

func init() {
	LogCmd.Flags().BoolVarP(&timestamps, "timestamps", "t", false, "Show timestamps for log entries.")
}
