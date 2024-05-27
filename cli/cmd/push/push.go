package push

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"numerous/cli/cmd/initialize"
	"numerous/cli/cmd/output"
	"numerous/cli/dotenv"
	"numerous/cli/internal/archive"
	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"
	"numerous/cli/internal/gql/build"
	"numerous/cli/manifest"
	"numerous/cli/tool"

	"github.com/spf13/cobra"
)

const (
	zipFileName       string      = ".zipped_project.zip"
	zipFilePermission fs.FileMode = 0o666
)

var verbose bool

var (
	carriageReturn       = "\r"
	greenColorEscapeANSI = "\033[32m"
	resetColorEscapeANSI = "\033[0m"
	unicodeCheckmark     = "\u2713"
	greenCheckmark       = carriageReturn + greenColorEscapeANSI + unicodeCheckmark + resetColorEscapeANSI
	unicodeHourglass     = "\u29D6"
)

const (
	ProjectArgLength       = 1
	ProjectAndAppArgLength = 2
)

var PushCmd = &cobra.Command{
	Use:   "push [project path] [app path]",
	Short: "Pushes the app and returns a shareable URL (login required)",
	Long: `Zip-compresses the tool project and pushes it to the numerous server, which
builds a docker image and runs it as a container.
A URL is generated which provides access to the tool, anyone with the URL can access the tool.`,
	Run: push,
}

func push(cmd *cobra.Command, args []string) {
	appDir, projectDir, appPath, ok := parseArguments(args)
	if !ok {
		os.Exit(1)
	}

	toolID, err := tool.ReadAppIDAndPrintErrors(appDir)
	if err != nil {
		os.Exit(1)
	}

	m, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
	if err != nil {
		output.PrintErrorAppNotInitialized()
		os.Exit(1)
	}

	if validated, err := m.ValidateApp(); err != nil {
		output.PrintErrorDetails("An error occurred validating the app", err)
		os.Exit(1)
	} else if !validated {
		os.Exit(1)
	}

	_, err = app.Query(string(toolID), gql.GetClient())
	if err != nil {
		if strings.Contains(err.Error(), "record not found") { // TODO: replace strings-check with GraphQL error type, when GraphQL types exist.
			output.PrintError(
				"Sorry, we can't find that app ID in our database.",
				strings.Join(
					[]string{
						"Please make sure you have the correct app ID entered.",
						"If you have used the \"numerous delete\" command to delete your app, please delete your .app_id",
						"file and reinitialize your app using the \"numerous init\" command.",
					},
					"\n",
				),
			)

			return
		}
	}

	if err := os.Chdir(projectDir); err != nil {
		output.PrintError("Could not access %q", "", projectDir)
		os.Exit(1)
	}

	// Remove if zip already exist
	if file, _ := os.Stat(zipFileName); file != nil {
		os.Remove(zipFileName)
	}

	if ok := prepareApp(m); !ok {
		os.Exit(1)
	}

	buildID, ok := uploadApp(appDir, toolID)
	if !ok {
		os.Exit(1)
	}

	if ok := buildApp(buildID, appPath); !ok {
		os.Exit(1)
	}

	if ok := deployApp(toolID); !ok {
		os.Exit(1)
	}

	if ok := printURL(toolID); !ok {
		os.Exit(1)
	}
}

func printURL(toolID string) (ok bool) {
	pushedTool, err := app.Query(string(toolID), gql.GetClient())
	if err != nil {
		output.PrintErrorDetails("Error reading the app.", err)
		return false
	}

	fmt.Printf("\nShareable url: %s\n", pushedTool.SharedURL)

	return true
}

func deployApp(toolID string) (ok bool) {
	fmt.Print(unicodeHourglass + "  Deploying app......")

	err := stopJobs(string(toolID))
	if err != nil {
		output.PrintErrorDetails("Error stopping previous jobs.", err)
		return false
	}

	err = getDeployEventLogs(string(toolID))
	if err != nil {
		output.PrintErrorDetails("Error listening for deploy logs.", err)
		return false
	}

	fmt.Println(greenCheckmark + "  Deploying app......Done")

	return true
}

func buildApp(buildID string, appPath string) (ok bool) {
	fmt.Print(unicodeHourglass + "  Building app.......")
	if verbose {
		// To allow nice printing of build messages from backend
		fmt.Println()
	}

	err := getBuildEventLogs(buildID, appPath, verbose)
	if err != nil {
		output.PrintErrorDetails("Error listening for build logs.", err)
		return false
	}

	fmt.Println(greenCheckmark + "  Building app.......Done")

	return true
}

func uploadApp(appDir string, toolID string) (buildID string, ok bool) {
	defer os.Remove(zipFileName)

	fmt.Print(unicodeHourglass + "  Uploading app......")

	secrets := loadSecretsFromEnv(appDir)
	buildID, err := pushBuild(zipFileName, string(toolID), secrets)
	if err != nil {
		fmt.Println("Sorry! An error occurred uploading your app")

		if strings.Contains(err.Error(), "server failure: failed to read file for key file") {
			output.PrintError(
				"The app folder is too large.",
				"The maximum size of an app folder is currently 256MB.\n"+
					"If you have large files, which are not needed for your app, consider adding them to the 'exclude' field in 'numerous.toml'",
			)
		} else {
			output.PrintErrorDetails("Error occurred uploading app.", err)
		}

		return "", false
	}

	fmt.Println(greenCheckmark + "  Uploading app......Done")

	return buildID, true
}

func prepareApp(m *manifest.Manifest) (ok bool) {
	if !verbose {
		fmt.Print(unicodeHourglass + "  Preparing upload...")
	}

	zipFile, err := os.OpenFile(zipFileName, os.O_CREATE|os.O_RDWR, zipFilePermission)
	if err != nil {
		output.PrintErrorDetails("Error preparing app.", err)

		return false
	}

	if err := archive.ZipFolder(zipFile, m.Exclude); err != nil {
		output.PrintErrorDetails("Error preparing app.", err)
		os.Remove(zipFileName)

		return false
	}

	zipFile.Close()

	fmt.Println(greenCheckmark + "  Preparing upload...Done")

	return true
}

func parseArguments(args []string) (string, string, string, bool) {
	appDir := "."
	projectDir := "."
	appPath := ""

	if len(args) == ProjectArgLength {
		appDir = args[0]
		projectDir = args[0]
	}

	if len(args) == ProjectAndAppArgLength {
		appDir = args[1]
		projectDir = args[0]
		result, rt, err := CheckAndReturnSubpath(projectDir, appDir)
		if err != nil {
			output.PrintErrorDetails("Error occurred validating app and project arguments.", err)
		}

		if !result {
			output.PrintError("Application path %s is not a subpath of project path %s", "", appDir, projectDir)
			return "", "", "", false
		}
		appPath = rt
	}

	return appDir, projectDir, appPath, true
}

func pushBuild(zipFilePath string, appID string, secrets map[string]string) (string, error) {
	var filePermission fs.FileMode = 0o666
	zipFile, err := os.OpenFile(zipFilePath, os.O_CREATE|os.O_RDWR, filePermission)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()

	build, err := build.Push(zipFile, appID, gql.GetClient(), secrets)
	if err != nil {
		return "", err
	}

	return build.BuildID, nil
}

func loadSecretsFromEnv(appDir string) map[string]string {
	env, _ := dotenv.Load(filepath.Join(appDir, initialize.EnvFileName))
	return env
}

func init() {
	PushCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide more verbose output of the push process")
}
