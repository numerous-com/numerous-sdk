package deploy

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeploy(t *testing.T) {
	const slug = "organization-slug"
	const appID = "app-id"
	const appSlug = "app-slug"
	const appVersionID = "app-version-id"
	const uploadURL = "https://upload/url"
	const deployVersionID = "deploy-version-id"

	mockVersionDeployWithDeployEventsRun := func(apps *mockAppService, deployEventsRun func(mock.Arguments)) {
		apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: uploadURL}, nil)
		apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)
		apps.On("DeployApp", mock.Anything, mock.Anything).Return(app.DeployAppOutput{DeploymentVersionID: deployVersionID}, nil)
		apps.On("DeployEvents", mock.Anything, mock.Anything).Run(deployEventsRun).Return(nil)
	}

	mockVersionDeploy := func(apps *mockAppService) {
		mockVersionDeployWithDeployEventsRun(apps, nil)
	}

	mockAppExistsWithDeployEventsRun := func(deployEventsRun func(mock.Arguments)) *mockAppService {
		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{AppID: appID}, nil)
		mockVersionDeployWithDeployEventsRun(apps, deployEventsRun)

		return apps
	}

	mockAppExists := func() *mockAppService {
		return mockAppExistsWithDeployEventsRun(nil)
	}

	mockAppNotExists := func() *mockAppService {
		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{}, app.ErrAppNotFound)
		apps.On("Create", mock.Anything, mock.Anything).Return(app.CreateAppOutput{AppID: appID}, nil)
		mockVersionDeploy(apps)

		return apps
	}

	t.Run("given no existing app then happy path can run", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppNotExists()

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug}
		err := deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
	})

	t.Run("given existing app then it does not create app", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppExists()

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug}
		err := deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
		apps.AssertNotCalled(t, "Create")
	})

	t.Run("given dir without numerous.toml then it returns error", func(t *testing.T) {
		dir := t.TempDir()

		input := deployInput{appDir: dir, orgSlug: slug, appSlug: appSlug}
		err := deploy(context.TODO(), nil, input)

		assert.EqualError(t, err, "open "+dir+"/numerous.toml: no such file or directory")
	})

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		input := deployInput{appDir: appDir, orgSlug: "Some Invalid Organization Slug", appSlug: appSlug}
		err := deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrInvalidOrganizationSlug)
	})

	t.Run("given no slug argument, no manifest deployment and no config then it returns error", func(t *testing.T) {
		oldConfigBaseDir := config.OverrideConfigBaseDir(t.TempDir())
		t.Cleanup(func() {
			config.OverrideConfigBaseDir(oldConfigBaseDir)
		})

		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)

		input := deployInput{appDir: appDir, appSlug: appSlug}
		err := deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrMissingOrganizationSlug)
	})

	t.Run("given slug and app slug arguments and no manifest deployment then it uses arguments", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)
		apps := mockAppNotExists()

		input := deployInput{appDir: appDir, appSlug: "app-slug-in-argument", orgSlug: "organization-slug-in-argument"}
		err := deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			apps.AssertCalled(t, "ReadApp", mock.Anything, app.ReadAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument"})
			apps.AssertCalled(t, "Create", mock.Anything, app.CreateAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument", DisplayName: "Streamlit App Without Deploy"})
		}
	})

	t.Run("given invalid app slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		input := deployInput{appDir: appDir, orgSlug: "organization-slug", appSlug: "Some Invalid App Name"}
		err := deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrInvalidAppSlug)
	})

	t.Run("given no app slug argument and no manifest deployment then it converts manifest app display name", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)
		apps := mockAppNotExists()

		input := deployInput{appDir: appDir, orgSlug: "organization-slug"}
		err := deploy(context.TODO(), apps, input)

		expectedAppSlug := "streamlit-app-without-deploy"
		if assert.NoError(t, err) {
			apps.AssertCalled(t, "ReadApp", mock.Anything, app.ReadAppInput{OrganizationSlug: "organization-slug", AppSlug: expectedAppSlug})
			apps.AssertCalled(t, "Create", mock.Anything, app.CreateAppInput{OrganizationSlug: "organization-slug", AppSlug: expectedAppSlug, DisplayName: "Streamlit App Without Deploy"})
		}
	})

	t.Run("given no slug or app slug arguments and manifest with deployment then it uses manifest deployment", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppNotExists()

		err := deploy(context.TODO(), apps, deployInput{appDir: appDir})

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppInput{OrganizationSlug: "organization-slug-in-manifest", AppSlug: "app-slug-in-manifest", DisplayName: "Streamlit App With Deploy"}
			apps.AssertCalled(t, "Create", mock.Anything, expectedInput)
		}
	})

	t.Run("given slug or app slug arguments and manifest with deployment and then arguments override manifest deployment", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppNotExists()

		input := deployInput{appDir: appDir, orgSlug: "organization-slug-in-argument", appSlug: "app-slug-in-argument"}
		err := deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			apps.AssertCalled(t, "ReadApp", mock.Anything, app.ReadAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument"})
			apps.AssertCalled(t, "Create", mock.Anything, app.CreateAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument", DisplayName: "Streamlit App With Deploy"})
		}
	})

	t.Run("given message and version arguments it creates app version with arguments", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		expectedVersion := "v1.2.3"
		expectedMessage := "expected message"
		apps := mockAppExists()

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug, version: expectedVersion, message: expectedMessage}
		err := deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppVersionInput{
				AppID:   appID,
				Version: expectedVersion,
				Message: expectedMessage,
			}
			apps.AssertCalled(t, "CreateVersion", mock.Anything, expectedInput)
		}
	})

	t.Run("given no message and version arguments it creates app version with empty values", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppExists()

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug}
		err := deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppVersionInput{AppID: appID}
			apps.AssertCalled(t, "CreateVersion", mock.Anything, expectedInput)
		}
	})

	t.Run("prints expected verbose messages", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		expectedVersion := "v1.2.3"
		expectedMessage := "expected message"
		apps := mockAppExistsWithDeployEventsRun(func(args mock.Arguments) {
			input := args.Get(1).(app.DeployEventsInput)
			input.Handler(app.DeployEvent{Typename: "AppBuildMessageEvent", BuildMessage: app.AppBuildMessageEvent{Message: "Build message 1"}})    // nolint:errcheck
			input.Handler(app.DeployEvent{Typename: "AppBuildMessageEvent", BuildMessage: app.AppBuildMessageEvent{Message: "Build message 2"}})    // nolint:errcheck
			input.Handler(app.DeployEvent{Typename: "AppDeploymentStatusEvent", DeploymentStatus: app.AppDeploymentStatusEvent{Status: "PENDING"}}) // nolint:errcheck
			input.Handler(app.DeployEvent{Typename: "AppDeploymentStatusEvent", DeploymentStatus: app.AppDeploymentStatusEvent{Status: "PENDING"}}) // nolint:errcheck
			input.Handler(app.DeployEvent{Typename: "AppDeploymentStatusEvent", DeploymentStatus: app.AppDeploymentStatusEvent{Status: "PENDING"}}) // nolint:errcheck
			input.Handler(app.DeployEvent{Typename: "AppDeploymentStatusEvent", DeploymentStatus: app.AppDeploymentStatusEvent{Status: "RUNNING"}}) // nolint:errcheck
		})

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug, version: expectedVersion, message: expectedMessage, verbose: true}
		stdoutR, err := test.RunEWithPatchedStdout(t, func() error {
			return deploy(context.TODO(), apps, input)
		})

		assert.NoError(t, err)
		expected := []string{
			"<non-ascii> Loading app configuration............................",
			"\r<non-ascii> Loading app configuration............................OK\n",
			"<non-ascii> Registering new version for organization-slug/app-...",
			"\r<non-ascii> Registering new version for organization-slug/app-...OK\n",
			"<non-ascii> Creating app archive.................................",
			"\r<non-ascii> Creating app archive.................................OK\n",
			"<non-ascii> Uploading app archive................................",
			"\r<non-ascii> Uploading app archive................................OK\n",
			"<non-ascii> Deploying app........................................\n",
			"Build Build message 1\n",
			"Build Build message 2\n",
			"\rDeploy Workload is pending",
			"\rDeploy Workload is pending.",
			"\rDeploy Workload is pending..\n",
			"\rDeploy Workload is running\n",
			"<non-ascii> Deploying app........................................OK\n",
			"<non-ascii> Access your app at: https://www.numerous.com/app/organization/organization-slug/private/app-slug\n",
			"\n",
			"To read the logs from your app you can:\n",
			"  numerous logs --organization=organization-slug --app=app-slug\n",
			"Or you can use the --follow flag:\n",
			"  numerous deploy --follow --organization=organization-slug --app=app-slug " + appDir + "\n",
		}
		output, _ := io.ReadAll(stdoutR)
		actual := cleanNonASCIIAndANSI(string(output))
		assert.Equal(t, strings.Join(expected, ""), actual)
	})

	t.Run("given follow flag it reads deployment logs", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		expectedVersion := "v1.2.3"
		expectedMessage := "expected message"
		apps := mockAppExists()
		ch := make(chan app.AppDeployLogEntry, 2)
		expectedEntry1 := app.AppDeployLogEntry{Timestamp: time.Date(2024, time.January, 1, 1, 1, 1, 0, time.UTC), Text: "Log entry 1"}
		expectedEntry2 := app.AppDeployLogEntry{Timestamp: time.Date(2024, time.January, 1, 1, 1, 2, 0, time.UTC), Text: "Log entry 2"}
		ch <- expectedEntry1
		ch <- expectedEntry2
		close(ch)
		apps.On("AppDeployLogs", appident.AppIdentifier{OrganizationSlug: slug, AppSlug: appSlug}, (*int)(nil), true).Once().Return(ch, nil)

		input := deployInput{appDir: appDir, orgSlug: slug, appSlug: appSlug, version: expectedVersion, message: expectedMessage, verbose: true, follow: true}
		stdoutR, err := test.RunEWithPatchedStdout(t, func() error {
			return deploy(context.TODO(), apps, input)
		})

		assert.NoError(t, err)
		output, _ := io.ReadAll(stdoutR)
		actual := cleanNonASCIIAndANSI(string(output))

		expected := strings.Join(
			[]string{
				"Following logs of organization-slug/app-slug:",
				expectedEntry1.Timestamp.Format(time.RFC3339Nano) + " " + expectedEntry1.Text,
				expectedEntry2.Timestamp.Format(time.RFC3339Nano) + " " + expectedEntry2.Text,
			}, "\n",
		)
		assert.Contains(t, actual, expected)
		apps.AssertExpectations(t)
	})
}

// Strips output of known ANSI terminal escapes, and non-ascii runes (e.g. icons).
func cleanNonASCIIAndANSI(s string) string {
	var cleaned string
	for _, r := range s {
		if r < 128 {
			cleaned += string(r)
		} else {
			cleaned += "<non-ascii>"
		}
	}

	for _, code := range []string{output.AnsiRed, output.AnsiReset, output.AnsiGreen, output.AnsiFaint, output.AnsiCyanBold} {
		cleaned = strings.ReplaceAll(cleaned, code, "")
	}

	return cleaned
}
