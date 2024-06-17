package logs

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var LogsCmd = &cobra.Command{
	Use:   "logs [app directory]",
	Run:   run,
	Short: "Deploy an app to an organization.",
	Long: `Read the logs of an application deployed to an organization on the
Numerous platform.

If <name> and <organization> flags are set, they define the app to read logs
from. If they are not the default deployment section in the manifest is used,
if it is defined.

If [app directory] is specified, that directory will be used to read the
app manifest for the default deployment information.

If no [app directory] is specified, the current working directory is used.`,
	Example: `To read the logs from a specific app deployment, use the following form:

    numerous app logs --organization "organization-slug-a2ecf59b" --name "my-app"

Otherwise, assuming an app has been initialized in the directory
"my_project/my_app" and has a default deployment defined in its manifest:

    numerous app logs my_project/my_app
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			fn := cmd.HelpFunc()
			fn(cmd, args)

			return fmt.Errorf("accepts only an optional [app directory] as a positional argument, you provided %d arguments", len(args))
		}

		if len(args) == 1 {
			appDir = args[0]
		}

		return nil
	},
}

var (
	slug       string
	appName    string
	timestamps bool
	appDir     string = "."
)

func run(cmd *cobra.Command, args []string) {
	var printer func(AppDeployLogEntry)
	if timestamps {
		printer = TimestampPrinter
	} else {
		printer = TextPrinter
	}

	if err := Logs(cmd.Context(), nil, appDir, slug, appName, printer); err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func init() {
	flags := LogsCmd.Flags()
	flags.StringVarP(&slug, "organization", "o", "", "The organization slug identifier of the app to read logs from.")
	flags.StringVarP(&appName, "name", "n", "", "The name of the app to read logs from.")
	flags.BoolVarP(&timestamps, "timestamps", "t", false, "Print a timestamp for each log entry.")
}
