package init

import (
	"errors"
	"fmt"
	"path/filepath"

	"numerous.com/cli/cmd/args"
	"numerous.com/cli/cmd/errorhandling"
	"numerous.com/cli/cmd/group"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/wizard"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "init [app directory]",
	Aliases: []string{"initialize"},
	GroupID: group.AppCommandsGroupID,
	Short:   "Initialize a Numerous project",
	Long:    `Helps the user bootstrap a python project as a numerous project.`,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
	RunE:    run,
}

var cmdArgs struct {
	name                 string
	description          string
	libraryKey           string
	appFilePath          string
	requirementsFilePath string
	dockerfilePath       string
	dockerContextPath    string
	dockerPort           uint
	appDir               string
}

var (
	ErrGetWorkDir            = errors.New("error getting working directory")
	ErrAppAlreadyInitialized = errors.New("app already initialized")
)

func run(cmd *cobra.Command, args []string) error {
	absAppDir, err := filepath.Abs(cmdArgs.appDir)
	if err != nil {
		return ErrGetWorkDir
	}

	params := InitializeParams{
		AppDir:           absAppDir,
		Name:             cmdArgs.name,
		Desc:             cmdArgs.description,
		LibraryKey:       cmdArgs.libraryKey,
		AppFile:          cmdArgs.appFilePath,
		RequirementsFile: cmdArgs.requirementsFilePath,
		Dockerfile:       cmdArgs.dockerfilePath,
		DockerContext:    cmdArgs.dockerContextPath,
		DockerPort:       int(cmdArgs.dockerPort),
	}
	_, err = Initialize(&wizard.SurveyAsker{}, params)
	if errors.Is(err, wizard.ErrStopInit) {
		return nil
	} else if err != nil {
		return errorhandling.ErrAlreadyPrinted
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
	Cmd.Flags().StringVarP(&cmdArgs.name, "name", "n", "", "Name of the app")
	Cmd.Flags().StringVarP(&cmdArgs.description, "description", "d", "", "Description of the app")
	Cmd.Flags().StringVarP(&cmdArgs.libraryKey, "app-library", "t", "", "Library the app is made with")

	Cmd.Flags().StringVarP(&cmdArgs.appFilePath, "app-file", "f", "", "Path to the entrypoint module of the python app")
	Cmd.Flags().StringVarP(&cmdArgs.requirementsFilePath, "requirements-file", "r", "", "Path to the requirements file of the python app")

	Cmd.Flags().StringVar(&cmdArgs.dockerfilePath, "dockerfile", "", "Path to the Dockerfile for the app")
	Cmd.Flags().StringVar(&cmdArgs.dockerContextPath, "docker-context", "", "Path used as the context for building the app Dockerfile")
	Cmd.Flags().UintVar(&cmdArgs.dockerPort, "docker-port", 0, "The port exposed in the Dockerfile")
}
