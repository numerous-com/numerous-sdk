package appident

import (
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestGetAppIdentifier(t *testing.T) {
	const slug = "organization-slug"
	const appSlug = "app-slug"
	expected := AppIdentifier{OrganizationSlug: slug, AppSlug: appSlug}

	t.Run("given valid slug and app slug then they are returned", func(t *testing.T) {
		actual, err := GetAppIdentifier("", nil, slug, appSlug)

		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	t.Run("given invalid organization slug then invalid organization slug error is returned", func(t *testing.T) {
		appident, err := GetAppIdentifier("", nil, "Some Invalid Organization Slug", appSlug)

		assert.Empty(t, appident)
		assert.ErrorIs(t, err, ErrInvalidOrganizationSlug)
	})

	t.Run("given invalid app slug then invalid app slug error is returned", func(t *testing.T) {
		actual, err := GetAppIdentifier("", nil, slug, "Some Invalid App Slug")

		assert.Equal(t, AppIdentifier{}, actual)
		assert.ErrorIs(t, err, ErrInvalidAppSlug)
	})

	t.Run("given app dir but no slug and no app slug then it is loaded from manifest", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		actual, err := GetAppIdentifier(appDir, nil, "", "")

		expected := AppIdentifier{
			OrganizationSlug: "organization-slug-in-manifest",
			AppSlug:          "app-slug-in-manifest",
		}
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})
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
