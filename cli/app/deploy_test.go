package app

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
	dir := t.TempDir()
	copyTo(t, "../testdata/streamlit_app", dir)

	apps := &mockAppService{}
	apps.On("Create", mock.Anything, mock.Anything).Return(app.CreateAppOutput{AppID: "app-id"}, nil)
	apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: "app-version-id"}, nil)
	apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: "http://upload/url"}, nil)
	apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)

	err := Deploy(context.TODO(), dir, "organization-slug", "app-name", apps)

	assert.NoError(t, err)
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
