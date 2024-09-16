package init

import (
	"errors"
	"fmt"
	"os"

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
	argName             string
	argDesc             string
	argLibraryKey       string
	argAppFile          string
	argRequirementsFile string
)

var (
	ErrGetWorkDir            = errors.New("error getting working directory")
	ErrAppAlreadyInitialized = errors.New("app already initialized")
)

func run(cmd *cobra.Command, args []string) error {
	appDir, err := os.Getwd()
	if err != nil {
		return ErrGetWorkDir
	}

	if len(args) != 0 {
		appDir = PathArgumentHandler(args[0], appDir)
	}

	params := InitializeParams{
		AppDir:           appDir,
		Name:             argName,
		Desc:             argDesc,
		LibraryKey:       argLibraryKey,
		AppFile:          argAppFile,
		RequirementsFile: argRequirementsFile,
	}
	_, err = Initialize(&wizard.SurveyAsker{}, params)
	if errors.Is(err, wizard.ErrStopInit) {
		return nil
	} else if err != nil {
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
	InitCmd.Flags().StringVarP(&argName, "name", "n", "", "Name of the app")
	InitCmd.Flags().StringVarP(&argDesc, "description", "d", "", "Description of your app")
	InitCmd.Flags().StringVarP(&argLibraryKey, "app-library", "t", "", "Library the app is made with")
	InitCmd.Flags().StringVarP(&argAppFile, "app-file", "f", "", "Path to that main file of the project")
	InitCmd.Flags().StringVarP(&argRequirementsFile, "requirements-file", "r", "", "Requirements file of the project")
}
