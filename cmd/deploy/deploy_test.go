package deploy

import (
	"context"
	"testing"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
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

	mockVersionDeploy := func(apps *mockAppService) {
		apps.On("CreateVersion", mock.Anything, mock.Anything).Return(app.CreateAppVersionOutput{AppVersionID: appVersionID}, nil)
		apps.On("AppVersionUploadURL", mock.Anything, mock.Anything).Return(app.AppVersionUploadURLOutput{UploadURL: uploadURL}, nil)
		apps.On("UploadAppSource", mock.Anything, mock.Anything).Return(nil)
		apps.On("DeployApp", mock.Anything, mock.Anything).Return(app.DeployAppOutput{DeploymentVersionID: deployVersionID}, nil)
		apps.On("DeployEvents", mock.Anything, mock.Anything).Return(nil)
	}

	mockAppExists := func() *mockAppService {
		apps := &mockAppService{}
		apps.On("ReadApp", mock.Anything, mock.Anything).Return(app.ReadAppOutput{AppID: appID}, nil)
		mockVersionDeploy(apps)

		return apps
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

		input := DeployInput{AppDir: appDir, OrgSlug: slug, AppSlug: appSlug}
		err := Deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
	})

	t.Run("given existing app then it does not create app", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppExists()

		input := DeployInput{AppDir: appDir, OrgSlug: slug, AppSlug: appSlug}
		err := Deploy(context.TODO(), apps, input)

		assert.NoError(t, err)
		apps.AssertNotCalled(t, "Create")
	})

	t.Run("given dir without numerous.toml then it returns error", func(t *testing.T) {
		dir := t.TempDir()

		input := DeployInput{AppDir: dir, OrgSlug: slug, AppSlug: appSlug}
		err := Deploy(context.TODO(), nil, input)

		assert.EqualError(t, err, "open "+dir+"/numerous.toml: no such file or directory")
	})

	t.Run("given invalid slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		input := DeployInput{AppDir: appDir, OrgSlug: "Some Invalid Organization Slug", AppSlug: appSlug}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrInvalidOrganizationSlug)
	})

	t.Run("given no slug argument and no manifest deployment then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)

		input := DeployInput{AppDir: appDir, AppSlug: appSlug}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrMissingOrganizationSlug)
	})

	t.Run("given slug and app slug arguments and no manifest deployment then it uses arguments", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)
		apps := mockAppNotExists()

		input := DeployInput{AppDir: appDir, AppSlug: "app-slug-in-argument", OrgSlug: "organization-slug-in-argument"}
		err := Deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			apps.AssertCalled(t, "ReadApp", mock.Anything, app.ReadAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument"})
			apps.AssertCalled(t, "Create", mock.Anything, app.CreateAppInput{OrganizationSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument", DisplayName: "Streamlit App Without Deploy"})
		}
	})

	t.Run("given invalid app slug then it returns error", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)

		input := DeployInput{AppDir: appDir, OrgSlug: "organization-slug", AppSlug: "Some Invalid App Name"}
		err := Deploy(context.TODO(), nil, input)

		assert.ErrorIs(t, err, appident.ErrInvalidAppSlug)
	})

	t.Run("given no app slug argument and no manifest deployment then it converts manifest app display name", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app_without_deploy", appDir)
		apps := mockAppNotExists()

		input := DeployInput{AppDir: appDir, OrgSlug: "organization-slug"}
		err := Deploy(context.TODO(), apps, input)

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

		err := Deploy(context.TODO(), apps, DeployInput{AppDir: appDir})

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppInput{OrganizationSlug: "organization-slug-in-manifest", AppSlug: "app-slug-in-manifest", DisplayName: "Streamlit App With Deploy"}
			apps.AssertCalled(t, "Create", mock.Anything, expectedInput)
		}
	})

	t.Run("given slug or app slug arguments and manifest with deployment and then arguments override manifest deployment", func(t *testing.T) {
		appDir := t.TempDir()
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppNotExists()

		input := DeployInput{AppDir: appDir, OrgSlug: "organization-slug-in-argument", AppSlug: "app-slug-in-argument"}
		err := Deploy(context.TODO(), apps, input)

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

		input := DeployInput{AppDir: appDir, OrgSlug: slug, AppSlug: appSlug, Version: expectedVersion, Message: expectedMessage}
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
		test.CopyDir(t, "../../testdata/streamlit_app", appDir)
		apps := mockAppExists()

		input := DeployInput{AppDir: appDir, OrgSlug: slug, AppSlug: appSlug}
		err := Deploy(context.TODO(), apps, input)

		if assert.NoError(t, err) {
			expectedInput := app.CreateAppVersionInput{AppID: appID}
			apps.AssertCalled(t, "CreateVersion", mock.Anything, expectedInput)
		}
	})
}
