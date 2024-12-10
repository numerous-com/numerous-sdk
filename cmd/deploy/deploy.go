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

	"numerous.com/cli/cmd/logs"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/archive"
	"numerous.com/cli/internal/dotenv"
	"numerous.com/cli/internal/links"
	"numerous.com/cli/internal/manifest"
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
	AppDeployLogs(appident.AppIdentifier) (chan app.AppDeployLogEntry, error)
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
}

func deploy(ctx context.Context, apps appService, input deployInput) error {
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

	if err := deployApp(ctx, appVersionOutput, secrets, apps, input, appRelativePath); err != nil {
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

	appVersionInput := app.CreateAppVersionInput{AppID: appID, Version: input.version, Message: input.message}
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

func deployApp(ctx context.Context, appVersionOutput app.CreateAppVersionOutput, secrets map[string]string, apps appService, input deployInput, appRelativePath string) error {
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
	ch, err := apps.AppDeployLogs(ai)
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
