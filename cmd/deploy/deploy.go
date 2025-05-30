package deploy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"numerous.com/cli/cmd/logs"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/archive"
	"numerous.com/cli/internal/dotenv"
	"numerous.com/cli/internal/links"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/output"
)

const maxUploadBytes int64 = 5368709120

var (
	errArchiveTooLarge   = fmt.Errorf("archive exceeds maximum size %d bytes", maxUploadBytes)
	errAppDirOutOfBounds = errors.New("app directory out of bounds of project directory")
)

type deployBuildError struct {
	Message string
}

func (e *deployBuildError) Error() string {
	return e.Message
}

type appService interface {
	ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error)
	Create(ctx context.Context, input app.CreateAppInput) (app.CreateAppOutput, error)
	CreateVersion(ctx context.Context, input app.CreateAppVersionInput) (app.CreateAppVersionOutput, error)
	AppVersionUploadURL(ctx context.Context, input app.AppVersionUploadURLInput) (app.AppVersionUploadURLOutput, error)
	UploadAppSource(uploadURL string, archive app.UploadArchive) error
	DeployApp(ctx context.Context, input app.DeployAppInput) (app.DeployAppOutput, error)
	DeployEvents(ctx context.Context, input app.DeployEventsInput) error
	AppDeployLogs(appident.AppIdentifier, *int, bool) (chan app.AppDeployLogEntry, error)
}

type deployInput struct {
	appDir     string
	projectDir string
	orgSlug    string
	appSlug    string
	version    string
	message    string
	verbose    bool
	follow     bool
	dryRun     bool
}

func deploy(ctx context.Context, apps appService, input deployInput) error {
	// Determine if it's an App or Task deployment
	taskManifestPath := filepath.Join(input.appDir, "numerous-task.toml")
	appManifestPath := filepath.Join(input.appDir, manifest.ManifestFileName)

	var isTaskDeployment bool
	if _, err := os.Stat(taskManifestPath); err == nil {
		isTaskDeployment = true
		if input.dryRun {
			output.PrintlnOK("DRY RUN: Found numerous-task.toml, would proceed with Task Collection deployment.")
		} else {
			output.PrintlnOK("Found numerous-task.toml, proceeding with Task Collection deployment.")
		}
	} else if _, err := os.Stat(appManifestPath); err == nil {
		isTaskDeployment = false
		if input.dryRun {
			output.PrintlnOK("DRY RUN: Found numerous.toml, would proceed with App deployment.")
		}
		// Note: No else statement for non-dry-run app case to avoid extra output
	} else {
		// Neither manifest found
		output.PrintError("No numerous.toml (for an App) or numerous-task.toml (for a Task Collection) found in directory %q.", "", input.appDir)
		return fmt.Errorf("no app or task manifest found in %s", input.appDir)
	}

	if isTaskDeployment {
		return deployTaskCollection(ctx, input)
	}

	// Existing app deployment logic
	return deployApp(ctx, apps, input)
}

func deployApp(ctx context.Context, apps appService, input deployInput) error {
	if input.dryRun {
		return dryRunApp(input)
	}

	appRelativePath, err := findAppRelativePath(input)
	if err != nil {
		output.PrintError("Project directory %q must be a parent of app directory %q", "", input.projectDir, input.appDir)
		return err
	}

	manifest, secrets, err := loadAppConfiguration(input)
	if err != nil {
		return err
	}

	appVersionOutput, orgSlug, appSlug, err := registerAppVersion(ctx, apps, input, manifest)
	if err != nil {
		return err
	}

	archive, err := createAppArchive(input, manifest)
	if err != nil {
		return err
	}

	if err := uploadAppArchive(ctx, apps, archive, appVersionOutput.AppVersionID); err != nil {
		return err
	}

	if err := archive.Close(); err != nil {
		slog.Error("Error closing temporary app archive", slog.String("error", err.Error()))
	}

	if err := os.Remove(archive.Name()); err != nil {
		slog.Error("Error removing temporary app archive", slog.String("error", err.Error()))
	}

	if err := deployAppFinal(ctx, appVersionOutput, secrets, apps, input, appRelativePath); err != nil {
		return err
	}

	output.PrintlnOK("Access your app at: " + links.GetAppURL(orgSlug, appSlug))

	if input.follow {
		output.Notify("Following logs of %s/%s:", "", orgSlug, appSlug)
		if err := followLogs(ctx, apps, orgSlug, appSlug); err != nil {
			return err
		}
	} else {
		fmt.Println()
		fmt.Println("To read the logs from your app you can:")
		fmt.Println("  " + output.Highlight("numerous logs --organization="+orgSlug+" --app="+appSlug))
		fmt.Println("Or you can use the " + output.Highlight("--follow") + " flag:")

		projectDirArg := ""
		if input.projectDir != "" {
			projectDirArg = " --project-dir=" + input.projectDir
		}

		appDirArg := ""
		if input.appDir != "" {
			appDirArg = " " + input.appDir
		}

		fmt.Println("  " + output.Highlight("numerous deploy --follow --organization="+orgSlug+" --app="+appSlug+projectDirArg+appDirArg))
	}

	return nil
}

