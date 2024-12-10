package args

import (
	"errors"
	"path/filepath"

	"github.com/spf13/cobra"
	"numerous.com/cli/cmd/output"
)

var ErrOptionalAppDirArgCount = errors.New("there must be at most 1 argument for optional app directory")

// Returns an arguments handler, which checks an optional app dir positional
// argument, and writes the absolute path into the given string reference.
func OptionalAppDir(appDir *string) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			fn := cmd.HelpFunc()
			fn(cmd, args)
			output.PrintError("accepts only an optional [app directory] as a positional argument, you provided %d arguments.", "", len(args))

			return ErrOptionalAppDirArgCount
		}

		// default to empty string
		appDirArg := ""
		if len(args) != 0 {
			appDirArg = args[0]
		}

		// find the absolute path - current working directory if appDirArg is empty
		absAppDir, err := filepath.Abs(appDirArg)
		if err != nil {
			return err
		}

		*appDir = absAppDir

		return nil
	}
}
