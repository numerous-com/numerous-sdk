package appident

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetAppIdentifier(t *testing.T) {
	const orgSlug = "organization-slug"
	const appSlug = "app-slug"

	t.Run("given invalid organization slug then invalid organization slug error is returned", func(t *testing.T) {
		config.OverrideConfigBaseDir(t.TempDir())

		actual, err := GetAppIdentifier("", nil, "Some Invalid Organization Slug", appSlug)

		assert.Equal(t, AppIdentifier{OrganizationSlug: "Some Invalid Organization Slug", AppSlug: appSlug}, actual)
		assert.ErrorIs(t, err, ErrInvalidOrganizationSlug)
	})

	t.Run("given invalid app slug then invalid app slug error is returned", func(t *testing.T) {
		config.OverrideConfigBaseDir(t.TempDir())

		actual, err := GetAppIdentifier("", nil, orgSlug, "Some Invalid App Slug")

		assert.Equal(t, AppIdentifier{OrganizationSlug: orgSlug, AppSlug: "Some Invalid App Slug"}, actual)
		assert.ErrorIs(t, err, ErrInvalidAppSlug)
	})

	t.Run("given error loading manifest and no app slug argument then error is returned", func(t *testing.T) {
		config.OverrideConfigBaseDir(t.TempDir())
		appDir := t.TempDir()
		// ensure error loading manifest by creating a folder at the manifest
		// location
		require.NoError(t, os.MkdirAll(appDir, fs.ModeDir))

		actual, err := GetAppIdentifier(appDir, nil, "", "")

		assert.Empty(t, actual)
		assert.ErrorIs(t, err, ErrAppNotInitialized)
	})

	type testCase struct {
		name   string
		argOrg string
		argApp string
		// manifest that is passed directly to the function
		manifestPreloaded *manifest.Manifest
		// manifest that is stored in the app folder
		manifestSaved *manifest.Manifest
		configOrg     string
		expected      AppIdentifier
	}

	for _, tc := range []testCase{
		{
			name:          "given valid arguments they are returned",
			argOrg:        "arg-org-slug",
			argApp:        "arg-app-slug",
			manifestSaved: &manifest.Manifest{Deployment: &manifest.Deployment{OrganizationSlug: "manifest-org-slug", AppSlug: "manifest-app-slug"}},
			configOrg:     "config-org-slug",
			expected: AppIdentifier{
				OrganizationSlug: "arg-org-slug",
				AppSlug:          "arg-app-slug",
			},
		},
		{
			name:      "given no organization in args or manifest it falls back to organization from config",
			argApp:    "arg-app-slug",
			configOrg: "config-org-slug",
			expected: AppIdentifier{
				OrganizationSlug: "config-org-slug",
				AppSlug:          "arg-app-slug",
			},
		},
		{
			name:          "given no organization in args it falls back to organization from loaded manifest",
			argApp:        "arg-app-slug",
			manifestSaved: &manifest.Manifest{Deployment: &manifest.Deployment{OrganizationSlug: "manifest-org-slug", AppSlug: "manifest-app-slug"}},
			expected: AppIdentifier{
				OrganizationSlug: "manifest-org-slug",
				AppSlug:          "arg-app-slug",
			},
		},
		{
			name:              "given no organization in args it falls back to organization from preloaded manifest",
			argApp:            "arg-app-slug",
			manifestPreloaded: &manifest.Manifest{Deployment: &manifest.Deployment{OrganizationSlug: "manifest-org-slug", AppSlug: "manifest-app-slug"}},
			expected: AppIdentifier{
				OrganizationSlug: "manifest-org-slug",
				AppSlug:          "arg-app-slug",
			},
		},
		{
			name:          "given no app in args it falls back to app from loaded manifest",
			argOrg:        "arg-org-slug",
			manifestSaved: &manifest.Manifest{Deployment: &manifest.Deployment{OrganizationSlug: "manifest-org-slug", AppSlug: "manifest-app-slug"}},
			expected: AppIdentifier{
				OrganizationSlug: "arg-org-slug",
				AppSlug:          "manifest-app-slug",
			},
		},
		{
			name:              "given no organization in args it falls back to organization from preloaded manifest",
			argOrg:            "arg-org-slug",
			manifestPreloaded: &manifest.Manifest{Deployment: &manifest.Deployment{OrganizationSlug: "manifest-org-slug", AppSlug: "manifest-app-slug"}},
			expected: AppIdentifier{
				OrganizationSlug: "arg-org-slug",
				AppSlug:          "manifest-app-slug",
			},
		},
		{
			name:              "given no app slug in args or manifest it falls back to converted preloaded manifest app name",
			argOrg:            "arg-org-slug",
			manifestPreloaded: &manifest.Manifest{App: manifest.App{Name: "App Name"}},
			expected: AppIdentifier{
				OrganizationSlug: "arg-org-slug",
				AppSlug:          "app-name",
			},
		},
		{
			name:          "given no app slug in args or manifest it falls back to converted loaded manifest app name",
			argOrg:        "arg-org-slug",
			manifestSaved: &manifest.Manifest{App: manifest.App{Name: "App Name"}},
			expected: AppIdentifier{
				OrganizationSlug: "arg-org-slug",
				AppSlug:          "app-name",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			appDir := t.TempDir()

			// setup config
			configDir := t.TempDir()
			config.OverrideConfigBaseDir(configDir)
			if tc.configOrg != "" {
				cfg := config.Config{OrganizationSlug: tc.configOrg}
				require.NoError(t, cfg.Save())
			}

			// setup saved manifest
			if tc.manifestSaved != nil {
				mToml, err := tc.manifestSaved.ToTOML()
				require.NoError(t, err)
				test.WriteFile(t, filepath.Join(appDir, manifest.ManifestFileName), []byte(mToml))
			}

			actual, err := GetAppIdentifier(appDir, tc.manifestPreloaded, tc.argOrg, tc.argApp)

			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestManifestAppNameToAppSlug(t *testing.T) {
	for _, tc := range []struct {
		ManifestAppName string
		ExpectedAppSlug string
	}{
		{ManifestAppName: "LOWERCASE", ExpectedAppSlug: "lowercase"},
		{ManifestAppName: "Replace Spaces With Dashes", ExpectedAppSlug: "replace-spaces-with-dashes"},
		{ManifestAppName: "Collapse  Spaces  In  App  Name", ExpectedAppSlug: "collapse-spaces-in-app-name"},
		{ManifestAppName: "Strip Special Characters Like !\"#¤'_,* From App Name", ExpectedAppSlug: "strip-special-characters-like-from-app-name"},
	} {
		testName := tc.ManifestAppName + " sanitizes to " + tc.ExpectedAppSlug
		t.Run(testName, func(t *testing.T) {
			actual := manifestAppNameToAppSlug(tc.ManifestAppName)
			assert.Equal(t, tc.ExpectedAppSlug, actual)
		})
	}
}
