package version

import (
	"github.com/spf13/cobra"
	"numerous.com/cli/internal/version"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of the Numerous CLI",
	Run: func(cmd *cobra.Command, args []string) {
		println("Numerous CLI version " + version.Version)
	},
}
