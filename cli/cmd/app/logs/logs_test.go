package logs

import (
	"context"
	"errors"
	"testing"
	"time"

	"numerous/cli/cmd/app/appident"
	"numerous/cli/internal/app"
	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
)

func dummyPrinter(entry app.AppDeployLogEntry) {}

func TestLogs(t *testing.T) {
	const slug = "organization-slug"
	const appName = "app-name"
	testError := errors.New("test error")

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		err := Logs(context.TODO(), nil, appDir, "Some Invalid Organization Slug", appName, dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrInvalidSlug)
	})

	t.Run("given invalid app name then it returns error", func(t *testing.T) {
		err := Logs(context.TODO(), nil, appDir, slug, "Some Invalid App Name", dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrInvalidAppName)
	})

	t.Run("given neither slug nor app name arguments and numerous.toml without deploy then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app_without_deploy", appDir)

		err := Logs(context.TODO(), nil, appDir, "", "", dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrInvalidSlug)
	})

	t.Run("given no slug and app name arguments and app dir without numerous.toml then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := Logs(context.TODO(), nil, appDir, "", "", dummyPrinter)

		assert.ErrorContains(t, err, "no such file or directory")
	})

	t.Run("given slug and app name arguments but not app dir then it calls service as expected", func(t *testing.T) {
		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", slug, appName).Return(closedCh, nil)

		err := Logs(context.TODO(), apps, "", slug, appName, dummyPrinter)

		assert.NoError(t, err)
	})

	t.Run("given numerous.toml with deploy section then it calls service as expected", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)

		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", "organization-slug-in-manifest", "app-name-in-manifest").Return(closedCh, nil)

		err := Logs(context.TODO(), apps, appDir, "", "", dummyPrinter)

		assert.NoError(t, err)
		apps.AssertExpectations(t)
	})

	t.Run("it stops when context is cancelled", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", slug, appName).Return(ch, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 10)
			cancel()
		}()
		err := Logs(ctx, apps, "", slug, appName, dummyPrinter)

		assert.NoError(t, err)
	})

	t.Run("given app service returns error, it returns the error", func(t *testing.T) {
		var nilChan chan app.AppDeployLogEntry = nil
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", slug, appName).Return(nilChan, testError)

		err := Logs(context.TODO(), apps, "", slug, appName, dummyPrinter)

		assert.ErrorIs(t, err, testError)
	})

	t.Run("prints expected entries", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", slug, appName).Return(ch, nil)

		entry1 := app.AppDeployLogEntry{Timestamp: time.Date(2024, time.March, 1, 1, 1, 1, 1, time.UTC)}
		entry2 := app.AppDeployLogEntry{Timestamp: time.Date(2024, time.March, 1, 2, 2, 2, 2, time.UTC)}
		expected := []app.AppDeployLogEntry{entry1, entry2}
		actual := []app.AppDeployLogEntry{}
		printer := func(e app.AppDeployLogEntry) {
			actual = append(actual, e)
		}
		go func() {
			defer close(ch)
			ch <- entry1
			ch <- entry2
		}()
		err := Logs(context.TODO(), apps, "", slug, appName, printer)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