func findAppRelativePath(input deployInput) (string, error) {
	if input.projectDir == "" {
		return "", nil
	}

	absProjectDir, err := filepath.Abs(input.projectDir)
	if err != nil {
		return "", err
	}

	absAppDir, err := filepath.Abs(input.appDir)
	if err != nil {
		return "", err
	}

	rel, err := filepath.Rel(absProjectDir, absAppDir)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", errAppDirOutOfBounds
	}

	return filepath.ToSlash(rel), nil
}

func loadAppConfiguration(input deployInput) (*manifest.Manifest, map[string]string, error) {
	task := output.StartTask("Loading app configuration")
	m, err := manifest.Load(filepath.Join(input.appDir, manifest.ManifestFileName))
	if err != nil {
		task.Error()
		output.PrintErrorAppNotInitialized(input.appDir)
		output.PrintManifestTOMLError(err)

		return nil, nil, err
	}

	secrets := loadSecretsFromEnv(input.appDir)

	// for validation
	ai, err := appident.GetAppIdentifier(input.appDir, m, input.orgSlug, input.appSlug)
	if err != nil {
		task.Error()
		appident.PrintGetAppIdentifierError(err, input.appDir, ai)

		return nil, nil, err
	}

	task.Done()

	return m, secrets, nil
}

func createAppArchive(input deployInput, manifest *manifest.Manifest) (*os.File, error) {
	srcPath := input.appDir
	if input.projectDir != "" {
		srcPath = input.projectDir
	}

	task := output.StartTask("Creating app archive")
	archivePath := path.Join(srcPath, ".tmp_app_archive.tar")

	if err := archive.TarCreate(srcPath, archivePath, manifest.Exclude); err != nil {
		task.Error()
		output.PrintErrorDetails("Error archiving app source", err)

		return nil, err
	}

	archive, err := os.Open(archivePath)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app source archive", err)
		os.Remove(archivePath) // nolint: errcheck

		return nil, err
	}
	task.Done()

	return archive, nil
}

func registerAppVersion(ctx context.Context, apps appService, input deployInput, manifest *manifest.Manifest) (app.CreateAppVersionOutput, string, string, error) {
	ai, err := appident.GetAppIdentifier("", manifest, input.orgSlug, input.appSlug)
	if err != nil {
		appident.PrintGetAppIdentifierError(err, input.appDir, ai)
		return app.CreateAppVersionOutput{}, "", "", err
	}

	task := output.StartTask("Registering new version for " + ai.OrganizationSlug + "/" + ai.AppSlug)
	appID, err := readOrCreateApp(ctx, apps, ai, manifest)
	if err != nil {
		task.Error()
		switch {
		case errors.Is(err, app.ErrAccessDenied):
			app.PrintErrorAccessDenied(ai)
		case !errors.Is(err, app.ErrAppNotFound):
			output.PrintErrorDetails("Error reading remote app", err)
		}

		return app.CreateAppVersionOutput{}, "", "", err
	}

	appVersionInput := app.CreateAppVersionInput{AppID: appID, Version: input.version, Message: input.message, Size: manifest.Size}
	appVersionOutput, err := apps.CreateVersion(ctx, appVersionInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app version remotely", err)

		return app.CreateAppVersionOutput{}, "", "", err
	}
	task.Done()

	return appVersionOutput, ai.OrganizationSlug, ai.AppSlug, nil
}

