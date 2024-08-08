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

	"numerous.com/cli/cmd/initialize"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/archive"
	"numerous.com/cli/internal/dotenv"
	"numerous.com/cli/internal/links"
	"numerous.com/cli/internal/manifest"
)

type AppService interface {
	ReadApp(ctx context.Context, input app.ReadAppInput) (app.ReadAppOutput, error)
	Create(ctx context.Context, input app.CreateAppInput) (app.CreateAppOutput, error)
	CreateVersion(ctx context.Context, input app.CreateAppVersionInput) (app.CreateAppVersionOutput, error)
	AppVersionUploadURL(ctx context.Context, input app.AppVersionUploadURLInput) (app.AppVersionUploadURLOutput, error)
	UploadAppSource(uploadURL string, archive io.Reader) error
	DeployApp(ctx context.Context, input app.DeployAppInput) (app.DeployAppOutput, error)
	DeployEvents(ctx context.Context, input app.DeployEventsInput) error
}

type DeployInput struct {
	AppDir     string
	ProjectDir string
	OrgSlug    string
	AppSlug    string
	Version    string
	Message    string
	Verbose    bool
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

	return nil
}

func loadAppConfiguration(input DeployInput) (*manifest.Manifest, map[string]string, error) {
	task := output.StartTask("Loading app configuration")
	m, err := manifest.LoadManifest(filepath.Join(input.AppDir, manifest.ManifestPath))
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
		output.PrintGetAppIdentiferError(err, input.AppDir, ai)
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
		output.PrintGetAppIdentiferError(err, input.AppDir, ai)
		return app.CreateAppVersionOutput{}, "", "", err
	}

	task := output.StartTask("Registering new version for " + ai.OrganizationSlug + "/" + ai.AppSlug)
	appID, err := readOrCreateApp(ctx, apps, ai, manifest)
	if err != nil {
		task.Error()
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
	switch {
	case err == nil:
		return appReadOutput.AppID, nil
	case errors.Is(err, app.ErrAccessDenied):
		output.PrintErrorAccessDenied(ai)
		return "", err
	case !errors.Is(err, app.ErrAppNotFound):
		output.PrintErrorDetails("Error reading remote app", err)
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

	err = apps.UploadAppSource(uploadURLOutput.UploadURL, archive)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error uploading app source archive", err)

		return err
	}
	task.Done()

	return nil
}

func deployApp(ctx context.Context, appVersionOutput app.CreateAppVersionOutput, secrets map[string]string, apps AppService, input DeployInput) error {
	task := output.StartTask("Deploying app")
	deployAppInput := app.DeployAppInput{AppVersionID: appVersionOutput.AppVersionID, Secrets: secrets}
	deployAppOutput, err := apps.DeployApp(ctx, deployAppInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error deploying app", err)
	}

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

				return fmt.Errorf("build error: %s", de.BuildError.Message)
			case "AppDeploymentStatusEvent":
				if input.Verbose {
					task.AddLine("Deploy", "Status is "+de.DeploymentStatus.Status)
				}
				switch de.DeploymentStatus.Status {
				case "PENDING", "RUNNING":
				default:
					return fmt.Errorf("got status %s while deploying", de.DeploymentStatus.Status)
				}
			}

			return nil
		},
	}

	err = apps.DeployEvents(ctx, eventsInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error occurred during deploy", err)

		return err
	}
	task.Done()

	return nil
}

func loadSecretsFromEnv(appDir string) map[string]string {
	env, _ := dotenv.Load(path.Join(appDir, initialize.EnvFileName))
	return env
}
