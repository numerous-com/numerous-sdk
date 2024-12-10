package logs

import (
	"context"
	"errors"
	"testing"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
)

func dummyPrinter(entry app.AppDeployLogEntry) {}

func TestLogs(t *testing.T) {
	const slug = "organization-slug"
	const appSlug = "app-slug"
	ai := appident.AppIdentifier{OrganizationSlug: slug, AppSlug: appSlug}
	testError := errors.New("test error")

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := Logs(context.TODO(), nil, appDir, "Some Invalid Organization Slug", appSlug, dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrInvalidOrganizationSlug)
	})

	t.Run("given invalid app slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := Logs(context.TODO(), nil, appDir, slug, "Some Invalid App Name", dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrInvalidAppSlug)
	})

	t.Run("given neither slug nor app slug arguments, numerous.toml without deploy and no config then it returns error", func(t *testing.T) {
		oldConfigBaseDir := config.OverrideConfigBaseDir(t.TempDir())
		t.Cleanup(func() {
			config.OverrideConfigBaseDir(oldConfigBaseDir)
		})

		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)

		err := Logs(context.TODO(), nil, appDir, "", "", dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrMissingOrganizationSlug)
	})

	t.Run("given no slug and app slug arguments and app dir without numerous.toml then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := Logs(context.TODO(), nil, appDir, "", "", dummyPrinter)

		assert.ErrorIs(t, err, appident.ErrAppNotInitialized)
	})

	t.Run("given slug and app slug arguments but not app dir then it calls service as expected", func(t *testing.T) {
		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai).Return(closedCh, nil)

		err := Logs(context.TODO(), apps, "", slug, appSlug, dummyPrinter)

		assert.NoError(t, err)
	})

	t.Run("given numerous.toml with deploy section then it calls service as expected", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		ai := appident.AppIdentifier{OrganizationSlug: "organization-slug-in-manifest", AppSlug: "app-slug-in-manifest"}
		apps.On("AppDeployLogs", ai).Return(closedCh, nil)

		err := Logs(context.TODO(), apps, appDir, "", "", dummyPrinter)

		assert.NoError(t, err)
		apps.AssertExpectations(t)
	})

	t.Run("it stops when context is cancelled", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai).Return(ch, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 10)
			cancel()
		}()
		err := Logs(ctx, apps, "", slug, appSlug, dummyPrinter)

		assert.NoError(t, err)
	})

	t.Run("given app service returns error, it returns the error", func(t *testing.T) {
		var nilChan chan app.AppDeployLogEntry = nil
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai).Return(nilChan, testError)

		err := Logs(context.TODO(), apps, "", slug, appSlug, dummyPrinter)

		assert.ErrorIs(t, err, testError)
	})

	t.Run("prints expected entries", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai).Return(ch, nil)

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
		err := Logs(context.TODO(), apps, "", slug, appSlug, printer)

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
