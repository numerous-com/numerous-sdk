package appident

import (
	"path/filepath"

	"numerous/cli/cmd/output"
	"numerous/cli/cmd/validate"
	"numerous/cli/manifest"
)

// Uses the given slug and appName, or loads from manifest, and validates.
func GetAppIdentifier(appDir string, slug string, appName string) (string, string, error) {
	if slug == "" && appName == "" {
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
