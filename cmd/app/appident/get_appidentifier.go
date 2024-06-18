package appident

import (
	"path/filepath"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/cmd/validate"
	"numerous.com/cli/internal/manifest"
)

type AppIdentifier struct {
	OrganizationSlug string
	Name             string
}

// Uses the given slug and appName, or loads from manifest, and validates.
func GetAppIdentifier(appDir string, slug string, appName string) (AppIdentifier, error) {
	if slug == "" && appName == "" {
		manifest, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
		if err != nil {
			output.PrintErrorAppNotInitialized(appDir)

			return AppIdentifier{}, err
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

		return AppIdentifier{}, ErrInvalidSlug
	}

	if !validate.IsValidIdentifier(appName) {
		if appName == "" {
			output.PrintErrorMissingAppName()
		} else {
			output.PrintErrorInvalidAppName(appName)
		}

		return AppIdentifier{}, ErrInvalidAppName
	}

	return AppIdentifier{OrganizationSlug: slug, Name: appName}, nil
}
