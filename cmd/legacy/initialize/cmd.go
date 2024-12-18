package initialize

import (
	"fmt"
	"path/filepath"

	"numerous.com/cli/cmd/args"
	cmdinit "numerous.com/cli/cmd/init"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/wizard"

	"github.com/spf13/cobra"
)

var InitCmd = &cobra.Command{
	Use:     "init [flags] [app directory]",
	Aliases: []string{"initialize"},
	Short:   "Initialize a Numerous project",
	Long:    `Helps the user bootstrap a python project as a Numerous project.`,
	Args:    args.OptionalAppDir(&cmdArgs.appDir),
	Run:     run,
}

var cmdArgs struct {
	name             string
	desc             string
	libraryKey       string
	appFile          string
	requirementsFile string
	appDir           string
}

func run(cmd *cobra.Command, args []string) {
	appDir, err := filepath.Abs(cmdArgs.appDir)
	if err != nil {
		return
	}

	if exists, _ := dir.AppIDExists(appDir); exists {
		printAlreadyInitialized(appDir, fmt.Sprintf("Files %q or %q exists", dir.AppIDFileName, dir.ToolIDFileName))
		return
	}

	if exists, _ := manifest.ManifestExists(appDir); exists {
		printAlreadyInitialized(appDir, fmt.Sprintf("File %q exists", manifest.ManifestFileName))
		return
	}

	params := cmdinit.InitializeParams{
		Name:             cmdArgs.name,
		Desc:             cmdArgs.desc,
		LibraryKey:       cmdArgs.libraryKey,
		AppFile:          cmdArgs.appFile,
		RequirementsFile: cmdArgs.requirementsFile,
		AppDir:           appDir,
	}
	m, err := cmdinit.Initialize(&wizard.SurveyAsker{}, params)
	if err != nil {
		return
	}

	a, err := app.Create(m, gql.GetClient())
	if err != nil {
		output.PrintErrorDetails("Error registering app remotely.", err)
		return
	}

	if err := manifest.BootstrapLegacyApp(appDir, a.ID); err != nil {
		output.PrintErrorDetails("Error writing legacy app ID file with app ID %q", err, a.ID)
	}

	printSuccess(a)
}

func printAlreadyInitialized(appDir, reason string) {
	output.PrintError(
		"An app is already initialized in \"%s\": %s",
		"💡 You can initialize an app in another folder by specifying a\n"+
			"   path in the command, like below:\n\n"+
			"numerous legacy init ./my-app-folder\n\n",
		appDir, reason,
	)
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
	InitCmd.Flags().StringVarP(&cmdArgs.name, "name", "n", "", "Name of the app")
	InitCmd.Flags().StringVarP(&cmdArgs.desc, "description", "d", "", "Description of your app")
	InitCmd.Flags().StringVarP(&cmdArgs.libraryKey, "app-library", "t", "", "Library the app is made with")
	InitCmd.Flags().StringVarP(&cmdArgs.appFile, "app-file", "f", "", "Path to that main file of the project")
	InitCmd.Flags().StringVarP(&cmdArgs.requirementsFile, "requirements-file", "r", "", "Requirements file of the project")
}
