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

func confirmAlways(appDir string) bool {
	return true
}

func getConfirmer(result bool, called *bool) func(string) bool {
	return func(appDir string) bool {
		*called = true
		return result
	}
}

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

		err := Download(context.TODO(), client, apps, Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}, confirmAlways)

		assert.NoError(t, err)
		if assert.DirExists(t, appDir) {
			assert.FileExists(t, appDir+"/numerous.toml")
			assert.FileExists(t, appDir+"/app.py")
			assert.FileExists(t, appDir+"/app_cover.jpg")
			assert.FileExists(t, appDir+"/requirements.txt")
		}
		apps.AssertExpectations(t)
	})

	t.Run("it does not confirm overwriting when app dir does not exist", func(t *testing.T) {
		appDir := t.TempDir() + "/some-app-dir"
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		confirmCalled := false

		err := Download(context.TODO(), client, apps, Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}, getConfirmer(true, &confirmCalled))

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
			Library: manifest.LibraryStreamlit,
			Deployment: &manifest.Deployment{
				OrganizationSlug: orgSlug,
				AppSlug:          appSlug,
			},
		}
		data, err := m.ToToml()
		require.NoError(t, err)
		var modeReadable fs.FileMode = 0o644
		require.NoError(t, os.WriteFile(appDir+"/numerous.toml", []byte(data), modeReadable))

		err = Download(context.TODO(), client, apps, Input{AppDir: appDir}, confirmAlways)

		assert.NoError(t, err)
		if assert.DirExists(t, appDir) {
			assert.FileExists(t, appDir+"/numerous.toml")
			assert.FileExists(t, appDir+"/app.py")
			assert.FileExists(t, appDir+"/app_cover.jpg")
			assert.FileExists(t, appDir+"/requirements.txt")
		}
		apps.AssertExpectations(t)
	})

	t.Run("it overwrites files if confirm returns true", func(t *testing.T) {
		appDir := t.TempDir()
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		filePath := appDir + "/app.py"
		originalData := []byte("some content that will be overwritten")
		test.WriteFile(t, filePath, originalData)
		confirmCalled := false

		err := Download(context.TODO(), client, apps, Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}, getConfirmer(true, &confirmCalled))

		assert.NoError(t, err)
		overwrittenData, err := os.ReadFile(filePath)
		if assert.NoError(t, err) {
			assert.NotEqual(t, originalData, overwrittenData)
		}
		assert.FileExists(t, appDir+"/numerous.toml")
		assert.FileExists(t, appDir+"/app_cover.jpg")
		assert.FileExists(t, appDir+"/requirements.txt")
		assert.True(t, confirmCalled)
		apps.AssertExpectations(t)
	})

	t.Run("it does not extract or overwrite files if confirm returns false", func(t *testing.T) {
		appDir := t.TempDir()
		client, downloadURL := newTestHTTPClientWithDownload(t, "streamlit_app.tar")
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)
		filePath := appDir + "/app.py"
		originalData := []byte("some content that will not be overwritten")
		test.WriteFile(t, filePath, originalData)
		confirmCalled := false

		err := Download(context.TODO(), client, apps, Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}, getConfirmer(false, &confirmCalled))

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

	t.Run("it returns error if download http request is not ok", func(t *testing.T) {
		appDir := t.TempDir()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		}))
		downloadURL := server.URL + "/some-file"
		apps := &mockAppService{}
		apps.On("CurrentAppVersion", mock.Anything, app.CurrentAppVersionInput{OrganizationSlug: orgSlug, AppSlug: appSlug}).Return(app.CurrentAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionDownloadURL", mock.Anything, app.AppVersionDownloadURLInput{AppVersionID: appVersionID}).Return(app.AppVersionDownloadURLOutput{DownloadURL: downloadURL}, nil)

		err := Download(context.TODO(), server.Client(), apps, Input{AppDir: appDir, AppSlug: appSlug, OrgSlug: orgSlug}, confirmAlways)

		assert.ErrorIs(t, err, ErrDownloadFailed)
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