func readOrCreateApp(ctx context.Context, apps appService, ai appident.AppIdentifier, manifest *manifest.Manifest) (string, error) {
	appReadInput := app.ReadAppInput{
		OrganizationSlug: ai.OrganizationSlug,
		AppSlug:          ai.AppSlug,
	}
	appReadOutput, err := apps.ReadApp(ctx, appReadInput)
	if err == nil {
		return appReadOutput.AppID, nil
	} else if !errors.Is(err, app.ErrAppNotFound) {
		return "", err
	}

	appCreateInput := app.CreateAppInput{
		OrganizationSlug: ai.OrganizationSlug,
		AppSlug:          ai.AppSlug,
		DisplayName:      manifest.Name,
		Description:      manifest.Description,
	}
	appCreateOutput, err := apps.Create(ctx, appCreateInput)
	if err != nil {
		output.PrintErrorDetails("Error creating app remotely", err)
		return "", err
	}

	return appCreateOutput.AppID, nil
}

type progressReader struct {
	r         io.Reader
	bytesSent int
	task      *output.Task
	totalSize int64
}

func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.r.Read(p)
	pr.bytesSent += n
	pr.report(errors.Is(err, io.EOF))

	return n, err
}

func (pr *progressReader) report(eof bool) {
	if eof {
		pr.task.Progress(100.0) // nolint:mnd
		return
	}

	percent := float32(0.0)
	if pr.bytesSent >= 0 {
		percent = 100.0 * float32(pr.bytesSent) / float32(pr.totalSize) // nolint:mnd
	}

	pr.task.Progress(percent)
}

func uploadAppArchive(ctx context.Context, apps appService, archive *os.File, appVersionID string) error {
	task := output.StartTask("Uploading app archive")
	uploadURLInput := app.AppVersionUploadURLInput(app.AppVersionUploadURLInput{AppVersionID: appVersionID})
	uploadURLOutput, err := apps.AppVersionUploadURL(ctx, uploadURLInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app version remotely", err)

		return err
	}

	stat, err := archive.Stat()
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error checking archive size", err)

		return err
	} else if stat.Size() > maxUploadBytes {
		task.Error()
		printAppSourceArchiveTooLarge(stat.Size())

		return errArchiveTooLarge
	}

	uploadArchive := app.UploadArchive{
		Reader: &progressReader{r: archive, totalSize: stat.Size(), task: task},
		Size:   stat.Size(),
	}

	err = apps.UploadAppSource(uploadURLOutput.UploadURL, uploadArchive)
	var appSourceUploadErr *app.AppSourceUploadError
	if errors.As(err, &appSourceUploadErr) {
		task.Error()
		printAppSourceUploadErr(appSourceUploadErr)

		return err
	} else if err != nil {
		task.Error()
		output.PrintErrorDetails("Error uploading app source archive", err)

		return err
	}
	task.Done()

	return nil
}

const appSourceArchiveTooLargeErrMsg = `App archive has size %s, but the maximum allowed size for an archived app is %s.

You can exclude files in the app folder from being uploaded by adding patterns to the "exclude" field in %s, e.g.:
  exclude = ["venv*", "*venv", "./my-test-data-folder"]

Read more at:
  https://www.numerous.com/docs/cli#exclude-certain-files-and-folders
`

func printAppSourceArchiveTooLarge(size int64) {
	output.PrintError(
		"App archive too large to upload",
		appSourceArchiveTooLargeErrMsg,
		humanizeBytes(size),
		humanizeBytes(maxUploadBytes),
		manifest.ManifestFileName,
	)
}

