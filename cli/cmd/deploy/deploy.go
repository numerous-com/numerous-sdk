package deploy

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"

	"numerous/cli/cmd/output"
	"numerous/cli/cmd/validate"
	"numerous/cli/internal/app"
	"numerous/cli/internal/archive"
	"numerous/cli/manifest"
)

type AppService interface {
	AppVersionUploadURL(ctx context.Context, input app.AppVersionUploadURLInput) (app.AppVersionUploadURLOutput, error)
	Create(ctx context.Context, input app.CreateAppInput) (app.CreateAppOutput, error)
	CreateVersion(ctx context.Context, input app.CreateAppVersionInput) (app.CreateAppVersionOutput, error)
	UploadAppSource(uploadURL string, archive io.Reader) error
	DeployApp(ctx context.Context, input app.DeployAppInput) (app.DeployAppOutput, error)
	DeployEvents(ctx context.Context, input app.DeployEventsInput) error
}

var (
	ErrInvalidSlug    = errors.New("invalid organization slug")
	ErrInvalidAppName = errors.New("invalid app name")
)

func Deploy(ctx context.Context, dir string, slug string, appName string, apps AppService) error {
	if !validate.IsValidIdentifier(slug) {
		output.PrintError("Error: Invalid organization %q.", "Must contain only lower-case alphanumerical characters and dashes.", slug)
		return ErrInvalidSlug
	}

	if !validate.IsValidIdentifier(appName) {
		output.PrintError("Error: Invalid app name %q.", "Must contain only lower-case alphanumerical characters and dashes.", appName)
		return ErrInvalidAppName
	}

	manifest, err := manifest.LoadManifest(filepath.Join(dir, manifest.ManifestPath))
	if err != nil {
		output.PrintErrorAppNotInitialized()
		return err
	}

	appInput := app.CreateAppInput{
		OrganizationSlug: slug,
		Name:             appName,
		DisplayName:      manifest.Name,
		Description:      manifest.Description,
	}
	appOutput, err := apps.Create(ctx, appInput)
	if err != nil {
		output.PrintErrorDetails("Error creating app remotely", err)
		return err
	}

	appVersionInput := app.CreateAppVersionInput(appOutput)
	appVersionOutput, err := apps.CreateVersion(ctx, appVersionInput)
	if err != nil {
		output.PrintErrorDetails("Error creating app version remotely", err)
		return err
	}

	uploadURLInput := app.AppVersionUploadURLInput(appVersionOutput)
	uploadURLOutput, err := apps.AppVersionUploadURL(ctx, uploadURLInput)
	if err != nil {
		output.PrintErrorDetails("Error creating app version remotely", err)
		return err
	}

	tarPath := path.Join(dir, ".tmp_app_archive.tar")
	err = archive.TarCreate(dir, tarPath, manifest.Exclude)
	if err != nil {
		output.PrintErrorDetails("Error archiving app source", err)
		return err
	}
	defer os.Remove(tarPath)

	archive, err := os.Open(tarPath)
	if err != nil {
		output.PrintErrorDetails("Error archiving app source", err)
		return err
	}

	err = apps.UploadAppSource(uploadURLOutput.UploadURL, archive)
	if err != nil {
		output.PrintErrorDetails("Error uploading app source archive", err)
		return err
	}

	deployAppInput := app.DeployAppInput(appVersionOutput)
	deployAppOutput, err := apps.DeployApp(ctx, deployAppInput)
	if err != nil {
		output.PrintErrorDetails("Error deploying app", err)
	}

	input := app.DeployEventsInput{
		DeploymentVersionID: deployAppOutput.DeploymentVersionID,
		Handler: func(de app.DeployEvent) bool {
			println("DEPLOY:", de.Message)
			return true
		},
	}
	err = apps.DeployEvents(ctx, input)
	if err != nil {
		output.PrintErrorDetails("Error receiving deploy logs", err)
	}

	return nil
}
