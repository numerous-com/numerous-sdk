package logs

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"numerous/cli/cmd/output"
	"numerous/cli/cmd/validate"
	"numerous/cli/manifest"
)

var (
	ErrInvalidSlug    = errors.New("invalid organization slug")
	ErrInvalidAppName = errors.New("invalid app name")
)

type AppDeployLogEntry struct {
	Timestamp time.Time
	Text      string
}

type AppService interface {
	AppDeployLogs(slug, appName string) (chan AppDeployLogEntry, error)
}

func Logs(ctx context.Context, apps AppService, appDir, slug, appName string, printer func(AppDeployLogEntry)) error {
	slug, appName, err := getAppIdentifier(appDir, slug, appName)
	if err != nil {
		return err
	}

	ch, err := apps.AppDeployLogs(slug, appName)
	if err != nil {
		return err
	}

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return nil
			}
			printer(entry)
		case <-ctx.Done():
			return nil
		}
	}
}

func getAppIdentifier(appDir string, slug string, appName string) (string, string, error) {
	// load manifest if either slug or appName is missing
	if slug == "" || appName == "" {
		manifest, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
		if err != nil {
			output.PrintErrorAppNotInitialized(appDir)

			return "", "", err
		}

		if slug == "" && manifest.Deployment != nil {
			slug = manifest.Deployment.OrganizationSlug
		}

		if appName == "" && manifest.Deployment != nil {
			appName = manifest.Deployment.AppName
		}
	}

	if !validate.IsValidIdentifier(slug) {
		if slug == "" {
			output.PrintErrorMissingOrganizationSlug()
		} else {
			output.PrintErrorInvalidOrganizationSlug(slug)
		}

		return "", "", ErrInvalidSlug
	}

	if !validate.IsValidIdentifier(appName) {
		if appName == "" {
			output.PrintErrorMissingAppName()
		} else {
			output.PrintErrorInvalidAppName(appName)
		}

		return "", "", ErrInvalidAppName
	}

	return slug, appName, nil
}

func TimestampPrinter(entry AppDeployLogEntry) {
	ts := output.AnsiFaint + entry.Timestamp.Format(time.RFC3339) + output.AnsiReset
	fmt.Println(ts + " " + entry.Text)
}

func TextPrinter(entry AppDeployLogEntry) {
	fmt.Println(entry.Text)
}
