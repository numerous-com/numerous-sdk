package init

import (
	"errors"
	"fmt"

	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/wizard"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:     "init [flags]",
	Aliases: []string{"initialize"},
	GroupID: group.AppCommandsGroupID,
	Short:   "Initialize a Numerous project",
	Long:    `Helps the user bootstrap a python project as a numerous project.`,
	Args:    cobra.MaximumNArgs(1),
	RunE:    run,
}

var (
	name             string
	desc             string
	libraryKey       string
	appFile          string
	requirementsFile string
)

var (
	ErrGetWorkDir            = errors.New("error getting working directory")
	ErrAppAlreadyInitialized = errors.New("app already initialized")
)

func run(cmd *cobra.Command, args []string) error {
	appDir, m, err := PrepareInit(args)
	if errors.Is(err, wizard.ErrStopInit) {
		return nil
	} else if err != nil {
		return err
	}

	err = m.BootstrapFiles("", appDir)
	switch {
	case err == manifest.ErrEncodingManifest:
		output.PrintErrorDetails("Error encoding manifest file", err)
	case err != nil:
		output.PrintErrorDetails("Error bootstrapping files.", err)
		return err
	}

	printSuccess()

	return nil
}

func printSuccess() {
	fmt.Printf(`
The app has been initialized! 🎉

The information you entered is now stored in %s.

Next steps:
 1. Use %s to login to Numerous.
 2. Use %s to manage the organizations to deploy apps to.
 3. Use %s to deploy your app.
`,
		output.Highlight(manifest.ManifestFileName),
		output.Highlight("numerous login"),
		output.Highlight("numerous organization"),
		output.Highlight("numerous deploy"),
	)
}

func init() {
	InitCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the app")
	InitCmd.Flags().StringVarP(&desc, "description", "d", "", "Description of your app")
	InitCmd.Flags().StringVarP(&libraryKey, "app-library", "t", "", "Library the app is made with")
	InitCmd.Flags().StringVarP(&appFile, "app-file", "f", "", "Path to that main file of the project")
	InitCmd.Flags().StringVarP(&requirementsFile, "requirements-file", "r", "", "Requirements file of the project")
}
