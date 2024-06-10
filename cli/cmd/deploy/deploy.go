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

	"numerous/cli/cmd/initialize"
	"numerous/cli/cmd/output"
	"numerous/cli/cmd/validate"
	"numerous/cli/dotenv"
	"numerous/cli/internal/app"
	"numerous/cli/internal/archive"
	"numerous/cli/manifest"
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

var (
	ErrInvalidSlug    = errors.New("invalid organization slug")
	ErrInvalidAppName = errors.New("invalid app name")
)

func Deploy(ctx context.Context, apps AppService, appDir, projectDir, slug string, appName string, verbose bool) error {
	if !validate.IsValidIdentifier(slug) {
		output.PrintError("Error: Invalid organization %q.", "Must contain only lower-case alphanumerical characters and dashes.", slug)
		return ErrInvalidSlug
	}

	if !validate.IsValidIdentifier(appName) {
		output.PrintError("Error: Invalid app name %q.", "Must contain only lower-case alphanumerical characters and dashes.", appName)
		return ErrInvalidAppName
	}

	task := output.StartTask("Loading app configuration")
	manifest, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
	if err != nil {
		task.Error()
		output.PrintErrorAppNotInitialized(appDir)

		return err
	}

	secrets := loadSecretsFromEnv(appDir)
	task.Done()

	task = output.StartTask("Registering new version")
	appID, err := readOrCreateApp(ctx, apps, slug, appName, manifest, task)
	if err != nil {
		return err
	}

	appVersionInput := app.CreateAppVersionInput{AppID: appID}
	appVersionOutput, err := apps.CreateVersion(ctx, appVersionInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error creating app version remotely", err)

		return err
	}
	task.Done()

	task = output.StartTask("Creating app archive")
	tarSrcDir := appDir
	if projectDir != "" {
		tarSrcDir = projectDir
	}
	tarPath := path.Join(tarSrcDir, ".tmp_app_archive.tar")
	err = archive.TarCreate(tarSrcDir, tarPath, manifest.Exclude)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error archiving app source", err)

		return err
	}
	defer os.Remove(tarPath)

	archive, err := os.Open(tarPath)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error archiving app source", err)

		return err
	}
	task.Done()

	task = output.StartTask("Uploading app archive")
	uploadURLInput := app.AppVersionUploadURLInput(appVersionOutput)
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

	task = output.StartTask("Deploying app")
	deployAppInput := app.DeployAppInput{AppVersionID: appVersionOutput.AppVersionID, Secrets: secrets}
	deployAppOutput, err := apps.DeployApp(ctx, deployAppInput)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error deploying app", err)
	}

	input := app.DeployEventsInput{
		DeploymentVersionID: deployAppOutput.DeploymentVersionID,
		Handler: func(de app.DeployEvent) error {
			switch de.Typename {
			case "AppBuildMessageEvent":
				for _, l := range strings.Split(de.BuildMessage.Message, "\n") {
					task.AddLine("Build", l)
				}
			case "AppBuildErrorEvent":
				for _, l := range strings.Split(de.BuildError.Message, "\n") {
					task.AddLine("Error", l)
				}

				return fmt.Errorf("build error: %s", de.BuildError.Message)
			case "AppDeployStatusEvent":
				task.AddLine("Deploy", "Status: "+de.DeploymentStatus.Status)
				switch de.DeploymentStatus.Status {
				case "PENDING", "RUNNING":
				default:
					return fmt.Errorf("got status %s while deploying", de.DeploymentStatus.Status)
				}
			}

			return nil
		},
	}
	err = apps.DeployEvents(ctx, input)
	if err != nil {
		task.Error()
		output.PrintErrorDetails("Error occurred during deploy", err)
	} else {
		task.Done()
	}

	return nil
}

func readOrCreateApp(ctx context.Context, apps AppService, slug string, appName string, manifest *manifest.Manifest, task *output.Task) (string, error) {
	appReadInput := app.ReadAppInput{
		OrganizationSlug: slug,
		Name:             appName,
	}
	appReadOutput, err := apps.ReadApp(ctx, appReadInput)
	switch {
	case err == nil:
		return appReadOutput.AppID, nil
	case errors.Is(err, app.ErrAppNotFound):
		appCreateInput := app.CreateAppInput{
			OrganizationSlug: slug,
			Name:             appName,
			DisplayName:      manifest.Name,
			Description:      manifest.Description,
		}
		appCreateOutput, err := apps.Create(ctx, appCreateInput)
		if err != nil {
			task.Error()
			output.PrintErrorDetails("Error creating app remotely", err)

			return "", err
		}

		return appCreateOutput.AppID, nil
	default:
		output.PrintErrorDetails("Error reading remote app", err)
		task.Error()

		return "", err
	}
}

func loadSecretsFromEnv(appDir string) map[string]string {
	env, _ := dotenv.Load(path.Join(appDir, initialize.EnvFileName))
	return env
}
