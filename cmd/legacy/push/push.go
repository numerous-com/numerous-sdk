package push

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"numerous.com/cli/cmd/errorhandling"
	cmdinit "numerous.com/cli/cmd/init"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/archive"
	"numerous.com/cli/internal/dir"
	"numerous.com/cli/internal/dotenv"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/gql/app"
	"numerous.com/cli/internal/gql/build"
	"numerous.com/cli/internal/manifest"

	"github.com/spf13/cobra"
)

const (
	zipFileName       string      = ".zipped_project.zip"
	zipFilePermission fs.FileMode = 0o666
)

var verbose bool

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
	RunE: func(cmd *cobra.Command, args []string) error {
		err := push(args)
		return errorhandling.ErrorAlreadyPrinted(err)
	},
}

var numerousAppEngineMsg string = `You can solve this by assigning your app definition to this name, for example:
			
@app
class MyApp:
	my_field: str

appdef = MyApp`

func push(args []string) error {
	appDir, projectDir, appPath, err := parseArguments(args)
	if err != nil {
		return err
	}

	toolID, err := dir.ReadAppID(appDir)
	if err != nil {
		output.PrintReadAppIDErrors(err, appDir)

		return err
	}

	m, err := manifest.Load(filepath.Join(appDir, manifest.ManifestPath))
	if err != nil {
		output.PrintErrorAppNotInitialized(appDir)
		output.PrintManifestTOMLError(err)

		return err
	}

	if err := m.ValidateApp(); err != nil {
		if errors.Is(err, manifest.ErrValidateNumerousApp) {
			output.PrintError("Your app file must have an app definition called 'appdef'", numerousAppEngineMsg)
		} else {
			output.PrintErrorDetails("An error occurred validating the app", err)
		}

		return err
	}

	_, err = app.Query(string(toolID), gql.GetClient())
	if err != nil {
		if strings.Contains(err.Error(), "record not found") { // TODO: replace strings-check with GraphQL error type, when GraphQL types exist.
			output.PrintError(
				"Sorry, we can't find the app ID %s in our database.",
				strings.Join(
					[]string{
						"Please make sure you have the correct app ID entered.",
						"If you have used the \"numerous delete\" command to delete your app, please delete your .app_id",
						"file and reinitialize your app using the \"numerous init\" command.",
					},
					"\n",
				),
				output.Highlight(toolID)+output.AnsiRed,
			)

			return err
		}
	}

	if err := os.Chdir(projectDir); err != nil {
		output.PrintError("Could not access %q", "", projectDir)
		return err
	}

	// Remove if zip already exist
	if file, _ := os.Stat(zipFileName); file != nil {
		os.Remove(zipFileName)
	}

	if err := prepareApp(m); err != nil {
		return err
	}

	buildID, err := uploadApp(appDir, toolID)
	if err != nil {
		return err
	}

	if err := buildApp(buildID, appPath); err != nil {
		return err
	}

	if err := deployApp(toolID); err != nil {
		return err
	}

	if err := printURL(toolID); err != nil {
		return err
	}

	return nil
}

func printURL(toolID string) error {
	pushedTool, err := app.Query(string(toolID), gql.GetClient())
	if err != nil {
		output.PrintErrorDetails("Error reading the app.", err)
		return err
	}

	fmt.Printf("\nShareable url: %s\n", pushedTool.SharedURL)

	return nil
}

func deployApp(toolID string) error {
	task := output.StartTask("Deploying app")

	err := stopJobs(string(toolID))
	if err != nil {
		output.PrintErrorDetails("Error stopping previous jobs.", err)
		return err
	}

	w := output.NewTaskLineWriter(task, "Deploy")
	err = getDeployEventLogs(w, string(toolID))
	if err != nil {
		output.PrintErrorDetails("Error listening for deploy logs.", err)
		return err
	}

	task.Done()

	return nil
}

func buildApp(buildID string, appPath string) error {
	task := output.StartTask("Building app")

	w := output.NewTaskLineWriter(task, "Build")
	err := getBuildEventLogs(w, buildID, appPath, verbose)
	if err != nil {
		output.PrintErrorDetails("Error listening for build logs.", err)
		return err
	}
	task.Done()

	return nil
}

func uploadApp(appDir string, toolID string) (buildID string, err error) {
	defer os.Remove(zipFileName)
	task := output.StartTask("Uploading app")

	secrets := loadSecretsFromEnv(appDir)
	buildID, err = pushBuild(zipFileName, string(toolID), secrets)
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

		return "", err
	}

	task.Done()

	return buildID, nil
}

func prepareApp(m *manifest.Manifest) error {
	task := output.StartTask("Preparing upload.")

	if err := archive.ZipCreate(".", zipFileName, m.Exclude); err != nil {
		output.PrintErrorDetails("Error preparing app.", err)
		os.Remove(zipFileName)

		return err
	}

	task.Done()

	return nil
}

var (
	ErrValidateAppAndProjectDirectories = errors.New("error validating app and project directories")
	ErrAppPathNotSubPathOfProjectPath   = errors.New("app path is not a subpath of the project path")
)

func parseArguments(args []string) (string, string, string, error) {
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
			return "", "", "", ErrValidateAppAndProjectDirectories
		}

		if !result {
			output.PrintError("Application path %s is not a subpath of project path %s", "", appDir, projectDir)
			return "", "", "", ErrAppPathNotSubPathOfProjectPath
		}
		appPath = rt
	}

	return appDir, projectDir, appPath, nil
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
	env, _ := dotenv.Load(filepath.Join(appDir, cmdinit.EnvFileName))
	return env
}

func init() {
	PushCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Provide more verbose output of the push process")
}
