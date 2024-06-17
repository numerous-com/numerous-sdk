package appident

import (
	"testing"

	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func TestGetAppIdentifier(t *testing.T) {
	const slug = "organization-slug"
	const appName = "app-name"

	t.Run("given valid slug and app name then they are returned", func(t *testing.T) {
		actualSlug, actualAppName, err := GetAppIdentifier("", slug, appName)

		assert.Equal(t, slug, actualSlug)
		assert.Equal(t, appName, actualAppName)
		assert.NoError(t, err)
	})

	t.Run("given invalid slug then invalid slug error is returned", func(t *testing.T) {
		actualSlug, actualAppName, err := GetAppIdentifier("", "Some Invalid Slug", appName)

		assert.Equal(t, "", actualSlug)
		assert.Equal(t, "", actualAppName)
		assert.ErrorIs(t, err, ErrInvalidSlug)
	})

	t.Run("given invalid app name then invalid app name error is returned", func(t *testing.T) {
		actualSlug, actualAppName, err := GetAppIdentifier("", slug, "Some Invalid App Name")

		assert.Equal(t, "", actualSlug)
		assert.Equal(t, "", actualAppName)
		assert.ErrorIs(t, err, ErrInvalidAppName)
	})

	t.Run("given app dir but no slug and no app name then it is loaded from manifest", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)

		actualSlug, actualAppName, err := GetAppIdentifier(appDir, "", "")

		assert.Equal(t, "organization-slug-in-manifest", actualSlug)
		assert.Equal(t, "app-name-in-manifest", actualAppName)
		assert.NoError(t, err)
	})
}
