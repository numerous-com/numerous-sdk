package download

import (
	"context"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/manifest"
	"numerous.com/cli/internal/test"
)

func TestDownload(t *testing.T) {
	appSlug := "app-slug"
	orgSlug := "org-slug"
	appVersionID := "app-version-id"

	t.Run("downloads and extracts from expected download url", func(t *testing.T) {
		appDir := t.TempDir() + "/some-app-dir"
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)

		err := download(context.TODO(), client, apps, downloadInput{appDir: appDir, appSlug: appSlug, orgSlug: orgSlug, overwriteConfirmer: confirmAlways})

		assert.NoError(t, err)
		if assert.DirExists(t, appDir) {
			assert.FileExists(t, appDir+"/numerous.toml")
			assert.FileExists(t, appDir+"/app.py")
			assert.FileExists(t, appDir+"/app_cover.jpg")
			assert.FileExists(t, appDir+"/requirements.txt")
		}
		apps.AssertExpectations(t)
	})

	t.Run("does not confirm overwriting when app dir does not exist", func(t *testing.T) {
		appDir := t.TempDir() + "/some-app-dir"
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		confirmCalled := false

		err := download(context.TODO(), client, apps, downloadInput{appDir: appDir, appSlug: appSlug, orgSlug: orgSlug, overwriteConfirmer: getConfirmer(true, &confirmCalled)})

		assert.NoError(t, err)
		assert.False(t, confirmCalled)
	})

	t.Run("uses default deployment configuration in existing numerous.toml", func(t *testing.T) {
		appDir := t.TempDir()
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		m := manifest.Manifest{
			Python: &manifest.Python{
				Library: manifest.LibraryStreamlit,
			},
			Deployment: &manifest.Deployment{
				OrganizationSlug: orgSlug,
				AppSlug:          appSlug,
			},
		}
		data, err := m.ToTOML()
		require.NoError(t, err)
		var modeReadable fs.FileMode = 0o644
		require.NoError(t, os.WriteFile(appDir+"/numerous.toml", []byte(data), modeReadable))

		err = download(context.TODO(), client, apps, downloadInput{appDir: appDir, overwriteConfirmer: confirmAlways})

		assert.NoError(t, err)
		if assert.DirExists(t, appDir) {
			assert.FileExists(t, appDir+"/numerous.toml")
			assert.FileExists(t, appDir+"/app.py")
			assert.FileExists(t, appDir+"/app_cover.jpg")
			assert.FileExists(t, appDir+"/requirements.txt")
		}
		apps.AssertExpectations(t)
	})

	t.Run("overwrites files if confirm returns true", func(t *testing.T) {
		appDir := t.TempDir()
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		confirmCalled := false

		err := download(context.TODO(), client, apps, downloadInput{appDir: appDir, appSlug: appSlug, orgSlug: orgSlug, overwriteConfirmer: getConfirmer(true, &confirmCalled)})

		assert.NoError(t, err)
		assertFileContentEqual(t, "../../testdata/streamlit_app/app.py", appDir+"/app.py")
		if assert.FileExists(t, appDir+"/numerous.toml") {
			assertFileContentEqual(t, appDir+"/numerous.toml", "../../testdata/streamlit_app/numerous.toml")
		}
		if assert.FileExists(t, appDir+"/app_cover.jpg") {
			assertFileContentEqual(t, appDir+"/app_cover.jpg", "../../testdata/streamlit_app/app_cover.jpg")
		}
		if assert.FileExists(t, appDir+"/requirements.txt") {
			assertFileContentEqual(t, appDir+"/requirements.txt", "../../testdata/streamlit_app/requirements.txt")
		}
		assert.True(t, confirmCalled)
		apps.AssertExpectations(t)
	})

	t.Run("does not extract or overwrite files if confirm returns false", func(t *testing.T) {
		appDir := t.TempDir()
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		filePath := appDir + "/app.py"
		originalData := []byte("some content that will not be overwritten")
		test.WriteFile(t, filePath, originalData)
		confirmCalled := false

		err := download(context.TODO(), client, apps, downloadInput{appDir: appDir, appSlug: appSlug, orgSlug: orgSlug, overwriteConfirmer: getConfirmer(false, &confirmCalled)})

		assert.NoError(t, err)
		notOverwrittenData, err := os.ReadFile(filePath)
		if assert.NoError(t, err) {
			assert.Equal(t, originalData, notOverwrittenData)
		}
		assert.NoFileExists(t, appDir+"/numerous.toml")
		assert.NoFileExists(t, appDir+"/app_cover.jpg")
		assert.NoFileExists(t, appDir+"/requirements.txt")
		assert.True(t, confirmCalled)
		apps.AssertExpectations(t)
	})

	t.Run("returns error if download http request is not ok", func(t *testing.T) {
		appDir := t.TempDir()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		downloadURL := server.URL + "/some-file"
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)

		err := download(context.TODO(), server.Client(), apps, downloadInput{appDir: appDir, appSlug: appSlug, orgSlug: orgSlug, overwriteConfirmer: confirmAlways})

		assert.ErrorIs(t, err, errDownloadFailed)
	})
}

func newTestHTTPClientWithDownload(t *testing.T, testdataFilePath string) (client *http.Client, url string) {
	t.Helper()

	contentFilePath := filepath.Join("../../testdata/", testdataFilePath)

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/"+testdataFilePath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		data, _ := os.ReadFile(contentFilePath)

		w.Write(data) // nolint:errcheck
	}))

	return s.Client(), s.URL + "/" + testdataFilePath
}

func confirmAlways(appDir string) bool {
	return true
}

func getConfirmer(result bool, called *bool) func(string) bool {
	return func(appDir string) bool {
		*called = true
		return result
	}
}

func assertFileContentEqual(t *testing.T, expectedContentPath string, actualContentPath string) {
	t.Helper()

	expected, err := os.ReadFile(expectedContentPath)
	if !assert.NoError(t, err) {
		return
	}

	actual, err := os.ReadFile(actualContentPath)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, expected, actual)
}
