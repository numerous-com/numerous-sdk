package deploy

import (
	"context"
	"errors"
	"fmt"
	"io"
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

var ErrArchiveTooLarge = fmt.Errorf("archive exceeds maximum size %d bytes", maxUploadBytes)

type deployBuildError struct {
	Message string
}

func (e *deployBuildError) Error() string {
	return e.Message
}

type AppService interface {
	ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error)
	Create(ctx context.Context, input app.CreateAppInput) (app.CreateAppOutput, error)
	CreateVersion(ctx context.Context, input app.CreateAppVersionInput) (app.CreateAppVersionOutput, error)
	AppVersionUploadURL(ctx context.Context, input app.AppVersionUploadURLInput) (app.AppVersionUploadURLOutput, error)
	UploadAppSource(uploadURL string, archive io.Reader) error
	DeployApp(ctx context.Context, input app.DeployAppInput) (app.DeployAppOutput, error)
	DeployEvents(ctx context.Context, input app.DeployEventsInput) error
	AppDeployLogs(appident.AppIdentifier) (chan app.AppDeployLogEntry, error)
}

type DeployInput struct {
	AppDir     string
	ProjectDir string
	OrgSlug    string
	AppSlug    string
	Version    string
	Message    string
	Verbose    bool
	Follow     bool
}

func Deploy(ctx context.Context, apps AppService, input DeployInput) error {
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

	if err := deployApp(ctx, appVersionOutput, secrets, apps, input); err != nil {
		return err
	}

	output.PrintlnOK("Access your app at: " + links.GetAppURL(orgSlug, appSlug))

	if input.Follow {
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
		if input.ProjectDir != "" {
			projectDirArg = " --project-dir=" + input.ProjectDir
		}

		appDirArg := ""
		if input.AppDir != "" {
			appDirArg = " " + input.AppDir
		}

		fmt.Println("  " + output.Highlight("numerous deploy --follow --organization="+orgSlug+" --app="+appSlug+projectDirArg+appDirArg))
	}

	return nil
}

func loadAppConfiguration(input DeployInput) (*manifest.Manifest, map[string]string, error) {
	task := output.StartTask("Loading app configuration")
	m, err := manifest.Load(filepath.Join(input.AppDir, manifest.ManifestFileName))
	if err != nil {
		task.Error()
		output.PrintErrorAppNotInitialized(input.AppDir)
		output.PrintManifestTOMLError(err)

		return nil, nil, err
	}

	secrets := loadSecretsFromEnv(input.AppDir)

	// for validation
	ai, err := appident.GetAppIdentifier(input.AppDir, m, input.OrgSlug, input.AppSlug)
	if err != nil {
		task.Error()
		appident.PrintGetAppIdentiferError(err, input.AppDir, ai)

		return nil, nil, err
	}

	task.Done()

	return m, secrets, nil
}

func createAppArchive(input DeployInput, manifest *manifest.Manifest) (*os.File, error) {
	task := output.StartTask("Creating app archive")
	tarSrcDir := input.AppDir
	if input.ProjectDir != "" {
		tarSrcDir = input.ProjectDir
	}
	tarPath := path.Join(tarSrcDir, ".tmp_app_archive.tar")

	if err := archive.TarCreate(tarSrcDir, tarPath, manifest.Exclude); err != nil {
		task.Error()
		output.PrintErrorDetails("Error archiving app source", err)

		return nil, err
	}
	defer os.Remove(tarPath)

	archive, err := os.Open(tarPath)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error archiving app source", err)

		return nil, err
	}
	task.Done()

	return archive, nil
}

func registerAppVersion(ctx context.Context, apps AppService, input DeployInput, manifest *manifest.Manifest) (app.CreateAppVersionOutput, string, string, error) {
	ai, err := appident.GetAppIdentifier("", manifest, input.OrgSlug, input.AppSlug)
	if err != nil {
		appident.PrintGetAppIdentiferError(err, input.AppDir, ai)
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

	appVersionInput := app.CreateAppVersionInput{AppID: appID, Version: input.Version, Message: input.Message}
	appVersionOutput, err := apps.CreateVersion(ctx, appVersionInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app version remotely", err)

		return app.CreateAppVersionOutput{}, "", "", err
	}
	task.Done()

	return appVersionOutput, ai.OrganizationSlug, ai.AppSlug, nil
}

func readOrCreateApp(ctx context.Context, apps AppService, ai appident.AppIdentifier, manifest *manifest.Manifest) (string, error) {
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

func uploadAppArchive(ctx context.Context, apps AppService, archive *os.File, appVersionID string) error {
	task := output.StartTask("Uploading app archive")
	uploadURLInput := app.AppVersionUploadURLInput(app.AppVersionUploadURLInput{AppVersionID: appVersionID})
	uploadURLOutput, err := apps.AppVersionUploadURL(ctx, uploadURLInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app version remotely", err)

		return err
	}

	if stat, err := archive.Stat(); err != nil {
		task.Error()
		output.PrintErrorDetails("Error checking archive size", err)

		return err
	} else if stat.Size() > maxUploadBytes {
		task.Error()
		printAppSourceArchiveTooLarge(stat.Size())

		return ErrArchiveTooLarge
	}

	err = apps.UploadAppSource(uploadURLOutput.UploadURL, archive)
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

func deployApp(ctx context.Context, appVersionOutput app.CreateAppVersionOutput, secrets map[string]string, apps AppService, input DeployInput) error {
	task := output.StartTask("Deploying app")
	deployAppInput := app.DeployAppInput{AppVersionID: appVersionOutput.AppVersionID, Secrets: secrets}
	deployAppOutput, err := apps.DeployApp(ctx, deployAppInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error deploying app", err)
	}

	appDeploymentStatusEventUpdater := statusUpdater{verbose: input.Verbose, task: task}
	eventsInput := app.DeployEventsInput{
		DeploymentVersionID: deployAppOutput.DeploymentVersionID,
		Handler: func(de app.DeployEvent) error {
			switch de.Typename {
			case "AppBuildMessageEvent":
				if input.Verbose {
					for _, l := range strings.Split(de.BuildMessage.Message, "\n") {
						task.AddLine("Build", l)
					}
				}
			case "AppBuildErrorEvent":
				if input.Verbose {
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

func followLogs(ctx context.Context, apps AppService, orgSlug, appSlug string) error {
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
