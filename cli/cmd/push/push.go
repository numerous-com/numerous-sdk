package push

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"numerous/cli/cmd/output"
	"numerous/cli/internal/gql"
	"numerous/cli/internal/gql/app"
	"numerous/cli/manifest"
	"numerous/cli/tool"

	"github.com/spf13/cobra"
)

const zipFileName string = ".zipped_project.zip"

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
			fmt.Printf("Error occurred validating app and project arguments:\n %s", err)
		}

		if !result {
			fmt.Printf("Error: Application path %s is not a subpath of project path %s", appDir, projectDir)
			return
		}
		appPath = rt
	}

	toolID, err := tool.ReadAppIDAndPrintErrors(appDir)
	if err != nil {
		return
	}

	m, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
	if err != nil {
		output.PrintErrorAppNotInitialized()
		return
	}

	if validated, err := m.ValidateApp(); err != nil {
		fmt.Printf("An error occurred validating the app: %s", err)
		os.Exit(1)
	} else if !validated {
		os.Exit(1)
	}

	_, err = app.Query(string(toolID), gql.GetClient())
	if err != nil {
		if strings.Contains(err.Error(), "record not found") { // TODO: replace strings-check with GraphQL error type, when GraphQL types exist.
			fmt.Println("Sorry, we can't find that app ID in our database. Please make sure you have the correct app ID entered.",
				"\nIf you have used the \"numerous delete\" command to delete your app, please delete your .app_id",
				"\nfile and reinitialize your app using the \"numerous init\" command.")

			return
		}
	}

	if err := os.Chdir(projectDir); err != nil {
		fmt.Printf("Could not access \"%s\"", projectDir)
		return
	}

	// Remove if zip already exist
	if file, _ := os.Stat(zipFileName); file != nil {
		os.Remove(zipFileName)
	}

	if !verbose {
		fmt.Print(unicodeHourglass + "  Preparing upload...")
	}
	var filePermission fs.FileMode = 0o666
	zipFile, err := os.OpenFile(zipFileName, os.O_CREATE|os.O_RDWR, filePermission)
	if err != nil {
		fmt.Printf("Error preparing app.\nError: %s", err)
		return
	}
	defer os.Remove(zipFileName)

	if err := ZipFolder(zipFile, m.Exclude); err != nil {
		fmt.Printf("Error preparing app.\nError: %s", err)
		return
	}
	zipFile.Close()

	fmt.Println(greenCheckmark + "  Preparing upload...Done")
	fmt.Print(unicodeHourglass + "  Uploading app......")

	buildID, err := uploadZipFile(zipFileName, string(toolID))
	if err != nil {
		fmt.Println("Sorry! An error occurred uploading your app")

		if strings.Contains(err.Error(), "server failure: failed to read file for key file") {
			fmt.Println("The app folder is too large. The maximum size of an app folder is currently 256MB.")
			fmt.Println("If you have large files, which are not needed for your app, consider adding them to the 'exclude' field in 'numerous.toml'")
		} else {
			fmt.Printf("Error in uploading app.\nError: %s", err)
		}

		return
	}

	fmt.Println(greenCheckmark + "  Uploading app......Done")

	fmt.Print(unicodeHourglass + "  Building app.......")
	if verbose { // To allow nice printing of build messages from backend
		fmt.Println()
	}

	err = getBuildEventLogs(buildID, appPath, verbose)
	if err != nil {
		fmt.Printf("Error listening for build logs.\nError: %s", err)
		return
	}

	fmt.Println(greenCheckmark + "  Building app.......Done")
	fmt.Print(unicodeHourglass + "  Deploying app......")

	err = stopJobs(string(toolID))
	if err != nil {
		fmt.Printf("Error stopping previous jobs.\nError: %s", err)
		return
	}

	err = getDeployEventLogs(string(toolID))
	if err != nil {
		fmt.Printf("Error listening for deploy logs.\nError: %s", err)
		return
	}

	fmt.Println(greenCheckmark + "  Deploying app......Done")

	pushedTool, err := app.Query(string(toolID), gql.GetClient())
	if err != nil {
		fmt.Printf("Error reading the app.\nError: %s\n", err)
		return
	}
	fmt.Printf("\nShareable url: %s\n", pushedTool.SharedURL)
}

func init() {
	PushCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide more verbose output of the push process")
}
