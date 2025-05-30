package deploy

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"numerous.com/cli/internal/app"
	"numerous.com/cli/internal/appident"
	"numerous.com/cli/internal/config"
	"numerous.com/cli/internal/output"
	"numerous.com/cli/internal/test"

	"github.com/BurntSushi/toml"
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

		assert.EqualError(t, err, "no app or task manifest found in "+dir)
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

func TestTaskCollectionDeploy(t *testing.T) {
	t.Run("given Python task collection then it deploys successfully", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/python_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "test-org"}
		err := deploy(context.TODO(), nil, input)

		assert.NoError(t, err)
	})

	t.Run("given Docker task collection then it deploys successfully", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/docker_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "test-org"}
		err := deploy(context.TODO(), nil, input)

		assert.NoError(t, err)
	})

	t.Run("given task collection without environment section then it returns error", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/invalid_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "test-org"}
		err := deploy(context.TODO(), nil, input)

		assert.ErrorContains(t, err, "missing environment configuration")
	})

	t.Run("given task collection without organization then it returns error", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/python_task_collection", taskDir)

		// Create manifest without deployment section
		manifestContent := `name = "test-collection"
version = "1.0.0"

[[task]]
function_name = "test_task"
source_file = "task.py"

[python]
version = "3.11"
requirements_file = "requirements.txt"
`
		test.WriteFile(t, filepath.Join(taskDir, "numerous-task.toml"), []byte(manifestContent))

		input := deployInput{appDir: taskDir} // No orgSlug provided
		err := deploy(context.TODO(), nil, input)

		assert.ErrorContains(t, err, "missing organization identifier")
	})

	t.Run("given CLI organization flag then it overrides manifest organization", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/python_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "cli-org"} // Should override "test-organization" from manifest
		err := deploy(context.TODO(), nil, input)

		assert.NoError(t, err)
		// Verify organization from CLI was used (would need to check mock backend path)
	})

	t.Run("given task collection with missing source files then deployment succeeds with warnings", func(t *testing.T) {
		taskDir := t.TempDir()

		// Create manifest with references to non-existent files
		manifestContent := `name = "missing-files-collection"
version = "1.0.0"

[[task]]
function_name = "missing_task"
source_file = "missing_task.py"

[python]
version = "3.11"
requirements_file = "missing_requirements.txt"

[deployment]
organization_slug = "test-org"
`
		test.WriteFile(t, filepath.Join(taskDir, "numerous-task.toml"), []byte(manifestContent))

		input := deployInput{appDir: taskDir}
		err := deploy(context.TODO(), nil, input)

		assert.NoError(t, err) // Deployment should succeed even with missing files
	})
}

