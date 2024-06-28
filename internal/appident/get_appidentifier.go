package appident

import (
	"path/filepath"

	"numerous.com/cli/cmd/output"
	"numerous.com/cli/cmd/validate"
	"numerous.com/cli/internal/manifest"
)

type AppIdentifier struct {
	OrganizationSlug string
	AppSlug          string
}

// Uses the given slug and appName, or loads from manifest, and validates.
func GetAppIdentifier(appDir string, orgSlug string, appSlug string) (AppIdentifier, error) {
	if orgSlug == "" && appSlug == "" {
		manifest, err := manifest.LoadManifest(filepath.Join(appDir, manifest.ManifestPath))
		if err != nil {
			output.PrintErrorAppNotInitialized(appDir)

			return AppIdentifier{}, err
		}

		if orgSlug == "" && manifest.Deployment != nil {
			orgSlug = manifest.Deployment.OrganizationSlug
		}

		if appSlug == "" && manifest.Deployment != nil {
			appSlug = manifest.Deployment.AppSlug
		}
	}

	if !validate.IsValidIdentifier(orgSlug) {
		if orgSlug == "" {
			output.PrintErrorMissingOrganizationSlug()
		} else {
			output.PrintErrorInvalidOrganizationSlug(orgSlug)
		}

		return AppIdentifier{}, ErrInvalidOrganizationSlug
	}

	if !validate.IsValidIdentifier(appSlug) {
		if appSlug == "" {
			output.PrintErrorMissingAppSlug()
		} else {
			output.PrintErrorInvalidAppSlug(appSlug)
		}

		return AppIdentifier{}, ErrInvalidAppSlug
	}

	return AppIdentifier{OrganizationSlug: orgSlug, AppSlug: appSlug}, nil
}
