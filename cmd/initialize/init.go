package initialize

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"numerous.com/cli/cmd/initialize/wizard"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/python"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:     "init [flags]",
	Aliases: []string{"initialize"},
	Short:   "Initialize a numerous project",
	Long:    `Helps the user bootstrap a python project as a numerous project.`,
	Args:    cobra.MaximumNArgs(1),
	Run:     runInit,
}

var (
	name             string
	desc             string
	libraryKey       string
	appFile          string
	requirementsFile string
)

func runInit(cmd *cobra.Command, args []string) {
	projectFolderPath, err := os.Getwd()
	if err != nil {
		slog.Info("An error occurred when trying to get the current user path during init process.", slog.String("error", err.Error()))
		fmt.Println(err)

		return
	}

	if len(args) != 0 {
		projectFolderPath = pathArgumentHandler(args[0], projectFolderPath)
	}

	if exist, _ := dir.AppIDExists(projectFolderPath); exist {
		output.PrintError(
			"An app is already initialized in \"%s\"",
			"💡 You can initialize an app in another folder by specifying a\n"+
				"   path in the command, like below:\n\n"+
				"numerous init ./my-app-folder\n\n",
			projectFolderPath,
		)

		return
	}

	lib, err := manifest.GetLibraryByKey(libraryKey)
	if libraryKey != "" && err != nil {
		output.PrintErrorDetails("Unsupported library", err)
		os.Exit(1)
	}

	pythonVersion := python.PythonVersion()

	m := manifest.New(lib, name, desc, pythonVersion, appFile, requirementsFile)
	if continueBootstrap, err := wizard.RunInitAppWizard(projectFolderPath, m); err != nil {
		output.PrintErrorDetails("Error running initialization wizard", err)
		return
	} else if !continueBootstrap {
		return
	}

	// Initialize and boostrap project files
	a, err := app.Create(m, gql.GetClient())
	if err != nil {
		output.PrintErrorDetails("Error registering app remotely.", err)
		return
	}

	if err := bootstrapFiles(m, a.ID, projectFolderPath); err != nil {
		output.PrintErrorDetails("Error bootstrapping files.", err)
		return
	}

	printSuccess(a)
}

func pathArgumentHandler(providedPath string, currentPath string) string {
	appPath := providedPath
	if providedPath != "." {
		pathBegin := string([]rune(providedPath)[0:2])
		if pathBegin == "./" || pathBegin == ".\\" {
			appPath = strings.Replace(appPath, ".", currentPath, 1)
		} else {
			appPath = providedPath
		}
	} else {
		appPath = currentPath
	}

	return appPath
}

func printSuccess(a app.App) {
	fmt.Printf(`
The app has been initialized! 🎉

The information you entered is now stored in "numerous.toml".
The App ID %q is stored in %q and is used to identify the app in commands which manage it.

If %q is removed, the CLI cannot identify your app. 

If you are logged in, you can use numerous list to find the App ID again.
`, a.ID, dir.AppIDFileName, dir.AppIDFileName)
}

func init() {
	InitCmd.Flags().StringVarP(&name, "name", "n", "", "Name of the app")
	InitCmd.Flags().StringVarP(&desc, "description", "d", "", "Description of your app")
	InitCmd.Flags().StringVarP(&libraryKey, "app-library", "t", "", "Library the app is made with")
	InitCmd.Flags().StringVarP(&appFile, "app-file", "f", "", "Path to that main file of the project")
	InitCmd.Flags().StringVarP(&requirementsFile, "requirements-file", "r", "", "Requirements file of the project")
}