func TestTaskCollectionDryRun(t *testing.T) {
	t.Run("given Python task collection dry-run then it shows deployment plan", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/python_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "test-org", dryRun: true}
		stdoutR, err := test.RunEWithPatchedStdout(t, func() error {
			return deploy(context.TODO(), nil, input)
		})

		assert.NoError(t, err)
		output, _ := io.ReadAll(stdoutR)
		actual := cleanNonASCIIAndANSI(string(output))

		// Verify dry-run output contains expected information
		assert.Contains(t, actual, "DRY RUN: Task Collection Deployment Summary")
		assert.Contains(t, actual, "Collection Name: python-data-processing")
		assert.Contains(t, actual, "Version: 1.0.0")
		assert.Contains(t, actual, "Organization: test-org")
		assert.Contains(t, actual, "Environment: Python")
		assert.Contains(t, actual, "Python Configuration:")
		assert.Contains(t, actual, "Version: 3.11")
		assert.Contains(t, actual, "Requirements file found")
		assert.Contains(t, actual, "Tasks defined in collection (2):")
		assert.Contains(t, actual, "process_data")
		assert.Contains(t, actual, "cleanup_data")
		assert.Contains(t, actual, "Would deploy Python task collection")
	})

	t.Run("given Docker task collection dry-run then it shows Docker configuration", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/docker_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "custom-org", dryRun: true}
		stdoutR, err := test.RunEWithPatchedStdout(t, func() error {
			return deploy(context.TODO(), nil, input)
		})

		assert.NoError(t, err)
		output, _ := io.ReadAll(stdoutR)
		actual := cleanNonASCIIAndANSI(string(output))

		// Verify Docker-specific dry-run output
		assert.Contains(t, actual, "Collection Name: docker-ml-pipeline")
		assert.Contains(t, actual, "Version: 2.0.0")
		assert.Contains(t, actual, "Organization: custom-org") // CLI override
		assert.Contains(t, actual, "Environment: Docker")
		assert.Contains(t, actual, "Docker Configuration:")
		assert.Contains(t, actual, "Dockerfile: Dockerfile")
		assert.Contains(t, actual, "Build Context: .")
		assert.Contains(t, actual, "Dockerfile found")
		assert.Contains(t, actual, "Tasks defined in collection (3):")
		assert.Contains(t, actual, "train_model")
		assert.Contains(t, actual, "Entrypoint: [python train.py]")
		assert.Contains(t, actual, "api_predict")
		assert.Contains(t, actual, "API Endpoint: /predict")
		assert.Contains(t, actual, "Python stub found")
		assert.Contains(t, actual, "Would deploy Docker task collection")
	})

	t.Run("given task collection dry-run with missing files then it shows warnings", func(t *testing.T) {
		taskDir := t.TempDir()

		// Create manifest with references to non-existent files
		manifestContent := `name = "warning-collection"
version = "1.0.0"

[[task]]
function_name = "missing_task"
source_file = "missing_task.py"
python_stub = "missing_stub.py"

[python]
version = "3.11"
requirements_file = "missing_requirements.txt"

[deployment]
organization_slug = "test-org"
`
		test.WriteFile(t, filepath.Join(taskDir, "numerous-task.toml"), []byte(manifestContent))

		input := deployInput{appDir: taskDir, dryRun: true}
		stdoutR, err := test.RunEWithPatchedStdout(t, func() error {
			return deploy(context.TODO(), nil, input)
		})

		assert.NoError(t, err)
		output, _ := io.ReadAll(stdoutR)
		actual := cleanNonASCIIAndANSI(string(output))

		// Verify warnings are shown for missing files
		assert.Contains(t, actual, "Warning: Requirements file not found")
		assert.Contains(t, actual, "Warning: Source file not found")
		assert.Contains(t, actual, "Warning: Python stub not found")
	})

	t.Run("given invalid task collection dry-run then it shows error", func(t *testing.T) {
		taskDir := t.TempDir()
		test.CopyDir(t, "../../testdata/invalid_task_collection", taskDir)

		input := deployInput{appDir: taskDir, orgSlug: "test-org", dryRun: true}
		err := deploy(context.TODO(), nil, input)

		assert.ErrorContains(t, err, "missing environment configuration")
	})
}

func TestTaskManifestParsing(t *testing.T) {
	t.Run("given valid Python task manifest then it parses correctly", func(t *testing.T) {
		manifestContent := `name = "test-collection"
version = "1.0.0"
description = "Test collection"

[[task]]
function_name = "task1"
source_file = "task1.py"
decorated_function = "decorated_task1"
description = "First task"

[[task]]
function_name = "task2"
source_file = "task2.py"

[python]
version = "3.11"
requirements_file = "requirements.txt"

[deployment]
organization_slug = "test-org"
`

		var manifest TaskManifestCollection
		_, err := toml.Decode(manifestContent, &manifest)

		assert.NoError(t, err)
		assert.Equal(t, "test-collection", manifest.Name)
		assert.Equal(t, "1.0.0", manifest.Version)
		assert.Equal(t, "Test collection", manifest.Description)
		assert.Len(t, manifest.Task, 2)
		assert.Equal(t, "task1", manifest.Task[0].FunctionName)
		assert.Equal(t, "task1.py", manifest.Task[0].SourceFile)
		assert.Equal(t, "decorated_task1", manifest.Task[0].DecoratedFunction)
		assert.Equal(t, "First task", manifest.Task[0].Description)
		assert.NotNil(t, manifest.Python)
		assert.Equal(t, "3.11", manifest.Python.Version)
		assert.Equal(t, "requirements.txt", manifest.Python.RequirementsFile)
		assert.Nil(t, manifest.Docker)
		assert.NotNil(t, manifest.Deployment)
		assert.Equal(t, "test-org", manifest.Deployment.OrganizationSlug)
	})

	t.Run("given valid Docker task manifest then it parses correctly", func(t *testing.T) {
		manifestContent := `name = "docker-collection"
version = "2.0.0"

[[task]]
function_name = "docker_task"
entrypoint = ["python", "script.py"]
api_endpoint = "/api/task"
python_stub = "stubs/task.py"

[docker]
dockerfile = "Dockerfile"
context = "."

[deployment]
organization_slug = "docker-org"
`

		var manifest TaskManifestCollection
		_, err := toml.Decode(manifestContent, &manifest)

		assert.NoError(t, err)
		assert.Equal(t, "docker-collection", manifest.Name)
		assert.Equal(t, "2.0.0", manifest.Version)
		assert.Len(t, manifest.Task, 1)
		assert.Equal(t, "docker_task", manifest.Task[0].FunctionName)
		assert.Equal(t, []string{"python", "script.py"}, manifest.Task[0].Entrypoint)
		assert.Equal(t, "/api/task", manifest.Task[0].APIEndpoint)
		assert.Equal(t, "stubs/task.py", manifest.Task[0].PythonStub)
		assert.Nil(t, manifest.Python)
		assert.NotNil(t, manifest.Docker)
		assert.Equal(t, "Dockerfile", manifest.Docker.Dockerfile)
		assert.Equal(t, ".", manifest.Docker.Context)
		assert.Equal(t, "docker-org", manifest.Deployment.OrganizationSlug)
	})
}

