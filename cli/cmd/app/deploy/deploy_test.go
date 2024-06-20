package deploy

import (
	"context"
	"testing"

	"numerous/cli/internal/app"
	"numerous/cli/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeploy(t *testing.T) {
	const slug = "organization-slug"
	const appID = "app-id"
	const appName = "app-name"
	const appVersionID = "app-version-id"
	const uploadURL = "https://upload/url"
	const deployVersionID = "deploy-version-id"

	mockService := func() *mockAppService {
		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{}, app.ErrAppNotFound)
		apps.On("Create", mock.Anything, mock.Anything).Return(app.CreateAppOutput{AppID: appID}, nil)
		apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: uploadURL}, nil)
		apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)
		apps.On("DeployApp", mock.Anything, mock.Anything).Return(app.DeployAppOutput{DeploymentVersionID: deployVersionID}, nil)
		apps.On("DeployEvents", mock.Anything, mock.Anything).Return(nil)

		return apps
	}

	t.Run("give no existing app then happy path can run", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		apps := mockService()

		input := DeployInput{AppDir: appDir, Slug: slug, AppName: appName}
		err := Deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
	})

	t.Run("give existing app then it does not create app", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		apps := mockService()

		input := DeployInput{AppDir: appDir, Slug: slug, AppName: appName}
		err := Deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
	})

	t.Run("given dir without numerous.toml then it returns error", func(t *testing.T) {
		dir := t.TempDir()

		input := DeployInput{AppDir: dir, Slug: slug, AppName: appName}
		err := Deploy(context.TODO(), nil, input)

		assert.EqualError(t, err, "open "+dir+"/numerous.toml: no such file or directory")
	})

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)

		input := DeployInput{AppDir: appDir, Slug: "Some Invalid Organization Slug", AppName: appName}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, ErrInvalidSlug)
	})

	t.Run("given no slug argument and no manifest deployment then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app_without_deploy", appDir)

		input := DeployInput{AppDir: appDir, AppName: appName}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, ErrInvalidSlug)
	})

	t.Run("given invalid app name then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)

		input := DeployInput{AppDir: appDir, Slug: slug, AppName: "Some Invalid App Name"}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, ErrInvalidAppName)
	})

	t.Run("given no app name argument and no manifest deployment then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app_without_deploy", appDir)

		input := DeployInput{AppDir: appDir, Slug: slug}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, ErrInvalidAppName)
	})

	t.Run("given no slug or app name arguments and manifest with deployment then it uses manifest deployment", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		apps := mockService()

		err := Deploy(context.TODO(), apps, DeployInput{AppDir: appDir})

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppInput{OrganizationSlug: "organization-slug-in-manifest", Name: "app-name-in-manifest", DisplayName: "Streamlit App With Deploy"}
			apps.AssertCalled(t, "Create", mock.Anything, expectedInput)
		}
	})

	t.Run("given slug or app name arguments and manifest with deployment and then arguments override manifest deployment", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		apps := mockService()

		input := DeployInput{AppDir: appDir, Slug: "organization-slug-in-argument", AppName: "app-name-in-argument"}
		err := Deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppInput{OrganizationSlug: "organization-slug-in-argument", Name: "app-name-in-argument", DisplayName: "Streamlit App With Deploy"}
			apps.AssertCalled(t, "Create", mock.Anything, expectedInput)
		}
	})

	t.Run("given message and version arguments it creates app version with arguments", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		expectedVersion := "v1.2.3"
		expectedMessage := "expected message"
		apps := mockService()

		input := DeployInput{AppDir: appDir, Slug: slug, AppName: appName, Version: expectedVersion, Message: expectedMessage}
		err := Deploy(context.TODO(), apps, input)

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
		test.CopyDir(t, "../../../testdata/streamlit_app", appDir)
		apps := mockService()

		input := DeployInput{AppDir: appDir, Slug: slug, AppName: appName}
		err := Deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppVersionInput{AppID: appID}
			apps.AssertCalled(t, "CreateVersion", mock.Anything, expectedInput)
		}
	})
}
