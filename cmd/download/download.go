package download

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"numerous.com/cli/cmd/output"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/archive"
)

var ErrDownloadFailed = errors.New("download failed")

type Input struct {
	AppDir  string
	AppSlug string
	OrgSlug string
}

type AppService interface {
	CurrentAppVersion(context.Context, app.CurrentAppVersionInput) (app.CurrentAppVersionOutput, error)
	AppVersionDownloadURL(context.Context, app.AppVersionDownloadURLInput) (app.AppVersionDownloadURLOutput, error)
}

func Download(ctx context.Context, client *http.Client, service AppService, input Input, confirmOverwrite func(appDir string) bool) error {
	ai, err := appident.GetAppIdentifier(input.AppDir, nil, input.OrgSlug, input.AppSlug)
	if errors.Is(err, appident.ErrAppNotInitialized) {
		// ErrAppNotInitialized is only returned if both organization and app
		// slugs are missing. We just print there error for missing the app
		// slug here.
		output.PrintErrorMissingAppSlug()
		return err
	} else if err != nil {
		output.PrintGetAppIdentiferError(err, input.AppDir, ai)
		return err
	}

	if input.AppDir == "" {
		input.AppDir = input.AppSlug
	}

	t := output.StartTask(fmt.Sprintf("Locating app version for %s/%s", ai.OrganizationSlug, ai.AppSlug))
	appVersionInput := app.CurrentAppVersionInput(ai)
	appVersionOutput, err := service.CurrentAppVersion(ctx, appVersionInput)
	if err != nil {
		output.PrintAppError(err, ai)
		return err
	}

	urlInput := app.AppVersionDownloadURLInput(appVersionOutput)
	urlOutput, err := service.AppVersionDownloadURL(ctx, urlInput)
	if err != nil {
		t.Error()
		output.PrintErrorDetails("Error getting app download URL", err)

		return err
	}
	t.Done()

	if dirExists(input.AppDir) && !confirmOverwrite(input.AppDir) {
		output.PrintError("Download interrupted", "")
		return nil
	}

	t = output.StartTask(fmt.Sprintf("Downloading app source into %q", input.AppDir))
	if err := downloadArchive(client, input.AppDir, urlOutput.DownloadURL); err != nil {
		t.Error()
		output.PrintErrorDetails("Error downloading app source", err)

		return err
	}
	t.Done()

	return nil
}

func downloadArchive(client *http.Client, appDir string, url string) error {
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ErrDownloadFailed
	}

	return archive.TarExtract(resp.Body, appDir)
}

func dirExists(dir string) bool {
	_, err := os.Stat(dir)
	return !errors.Is(err, os.ErrNotExist)
}

func surveyConfirmOverwrite(dir string) bool {
	overwriteOK := false
	abspath, err := filepath.Abs(dir)
	if err != nil {
		return false
	}

	msg := fmt.Sprintf("The directory %q already exists, so downloading the app source may overwrite local files. Continue?", abspath)
	err = survey.AskOne(&survey.Confirm{Message: msg}, &overwriteOK)
	if err != nil {
		return false
	}

	return overwriteOK
}