func humanizeBytes(bytes int64) string {
	var KB int64 = 1024
	MB := KB * KB
	GB := KB * KB * KB

	switch {
	case bytes > GB:
		return fmt.Sprintf("%.2fG", float64(bytes)/float64(GB))
	case bytes > MB:
		return fmt.Sprintf("%.2fM", float64(bytes)/float64(MB))
	case bytes > KB:
		return fmt.Sprintf("%.2fK", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%dB", bytes)
	}
}

const appSourceUploadErrMsg string = `When uploading the app source archive, the file storage server responded with an error.
  HTTP Status %d: %q
  Upload URL: %s

--- Response body:
%s
`

func printAppSourceUploadErr(appSourceUploadErr *app.AppSourceUploadError) {
	output.PrintError("Error uploading app source archive",
		appSourceUploadErrMsg,
		appSourceUploadErr.HTTPStatusCode,
		appSourceUploadErr.HTTPStatus,
		appSourceUploadErr.UploadURL,
		string(appSourceUploadErr.ResponseBody),
	)
}

func deployAppFinal(ctx context.Context, appVersionOutput app.CreateAppVersionOutput, secrets map[string]string, apps appService, input deployInput, appRelativePath string) error {
	task := output.StartTask("Deploying app")

	deployAppInput := app.DeployAppInput{AppVersionID: appVersionOutput.AppVersionID, Secrets: secrets, AppRelativePath: appRelativePath}
	deployAppOutput, err := apps.DeployApp(ctx, deployAppInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error deploying app", err)

		return err
	}

	appDeploymentStatusEventUpdater := statusUpdater{verbose: input.verbose, task: task}
	eventsInput := app.DeployEventsInput{
		DeploymentVersionID: deployAppOutput.DeploymentVersionID,
		Handler: func(de app.DeployEvent) error {
			switch de.Typename {
			case "AppBuildMessageEvent":
				if input.verbose {
					for _, l := range strings.Split(de.BuildMessage.Message, "\n") {
						task.AddLine("Build", l)
					}
				}
			case "AppBuildErrorEvent":
				if input.verbose {
					for _, l := range strings.Split(de.BuildError.Message, "\n") {
						task.AddLine("Error", l)
					}
				}

				return &deployBuildError{Message: de.BuildError.Message}
			case "AppDeploymentStatusEvent":
				if err := appDeploymentStatusEventUpdater.update(de.DeploymentStatus.Status); err != nil {
					return err
				}
			}

			return nil
		},
	}

	err = apps.DeployEvents(ctx, eventsInput)
	if err != nil {
		var buildError *deployBuildError
		task.Error()
		if errors.As(err, &buildError) {
			output.PrintError("Build error", buildError.Message)
		} else {
			output.PrintErrorDetails("Error occurred during deploy", err)
		}

		return err
	}
	task.Done()

	return nil
}

type statusUpdater struct {
	verbose               bool
	lastStatus            *string
	sameStatusUpdateCount uint
	task                  *output.Task
}

func (s *statusUpdater) update(status string) error {
	if s.verbose {
		if s.lastStatus != nil && *s.lastStatus != status {
			s.task.EndUpdateLine()
		}

		isDifferent := s.lastStatus == nil || *s.lastStatus != status
		if isDifferent {
			s.sameStatusUpdateCount = 0
		} else {
			s.sameStatusUpdateCount++
		}

		s.lastStatus = &status
		s.task.UpdateLine("Deploy", "Workload is "+strings.ToLower(status)+strings.Repeat(".", int(s.sameStatusUpdateCount)))
	}

	switch status {
	case "PENDING", "RUNNING":
	default:
		return fmt.Errorf("got status %s while deploying", status)
	}

	return nil
}

func loadSecretsFromEnv(appDir string) map[string]string {
	env, _ := dotenv.Load(path.Join(appDir, manifest.EnvFileName))
	return env
}

func followLogs(ctx context.Context, apps appService, orgSlug, appSlug string) error {
	ai := appident.AppIdentifier{OrganizationSlug: orgSlug, AppSlug: appSlug}
	ch, err := apps.AppDeployLogs(ai, nil, true)
	if err != nil {
		app.PrintAppError(err, ai)
		return err
	}

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return nil
			}
			logs.TimestampPrinter(entry)
		case <-ctx.Done():
			return nil
		}
	}
}

