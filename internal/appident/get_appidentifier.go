package appident

import (
	"path/filepath"
	"regexp"
	"strings"

	"numerous.com/cli/cmd/validate"
	"numerous.com/cli/internal/manifest"
)

type AppIdentifier struct {
	OrganizationSlug string
	AppSlug          string
}

// Uses the given slug and appName, or loads from manifest, and validates.
func GetAppIdentifier(appDir string, m *manifest.Manifest, orgSlug string, appSlug string) (AppIdentifier, error) {
	if orgSlug == "" || appSlug == "" {
		if m == nil {
			loaded, err := manifest.Load(filepath.Join(appDir, manifest.ManifestFileName))
			if err != nil {
				return AppIdentifier{}, ErrAppNotInitialized
			}
			m = loaded
		}

		orgSlug = GetOrgSlug(m, orgSlug)
		appSlug = GetAppSlug(m, appSlug)
	}

	ai := AppIdentifier{OrganizationSlug: orgSlug, AppSlug: appSlug}

	if orgSlug == "" {
		return AppIdentifier{}, ErrMissingOrganizationSlug
	}

	if !validate.IsValidIdentifier(orgSlug) {
		return ai, ErrInvalidOrganizationSlug
	}

	if appSlug == "" {
		return AppIdentifier{}, ErrMissingAppSlug
	}

	if !validate.IsValidIdentifier(appSlug) {
		return ai, ErrInvalidAppSlug
	}

	return AppIdentifier{OrganizationSlug: orgSlug, AppSlug: appSlug}, nil
}

var (
	appNameWhitespaceRegexp *regexp.Regexp = regexp.MustCompile(`\s+`)
	appNameSanitizeRegexp   *regexp.Regexp = regexp.MustCompile(`[^0-9a-z-\s]`)
)

func GetAppSlug(m *manifest.Manifest, argAppSlug string) string {
	if argAppSlug != "" {
		return argAppSlug
	}

	if m.Deployment != nil && m.Deployment.AppSlug != "" {
		return m.Deployment.AppSlug
	}

	return manifestAppNameToAppSlug(m.Name)
}

func GetOrgSlug(m *manifest.Manifest, argOrgSlug string) string {
	if argOrgSlug != "" {
		return argOrgSlug
	}

	if m.Deployment != nil && m.Deployment.OrganizationSlug != "" {
		return m.Deployment.OrganizationSlug
	}

	// TODO: introduce error here
	return ""
}

// removes all characters except a-z, A-Z, 0-9, dashes and replaces all spaces
// with dashes
func manifestAppNameToAppSlug(name string) string {
	sanitized := strings.ToLower(name)
	sanitized = appNameSanitizeRegexp.ReplaceAllString(sanitized, "")
	sanitized = appNameWhitespaceRegexp.ReplaceAllString(sanitized, "-")

	return sanitized
}
