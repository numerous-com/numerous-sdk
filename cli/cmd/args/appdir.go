package args

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Returns an arguments handler, which checks an optional app dir positional
// argument, and writes it into the given string reference.
func OptionalAppDir(appDir *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			fn := cmd.HelpFunc()
			fn(cmd, args)

			return fmt.Errorf("accepts only an optional [app directory] as a positional argument, you provided %d arguments", len(args))
		}

		if len(args) == 1 {
			*appDir = args[0]
		}

		return nil
	}
}