func dryRunApp(input deployInput) error {
	task := output.StartTask("DRY RUN: Analyzing app configuration")

	// Load and validate app configuration (this is safe to do in dry-run)
	manifest, _, err := loadAppConfiguration(input)
	if err != nil {
		task.Error()
		return err
	}

	// Get app identifier
	ai, err := appident.GetAppIdentifier(input.appDir, manifest, input.orgSlug, input.appSlug)
	if err != nil {
		task.Error()
		appident.PrintGetAppIdentifierError(err, input.appDir, ai)
		return err
	}

	task.Done()

	// Show what would be deployed
	fmt.Println()
	output.PrintlnOK("=== DRY RUN: App Deployment Summary ===")
	fmt.Printf("App Name: %s\n", output.Highlight(manifest.Name))
	fmt.Printf("Description: %s\n", manifest.Description)
	fmt.Printf("Organization: %s\n", output.Highlight(ai.OrganizationSlug))
	fmt.Printf("App Slug: %s\n", output.Highlight(ai.AppSlug))
	fmt.Printf("Directory: %s\n", input.appDir)

	if input.version != "" {
		fmt.Printf("Version: %s\n", input.version)
	}
	if input.message != "" {
		fmt.Printf("Message: %s\n", input.message)
	}

	// Show app type
	if manifest.Python != nil {
		fmt.Printf("App Type: Python (%s)\n", manifest.Python.Library.Key)
		if manifest.Python.AppFile != "" {
			fmt.Printf("App File: %s\n", manifest.Python.AppFile)
		}
		if manifest.Python.RequirementsFile != "" {
			fmt.Printf("Requirements File: %s\n", manifest.Python.RequirementsFile)
		}
	}
	if manifest.Docker != nil {
		fmt.Printf("App Type: Docker\n")
		if manifest.Docker.Dockerfile != "" {
			fmt.Printf("Dockerfile: %s\n", manifest.Docker.Dockerfile)
		}
	}

	// Show exclusions
	if len(manifest.Exclude) > 0 {
		fmt.Printf("Excluded Files: %v\n", manifest.Exclude)
	}

	fmt.Println()
	output.PrintlnOK("DRY RUN: Would deploy app to Numerous platform")
	return nil
}

// TaskManifestCollection represents the structure of numerous-task.toml
type TaskManifestCollection struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	// Task definitions
	Task []TaskDefinition `toml:"task"`
	// Environment configuration (reuse app manifest structures)
	Python *TaskPython `toml:"python,omitempty"`
	Docker *TaskDocker `toml:"docker,omitempty"`
	// Optional deployment information
	Deployment *TaskDeployment `toml:"deployment,omitempty"`
}

type TaskDefinition struct {
	FunctionName      string `toml:"function_name"`
	SourceFile        string `toml:"source_file"`
	DecoratedFunction string `toml:"decorated_function,omitempty"`
	Description       string `toml:"description,omitempty"`
	// For Docker collections - specify how to run this specific task
	Entrypoint []string `toml:"entrypoint,omitempty"`
	// For API collections - endpoint for this task
	APIEndpoint string `toml:"api_endpoint,omitempty"`
	// Optional Python stub for type checking
	PythonStub string `toml:"python_stub,omitempty"`
}

// Reuse app manifest Docker structure exactly
type TaskDocker struct {
	Dockerfile string `toml:"dockerfile,omitempty"`
	Context    string `toml:"context,omitempty"`
}

// Python environment for task collections (similar to apps)
type TaskPython struct {
	Version          string `toml:"version"`
	RequirementsFile string `toml:"requirements_file"`
}

type TaskDeployment struct {
	OrganizationSlug string `toml:"organization_slug,omitempty"`
}

// getTaskOrganizationIdentifier determines organization slug from CLI flags or manifest
func getTaskOrganizationIdentifier(manifest *TaskManifestCollection, cliOrgSlug string) (string, error) {
	// CLI flags take precedence
	if cliOrgSlug != "" {
		return cliOrgSlug, nil
	}

	// Fall back to manifest deployment section
	if manifest.Deployment != nil && manifest.Deployment.OrganizationSlug != "" {
		return manifest.Deployment.OrganizationSlug, nil
	}

	// No organization specified
	return "", fmt.Errorf("missing organization identifier")
}

// getTaskEnvironmentType determines whether this is a Python or Docker task collection
func getTaskEnvironmentType(manifest *TaskManifestCollection) string {
	if manifest.Docker != nil {
		return "Docker"
	}
	if manifest.Python != nil {
		return "Python"
	}
	return "Unknown"
}

