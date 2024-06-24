package appident

import (
	"testing"

	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func TestGetAppIdentifier(t *testing.T) {
	const slug = "organization-slug"
	const appName = "app-name"
	expected := AppIdentifier{OrganizationSlug: slug, Name: appName}

	t.Run("given valid slug and app name then they are returned", func(t *testing.T) {
		actual, err := GetAppIdentifier("", slug, appName)

		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})

	t.Run("given invalid slug then invalid slug error is returned", func(t *testing.T) {
		appident, err := GetAppIdentifier("", "Some Invalid Slug", appName)

		assert.Empty(t, appident)
		assert.ErrorIs(t, err, ErrInvalidSlug)
	})

	t.Run("given invalid app name then invalid app name error is returned", func(t *testing.T) {
		actual, err := GetAppIdentifier("", slug, "Some Invalid App Name")

		assert.Equal(t, AppIdentifier{}, actual)
		assert.ErrorIs(t, err, ErrInvalidAppName)
	})

	t.Run("given app dir but no slug and no app name then it is loaded from manifest", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)

		actual, err := GetAppIdentifier(appDir, "", "")

		expected := AppIdentifier{
			OrganizationSlug: "organization-slug-in-manifest",
			Name:             "app-name-in-manifest",
		}
		assert.Equal(t, expected, actual)
		assert.NoError(t, err)
	})
}
