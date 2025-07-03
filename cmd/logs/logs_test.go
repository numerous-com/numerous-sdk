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
	const orgSlug = "organization-slug"
	const appSlug = "app-slug"
	ai := appident.AppIdentifier{OrganizationSlug: orgSlug, AppSlug: appSlug}
	testError := errors.New("test error")

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := logs(context.TODO(), nil, logsInput{appDir: appDir, orgSlug: "Some Invalid Organization Slug", appSlug: appSlug, printer: dummyPrinter})

		assert.ErrorIs(t, err, appident.ErrInvalidOrganizationSlug)
	})

	t.Run("given invalid app slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := logs(context.TODO(), nil, logsInput{appDir: appDir, orgSlug: orgSlug, appSlug: "Some Invalid App Name", printer: dummyPrinter})

		assert.ErrorIs(t, err, appident.ErrInvalidAppSlug)
	})

	t.Run("given neither slug nor app slug arguments, numerous.toml without deploy and no config then it returns error", func(t *testing.T) {
		oldConfigBaseDir := config.OverrideConfigBaseDir(t.TempDir())
		t.Cleanup(func() {
			config.OverrideConfigBaseDir(oldConfigBaseDir)
		})

		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)

		err := logs(context.TODO(), nil, logsInput{appDir: appDir, orgSlug: "", appSlug: "", printer: dummyPrinter})

		assert.ErrorIs(t, err, appident.ErrMissingOrganizationSlug)
	})

	t.Run("given no slug and app slug arguments and app dir without numerous.toml then it returns error", func(t *testing.T) {
		appDir := t.TempDir()

		err := logs(context.TODO(), nil, logsInput{appDir: appDir, orgSlug: "", appSlug: "", printer: dummyPrinter})

		assert.ErrorIs(t, err, appident.ErrAppNotInitialized)
	})

	t.Run("given slug and app slug arguments but not app dir then it calls service as expected", func(t *testing.T) {
		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai, (*int)(nil), true).Return(closedCh, nil)

		err := logs(context.TODO(), apps, logsInput{appDir: "", orgSlug: orgSlug, appSlug: appSlug, tail: 0, follow: true, printer: dummyPrinter})

		assert.NoError(t, err)
	})

	t.Run("given numerous.toml with deploy section then it calls service as expected", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		ai := appident.AppIdentifier{OrganizationSlug: "organization-slug-in-manifest", AppSlug: "app-slug-in-manifest"}
		apps.On("AppDeployLogs", ai, (*int)(nil), true).Return(closedCh, nil)

		err := logs(context.TODO(), apps, logsInput{appDir: appDir, orgSlug: "", appSlug: "", tail: 0, follow: true, printer: dummyPrinter})

		assert.NoError(t, err)
		apps.AssertExpectations(t)
	})

	t.Run("it stops when context is cancelled", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai, (*int)(nil), true).Return(ch, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			time.Sleep(time.Millisecond * 10)
			cancel()
		}()
		err := logs(ctx, apps, logsInput{appDir: "", orgSlug: orgSlug, appSlug: appSlug, tail: 0, follow: true, printer: dummyPrinter})

		assert.NoError(t, err)
	})

	t.Run("given app service returns error, it returns the error", func(t *testing.T) {
		var nilChan chan app.AppDeployLogEntry = nil
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai, (*int)(nil), true).Return(nilChan, testError)

		err := logs(context.TODO(), apps, logsInput{appDir: "", orgSlug: orgSlug, appSlug: appSlug, tail: 0, follow: true, printer: dummyPrinter})

		assert.ErrorIs(t, err, testError)
	})

	t.Run("prints expected entries", func(t *testing.T) {
		ch := make(chan app.AppDeployLogEntry)
		apps := &AppServiceMock{}
		apps.On("AppDeployLogs", ai, (*int)(nil), true).Return(ch, nil)

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
		err := logs(context.TODO(), apps, logsInput{appDir: "", orgSlug: orgSlug, appSlug: appSlug, tail: 0, follow: true, printer: printer})

		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("calls service with correct tail and follow parameters", func(t *testing.T) {
		closedCh := make(chan app.AppDeployLogEntry)
		close(closedCh)
		apps := &AppServiceMock{}
		tail := 100
		apps.On("AppDeployLogs", ai, &tail, false).Return(closedCh, nil)

		err := logs(context.TODO(), apps, logsInput{appDir: "", orgSlug: orgSlug, appSlug: appSlug, tail: tail, follow: false, printer: dummyPrinter})

		assert.NoError(t, err)
		apps.AssertExpectations(t)
	})
}