// deployTaskCollection handles deployment of task collections
func deployTaskCollection(ctx context.Context, input deployInput) error {
	if input.dryRun {
		return dryRunTaskCollection(input)
	}

	// Load and parse numerous-task.toml
	task := output.StartTask("Loading task collection configuration")
	taskManifestPath := filepath.Join(input.appDir, "numerous-task.toml")

	var taskManifest TaskManifestCollection
	if _, err := toml.DecodeFile(taskManifestPath, &taskManifest); err != nil {
		task.Error()
		output.PrintErrorDetails("Error loading task manifest", err)
		return err
	}

	// Environment detection for proper handling
	envType := "Python" // default
	if taskManifest.Docker != nil {
		envType = "Docker"
	}

	// Determine organization for deployment
	orgSlug, err := getTaskOrganizationIdentifier(&taskManifest, input.orgSlug)
	if err != nil {
		task.Error()
		output.PrintError("Organization identifier required for task collection deployment.", "Provide organization using --organization flag or in the task manifest [deployment] section.")
		return fmt.Errorf("missing organization identifier for task deployment")
	}

	task.Done()

	// Check deployment method configuration
	authToken := os.Getenv("NUMEROUS_TOKEN")
	if authToken == "" {
		authToken = os.Getenv("NUMEROUS_API_ACCESS_TOKEN")
	}

	// Determine API endpoint
	apiURL := os.Getenv("NUMEROUS_GRAPHQL_HTTP_URL")
	if apiURL == "" {
		apiURL = os.Getenv("NUMEROUS_API_URL") // Fallback for backward compatibility
	}
	if apiURL == "" {
		apiURL = "https://app.numerous.com/api/graphql"
	}

	graphqlEndpoint := apiURL
	// Only append /graphql if we're using the fallback URLs that don't already include the correct path
	if apiURL != os.Getenv("NUMEROUS_GRAPHQL_HTTP_URL") && !strings.HasSuffix(graphqlEndpoint, "/graphql") && !strings.HasSuffix(graphqlEndpoint, "/query") {
		if strings.HasSuffix(graphqlEndpoint, "/") {
			graphqlEndpoint += "graphql"
		} else {
			graphqlEndpoint += "/graphql"
		}
	}

	if shouldUseGraphQL(graphqlEndpoint) {
		if err := deployViaGraphQL(ctx, &taskManifest, orgSlug, graphqlEndpoint, authToken, input.appDir, input.verbose); err != nil {
			task.Error()
			output.PrintErrorDetails("GraphQL deployment failed", err)
			return err
		} else {
			task.Done()
			fmt.Printf("Task collection '%s' (v%s) deployed successfully via GraphQL!\n", taskManifest.Name, taskManifest.Version)
			fmt.Printf("Organization: %s\n", orgSlug)
			fmt.Printf("Environment: %s\n", envType)
			return nil
		}
	}

	// If GraphQL deployment is disabled, return error instead of falling back to mock
	task.Error()
	err = errors.New("GraphQL deployment is disabled. Set USE_MOCK_DEPLOYMENT=false to enable GraphQL deployment")
	output.PrintErrorDetails("Deployment configuration error", err)
	return err
}

// shouldUseGraphQL checks if we should attempt GraphQL deployment
func shouldUseGraphQL(endpoint string) bool {
	// Use GraphQL by default since it's integrated into main API
	return os.Getenv("USE_MOCK_DEPLOYMENT") != "true"
}

// deployViaGraphQL deploys the task collection using the GraphQL API
func deployViaGraphQL(ctx context.Context, manifest *TaskManifestCollection, orgSlug, endpoint, token string, sourceDir string, verbose bool) error {
	client := NewGraphQLClient(endpoint, token)

	// Convert manifest to GraphQL input
	input := convertTaskManifestToGraphQLInput(manifest, orgSlug)

	// Deploy via GraphQL using the new multi-step process
	response, err := client.DeployTaskCollectionGraphQL(ctx, input, sourceDir, verbose)
	if err != nil {
		return fmt.Errorf("GraphQL deployment failed: %w", err)
	}

	if !response.DeployTaskCollection.Success {
		return fmt.Errorf("deployment failed: %s", *response.DeployTaskCollection.Error)
	}

	return nil
}

// createSourceArchive creates a tar.gz archive of the source directory
func createSourceArchive(sourceDir, archivePath string) error {
	// For now, just copy the source directory structure (simple implementation)
	// In a production system, this would create a proper tar.gz archive
	return os.MkdirAll(filepath.Dir(archivePath), 0755)
}