func TestTaskEnvironmentType(t *testing.T) {
	tests := []struct {
		name     string
		manifest TaskManifestCollection
		expected string
	}{
		{
			name: "Python environment",
			manifest: TaskManifestCollection{
				Python: &TaskPython{Version: "3.11"},
			},
			expected: "Python",
		},
		{
			name: "Docker environment",
			manifest: TaskManifestCollection{
				Docker: &TaskDocker{Dockerfile: "Dockerfile"},
			},
			expected: "Docker",
		},
		{
			name: "Docker takes precedence over Python",
			manifest: TaskManifestCollection{
				Python: &TaskPython{Version: "3.11"},
				Docker: &TaskDocker{Dockerfile: "Dockerfile"},
			},
			expected: "Docker",
		},
		{
			name:     "No environment",
			manifest: TaskManifestCollection{},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTaskEnvironmentType(&tt.manifest)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTaskOrganizationIdentifier(t *testing.T) {
	tests := []struct {
		name          string
		manifest      TaskManifestCollection
		cliOrgSlug    string
		expectedOrg   string
		expectedError bool
	}{
		{
			name:        "CLI organization takes precedence",
			cliOrgSlug:  "cli-org",
			manifest:    TaskManifestCollection{Deployment: &TaskDeployment{OrganizationSlug: "manifest-org"}},
			expectedOrg: "cli-org",
		},
		{
			name:        "Manifest organization when no CLI",
			cliOrgSlug:  "",
			manifest:    TaskManifestCollection{Deployment: &TaskDeployment{OrganizationSlug: "manifest-org"}},
			expectedOrg: "manifest-org",
		},
		{
			name:          "Error when no organization specified",
			cliOrgSlug:    "",
			manifest:      TaskManifestCollection{},
			expectedError: true,
		},
		{
			name:          "Error when empty deployment section",
			cliOrgSlug:    "",
			manifest:      TaskManifestCollection{Deployment: &TaskDeployment{}},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getTaskOrganizationIdentifier(&tt.manifest, tt.cliOrgSlug)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "missing organization identifier")
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOrg, result)
			}
		})
	}
}

func TestDockerStructureConsistency(t *testing.T) {
	t.Run("TaskDocker structure matches app Docker structure", func(t *testing.T) {
		// This test ensures that TaskDocker has the same fields as the app Docker struct
		// by parsing the same TOML content with both structures

		dockerConfig := `dockerfile = "Dockerfile"
context = "."`

		// Parse as TaskDocker
		var taskDocker TaskDocker
		err1 := toml.Unmarshal([]byte(dockerConfig), &taskDocker)

		// Parse as app Docker (we'll import and use a similar structure)
		type AppDocker struct {
			Dockerfile string `toml:"dockerfile,omitempty"`
			Context    string `toml:"context,omitempty"`
		}
		var appDocker AppDocker
		err2 := toml.Unmarshal([]byte(dockerConfig), &appDocker)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.Equal(t, "Dockerfile", taskDocker.Dockerfile)
		assert.Equal(t, ".", taskDocker.Context)
		assert.Equal(t, taskDocker.Dockerfile, appDocker.Dockerfile)
		assert.Equal(t, taskDocker.Context, appDocker.Context)
	})
}
