package initialize

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"numerous/cli/cmd/initialize/wizard"
	"numerous/cli/cmd/output"
	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"
	"numerous/cli/tool"

	"github.com/spf13/cobra"
)

var (
	appLibraryString string
	newApp           = tool.Tool{CoverImage: "app_cover.jpg"}
	InitCmd          = &cobra.Command{
		Use:     "init [flags]",
		Aliases: []string{"initialize"},
		Short:   "Initialize a numerous project",
		Long:    `Helps the user bootstrap a python project as a numerous project.`,
		Args:    cobra.MaximumNArgs(1),
		Run:     runInit,
	}
)

func setupFlags(a *tool.Tool) {
	InitCmd.Flags().StringVarP(&a.Name, "name", "n", "", "Name of the app")
	InitCmd.Flags().StringVarP(&a.Description, "description", "d", "", "Description of your app")
	InitCmd.Flags().StringVarP(&appLibraryString, "app-library", "t", "", "Library the app is made with")
	InitCmd.Flags().StringVarP(&a.AppFile, "app-file", "f", "", "Path to that main file of the project")
	InitCmd.Flags().StringVarP(&a.RequirementsFile, "requirements-file", "r", "", "Requirements file of the project")
}

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

	if exist, _ := tool.AppIDExistsInCurrentDir(projectFolderPath); exist {
		output.PrintError(
			"An app is already initialized in %q",
			"💡 You can initialize an app in another folder by specifying a\n"+
				"   path in the command, like below:\n\n"+
				"numerous init ./my-app-folder\n\n",
			projectFolderPath,
		)

		return
	}

	if err := validateAndSetAppLibrary(&newApp, appLibraryString); err != nil {
		fmt.Println(err)
		return
	}

	setPython(&newApp)

	if continueBootstrap, err := wizard.RunInitAppWizard(projectFolderPath, &newApp); err != nil {
		output.PrintErrorDetails("Error running initialization wizard", err)
		return
	} else if !continueBootstrap {
		return
	}

	// Initialize and boostrap project files
	a, err := app.Create(newApp, gql.GetClient())
	if err != nil {
		output.PrintErrorDetails("Error registering app remotely.", err)
		return
	}

	if err := bootstrapFiles(newApp, a.ID, projectFolderPath); err != nil {
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

func validateAndSetAppLibrary(a *tool.Tool, l string) error {
	if l == "" {
		return nil
	}
	lib, err := tool.GetLibraryByKey(l)
	if err != nil {
		return err
	}
	a.Library = lib

	return nil
}

func setPython(a *tool.Tool) {
	fallbackVersion := "3.11"

	if version, err := getPythonVersion(); errors.Is(err, ErrDetectPythonExecutable) {
		fmt.Printf("Python interpeter not found, setting Python version to '%s' for the app.\n", fallbackVersion)
		a.Python = fallbackVersion
	} else if errors.Is(err, ErrDetectPythonVersion) {
		fmt.Printf("Could not parse python version '%s', setting Python version to '%s' for the app.\n", version, fallbackVersion)
		a.Python = fallbackVersion
	} else {
		a.Python = version
	}
}

func printSuccess(a app.App) {
	fmt.Printf(`
🥳 The app has been initialized! 🎉"

💾 The information you entered is now stored in "numerous.toml".
🔍 The App ID %q is stored in %q,
   and is used to identify the app in commands which manage it.
❌ If %q is removed the CLI cannot identify your app, so make
   sure not to delete it!
💡 If you are logged in, you can use "numerous list" to find the
   App ID again.
`, a.ID, tool.AppIDFileName, tool.AppIDFileName)
}

func init() {
	setupFlags(&newApp)
}
