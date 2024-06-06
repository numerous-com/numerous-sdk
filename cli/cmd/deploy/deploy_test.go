package deploy

import (
	"context"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"testing"

	"numerous/cli/internal/app"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDeploy(t *testing.T) {
	const slug = "organization-slug"
	const appID = "app-id"
	const appName = "app-name"
	const appVersionID = "app-version-id"
	const uploadURL = "https://upload/url"
	const deployVersionID = "deploy-version-id"

	t.Run("give no existing app then happy path can run", func(t *testing.T) {
		dir := t.TempDir()
		copyTo(t, "../../testdata/streamlit_app", dir)

		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{}, app.ErrAppNotFound)
		apps.On("Create", mock.Anything, mock.Anything).Return(app.CreateAppOutput{AppID: appID}, nil)
		apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: uploadURL}, nil)
		apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)
		apps.On("DeployApp", mock.Anything, mock.Anything).Return(app.DeployAppOutput{DeploymentVersionID: deployVersionID}, nil)
		apps.On("DeployEvents", mock.Anything, mock.Anything).Return(nil)

		err := Deploy(context.TODO(), dir, slug, false, appName, apps)

		assert.NoError(t, err)
	})

	t.Run("give existing app then it does not create app", func(t *testing.T) {
		dir := t.TempDir()
		copyTo(t, "../../testdata/streamlit_app", dir)

		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{AppID: appID}, nil)
		apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: uploadURL}, nil)
		apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)
		apps.On("DeployApp", mock.Anything, mock.Anything).Return(app.DeployAppOutput{DeploymentVersionID: deployVersionID}, nil)
		apps.On("DeployEvents", mock.Anything, mock.Anything).Return(nil)

		err := Deploy(context.TODO(), dir, slug, false, appName, apps)

		assert.NoError(t, err)
	})

	t.Run("given dir without numerous.toml then it returns error", func(t *testing.T) {
		dir := t.TempDir()

		err := Deploy(context.TODO(), dir, slug, false, appName, nil)

		assert.EqualError(t, err, "open "+dir+"/numerous.toml: no such file or directory")
	})

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		err := Deploy(context.TODO(), ".", "Some Invalid Organization Slug", false, appName, nil)

		assert.ErrorIs(t, err, ErrInvalidSlug)
	})

	t.Run("given invalid app name then it returns error", func(t *testing.T) {
		err := Deploy(context.TODO(), ".", slug, false, "Some Invalid App Name", nil)

		assert.ErrorIs(t, err, ErrInvalidAppName)
	})
}

func copyTo(t *testing.T, src string, dest string) {
	t.Helper()

	err := filepath.Walk(src, func(p string, info fs.FileInfo, err error) error {
		require.NoError(t, err)
		if p == src {
			return nil
		}

		rel, err := filepath.Rel(src, p)
		require.NoError(t, err)
		destPath := path.Join(dest, rel)

		if info.IsDir() {
			err := os.Mkdir(p, os.ModePerm)
			require.NoError(t, err)

			return nil
		}

		file, err := os.Open(p)
		require.NoError(t, err)

		data, err := io.ReadAll(file)
		require.NoError(t, err)

		err = os.WriteFile(destPath, data, fs.ModePerm)
		require.NoError(t, err)

		return nil
	})

	require.NoError(t, err)
}