// deployTaskCollectionMock provides the existing mock deployment functionality
func deployTaskCollectionMock(ctx context.Context, input deployInput, task *output.Task, taskManifest *TaskManifestCollection, envType, organizationSlug string) error {
	// Existing mock deployment logic
	mockBackendDir := filepath.Join(os.Getenv("HOME"), ".numerous", "mock_backend", "organizations", organizationSlug, "tasks", taskManifest.Name, taskManifest.Version)

	if err := os.MkdirAll(mockBackendDir, 0755); err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating mock backend directory", err)
		return err
	}

	manifestDestPath := filepath.Join(mockBackendDir, "numerous-task.toml")
	manifestSrcPath := filepath.Join(input.appDir, "numerous-task.toml")

	manifestContent, err := os.ReadFile(manifestSrcPath)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error reading manifest for storage", err)
		return err
	}

	if err := os.WriteFile(manifestDestPath, manifestContent, 0644); err != nil {
		task.Error()
		output.PrintErrorDetails("Error storing manifest in mock backend", err)
		return err
	}

	// Archive source files (existing logic)
	archivePath := filepath.Join(mockBackendDir, "source.tar.gz")
	if err := createSourceArchive(input.appDir, archivePath); err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating source archive", err)
		return err
	}

	task.Done()

	// Success output
	fmt.Printf("Task collection '%s' (v%s) deployed successfully!\n", taskManifest.Name, taskManifest.Version)
	fmt.Printf("Organization: %s\n", organizationSlug)
	fmt.Printf("Environment: %s\n", envType)
	fmt.Printf("Tasks: %d\n", len(taskManifest.Task))

	if envType == "Python" && taskManifest.Python != nil {
		fmt.Printf("Python Version: %s\n", taskManifest.Python.Version)
	} else if envType == "Docker" && taskManifest.Docker != nil {
		fmt.Printf("Dockerfile: %s\n", taskManifest.Docker.Dockerfile)
	}

	fmt.Printf("\nMock deployment stored at: %s\n", mockBackendDir)
	fmt.Printf("\nTo enable real deployment, set:\n")
	fmt.Printf("  export NUMEROUS_API_URL=https://api.numerous.com/graphql\n")
	fmt.Printf("  export NUMEROUS_ACCESS_TOKEN=your-token\n")
	fmt.Printf("  export USE_GRAPHQL_DEPLOYMENT=true\n")

	return nil
}

// dryRunTaskCollection shows what would be deployed for a task collection without actually deploying
func dryRunTaskCollection(input deployInput) error {
	task := output.StartTask("DRY RUN: Analyzing task collection configuration")

	// Load and validate task manifest (this is safe to do in dry-run)
	taskManifestPath := filepath.Join(input.appDir, "numerous-task.toml")

	var taskManifest TaskManifestCollection
	tomlData, err := os.ReadFile(taskManifestPath)
	if err != nil {
		task.Error()
		output.PrintErrorDetails(fmt.Sprintf("Error reading task manifest %s", taskManifestPath), err)
		return err
	}

	if _, err := toml.Decode(string(tomlData), &taskManifest); err != nil {
		task.Error()
		output.PrintErrorDetails(fmt.Sprintf("Error parsing task manifest %s", taskManifestPath), err)
		return err
	}

	if taskManifest.Name == "" || taskManifest.Version == "" {
		task.Error()
		output.PrintError("Task manifest %s is missing required 'name' or 'version' fields.", "", taskManifestPath)
		return fmt.Errorf("task manifest missing name or version")
	}

	// Validate environment configuration
	envType := getTaskEnvironmentType(&taskManifest)
	if envType == "Unknown" {
		task.Error()
		output.PrintError("Task manifest must specify either [python] or [docker] environment section.", "")
		return fmt.Errorf("missing environment configuration")
	}

	// Determine organization (CLI flags take precedence over manifest)
	orgSlug, err := getTaskOrganizationIdentifier(&taskManifest, input.orgSlug)
	if err != nil {
		task.Error()
		output.PrintError("Organization identifier required for task collection deployment.", "Provide organization using --organization flag or in the task manifest [deployment] section.")
		return fmt.Errorf("missing organization identifier for task deployment")
	}

	task.Done()

	// Show what would be deployed
	fmt.Println()
	output.PrintlnOK("=== DRY RUN: Task Collection Deployment Summary ===")
	fmt.Printf("Collection Name: %s\n", output.Highlight(taskManifest.Name))
	fmt.Printf("Version: %s\n", output.Highlight(taskManifest.Version))
	fmt.Printf("Organization: %s\n", output.Highlight(orgSlug))
	fmt.Printf("Environment: %s\n", output.Highlight(envType))
	if taskManifest.Description != "" {
		fmt.Printf("Description: %s\n", taskManifest.Description)
	}
	fmt.Printf("Directory: %s\n", input.appDir)
	fmt.Printf("Manifest: %s\n", taskManifestPath)

	// Show environment configuration details
	if taskManifest.Python != nil {
		fmt.Printf("\nPython Configuration:\n")
		fmt.Printf("  Version: %s\n", taskManifest.Python.Version)
		if taskManifest.Python.RequirementsFile != "" {
			fmt.Printf("  Requirements File: %s\n", taskManifest.Python.RequirementsFile)
			// Check if requirements file exists
			reqFilePath := filepath.Join(input.appDir, taskManifest.Python.RequirementsFile)
			if _, err := os.Stat(reqFilePath); err != nil {
				fmt.Printf("  ⚠️  Warning: Requirements file not found: %s\n", taskManifest.Python.RequirementsFile)
			} else {
				fmt.Printf("  ✅ Requirements file found\n")
			}
		}
	}

	if taskManifest.Docker != nil {
		fmt.Printf("\nDocker Configuration:\n")
		fmt.Printf("  Dockerfile: %s\n", taskManifest.Docker.Dockerfile)
		if taskManifest.Docker.Context != "" {
			fmt.Printf("  Build Context: %s\n", taskManifest.Docker.Context)
		}
		// Check if Dockerfile exists
		dockerfilePath := filepath.Join(input.appDir, taskManifest.Docker.Dockerfile)
		if _, err := os.Stat(dockerfilePath); err != nil {
			fmt.Printf("  ⚠️  Warning: Dockerfile not found: %s\n", taskManifest.Docker.Dockerfile)
		} else {
			fmt.Printf("  ✅ Dockerfile found\n")
		}
	}

	// Show tasks found in manifest
	if len(taskManifest.Task) > 0 {
		fmt.Printf("\nTasks defined in collection (%d):\n", len(taskManifest.Task))
		for i, taskDef := range taskManifest.Task {
			fmt.Printf("  %d. %s\n", i+1, output.Highlight(taskDef.FunctionName))

			// Show task-specific configuration based on environment type
			if envType == "Python" && taskDef.SourceFile != "" {
				fmt.Printf("     Source: %s\n", taskDef.SourceFile)
				if taskDef.DecoratedFunction != "" && taskDef.DecoratedFunction != taskDef.FunctionName {
					fmt.Printf("     Function: %s\n", taskDef.DecoratedFunction)
				}
				// Check if source file exists
				sourceFilePath := filepath.Join(input.appDir, taskDef.SourceFile)
				if _, err := os.Stat(sourceFilePath); err != nil {
					fmt.Printf("     ⚠️  Warning: Source file not found: %s\n", taskDef.SourceFile)
				} else {
					fmt.Printf("     ✅ Source file found\n")
				}
			}

			if envType == "Docker" && len(taskDef.Entrypoint) > 0 {
				fmt.Printf("     Entrypoint: %v\n", taskDef.Entrypoint)
			}

			if taskDef.APIEndpoint != "" {
				fmt.Printf("     API Endpoint: %s\n", taskDef.APIEndpoint)
			}

			if taskDef.PythonStub != "" {
				fmt.Printf("     Python Stub: %s\n", taskDef.PythonStub)
				// Check if stub exists
				stubPath := filepath.Join(input.appDir, taskDef.PythonStub)
				if _, err := os.Stat(stubPath); err != nil {
					fmt.Printf("     ⚠️  Warning: Python stub not found: %s\n", taskDef.PythonStub)
				} else {
					fmt.Printf("     ✅ Python stub found\n")
				}
			}

			if taskDef.Description != "" {
				fmt.Printf("     Description: %s\n", taskDef.Description)
			}
		}
	} else {
		fmt.Printf("\n⚠️  Warning: No tasks defined in manifest\n")
	}

	// Show organization-scoped mock deployment target
	homeDir, _ := os.UserHomeDir()
	mockBackendPath := filepath.Join(homeDir, ".numerous", "mock_backend", "organizations", orgSlug, "tasks", taskManifest.Name, taskManifest.Version)
	fmt.Printf("\nWould deploy to organization-scoped mock backend: %s\n", mockBackendPath)

	fmt.Println()
	output.PrintlnOK(fmt.Sprintf("DRY RUN: Would deploy %s task collection to organization '%s'", envType, orgSlug))
	return nil
}
