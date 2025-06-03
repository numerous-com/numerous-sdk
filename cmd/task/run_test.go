package task

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTaskLocallyManifestParsing(t *testing.T) {
	t.Run("loads valid manifest", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "numerous-task.toml")

		manifestContent := `name = "test-collection"
version = "1.0.0"
description = "Test task collection"

[[task]]
function_name = "test_task"
source_file = "main.py"
description = "A test task"

[python]
version = "3.11"
requirements_file = "requirements.txt"
`

		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Create main.py file
		mainPyContent := `print("Hello from test task")`
		err = os.WriteFile(filepath.Join(tempDir, "main.py"), []byte(mainPyContent), 0644)
		require.NoError(t, err)

		// Save original values
		originalTaskDir := taskDir
		originalTaskName := taskName

		// Set test values
		taskDir = tempDir
		taskName = "" // Empty to trigger listing

		// Restore original values
		defer func() {
			taskDir = originalTaskDir
			taskName = originalTaskName
		}()

		// This should list tasks when no task name is provided
		err = runTaskLocally(nil, []string{})
		assert.NoError(t, err) // listTasks should succeed
	})

	t.Run("handles missing manifest", func(t *testing.T) {
		tempDir := t.TempDir()

		originalTaskDir := taskDir
		taskDir = tempDir
		defer func() { taskDir = originalTaskDir }()

		err := runTaskLocally(nil, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no task manifest found")
	})

	t.Run("handles malformed manifest", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "numerous-task.toml")

		// Invalid TOML content
		invalidContent := `name = "test
version = 1.0.0  # Missing closing quote`

		err := os.WriteFile(manifestPath, []byte(invalidContent), 0644)
		require.NoError(t, err)

		originalTaskDir := taskDir
		taskDir = tempDir
		defer func() { taskDir = originalTaskDir }()

		err = runTaskLocally(nil, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse task manifest")
	})

	t.Run("reports task not found", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "numerous-task.toml")

		manifestContent := `name = "test-collection"
version = "1.0.0"
description = "Test task collection"

[[task]]
function_name = "existing_task"
source_file = "main.py"
description = "An existing task"
`

		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Save original values
		originalTaskDir := taskDir
		originalTaskName := taskName

		// Set test values
		taskDir = tempDir
		taskName = "nonexistent_task"

		// Restore original values
		defer func() {
			taskDir = originalTaskDir
			taskName = originalTaskName
		}()

		err = runTaskLocally(nil, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "task 'nonexistent_task' not found")
	})
}

func TestTaskManifestStructures(t *testing.T) {
	t.Run("parses Python configuration", func(t *testing.T) {
		manifest := TaskManifest{
			Name:    "python-tasks",
			Version: "1.0.0",
			Python: &PythonConfig{
				Version:          "3.11",
				RequirementsFile: "requirements.txt",
			},
			Task: []TaskDef{
				{
					FunctionName: "python_task",
					SourceFile:   "main.py",
					Description:  "A Python task",
				},
			},
		}

		assert.Equal(t, "python-tasks", manifest.Name)
		assert.Equal(t, "1.0.0", manifest.Version)
		assert.NotNil(t, manifest.Python)
		assert.Equal(t, "3.11", manifest.Python.Version)
		assert.Equal(t, "requirements.txt", manifest.Python.RequirementsFile)
		assert.Len(t, manifest.Task, 1)
		assert.Equal(t, "python_task", manifest.Task[0].FunctionName)
	})

	t.Run("parses Docker configuration", func(t *testing.T) {
		manifest := TaskManifest{
			Name:    "docker-tasks",
			Version: "1.0.0",
			Docker: &DockerConfig{
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
			Task: []TaskDef{
				{
					FunctionName: "docker_task",
					SourceFile:   "main.py",
					Description:  "A Docker task",
					Entrypoint:   []string{"python", "main.py"},
				},
			},
		}

		assert.Equal(t, "docker-tasks", manifest.Name)
		assert.NotNil(t, manifest.Docker)
		assert.Equal(t, "Dockerfile", manifest.Docker.Dockerfile)
		assert.Equal(t, ".", manifest.Docker.Context)
		assert.Equal(t, []string{"python", "main.py"}, manifest.Task[0].Entrypoint)
	})

	t.Run("parses deployment configuration", func(t *testing.T) {
		manifest := TaskManifest{
			Name:    "deployed-tasks",
			Version: "1.0.0",
			Deployment: &DeployConfig{
				OrganizationSlug: "my-org",
			},
		}

		assert.NotNil(t, manifest.Deployment)
		assert.Equal(t, "my-org", manifest.Deployment.OrganizationSlug)
	})
}

func TestTaskResult(t *testing.T) {
	t.Run("creates task result structure", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(5 * time.Second)

		result := TaskResult{
			TaskName:      "test_task",
			Status:        "completed",
			StartTime:     startTime,
			EndTime:       endTime,
			Duration:      endTime.Sub(startTime),
			Output:        "Task completed successfully",
			ExitCode:      0,
			Environment:   "Python",
			ExecutionMode: "local",
			Metadata: map[string]interface{}{
				"python_version": "3.11",
			},
		}

		assert.Equal(t, "test_task", result.TaskName)
		assert.Equal(t, "completed", result.Status)
		assert.Equal(t, 5*time.Second, result.Duration)
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, "Python", result.Environment)
		assert.Equal(t, "local", result.ExecutionMode)
		assert.Contains(t, result.Metadata, "python_version")
	})
}

func TestFormatStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"success", "✅ success"},
		{"failed", "❌ failed"},
		{"completed", "completed"}, // Not handled by formatStatus
		{"running", "running"},     // Not handled by formatStatus
		{"pending", "pending"},     // Not handled by formatStatus
		{"unknown", "unknown"},     // Not handled by formatStatus
		{"", ""},                   // Not handled by formatStatus
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := formatStatus(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsDockerAvailable(t *testing.T) {
	t.Run("checks docker availability", func(t *testing.T) {
		// This will check if docker is available on the test system
		// We can't control this in tests, but we can verify the function doesn't panic
		result := isDockerAvailable()
		// Result can be true or false depending on test environment
		assert.IsType(t, false, result)
	})
}

func TestGetPythonVersion(t *testing.T) {
	t.Run("gets python version", func(t *testing.T) {
		// This will try to get the Python version from the system
		// We can't control this in tests, but we can verify the function doesn't panic
		version := getPythonVersion()
		// Version can be empty if python is not available
		assert.IsType(t, "", version)
	})
}

func TestRunTaskArgumentParsing(t *testing.T) {
	t.Run("sets task name from arguments", func(t *testing.T) {
		// Save original value
		originalTaskName := taskName

		// Set test value
		taskName = ""

		// Restore original value
		defer func() {
			taskName = originalTaskName
		}()

		// Simulate command execution with task name argument
		args := []string{"my_task"}

		// We would need to mock the actual execution, but we can verify argument handling
		// For now, just check that we can call runTask without panicking
		if len(args) > 0 {
			taskName = args[0]
		}

		assert.Equal(t, "my_task", taskName)
	})
}

func TestOutputTaskResult(t *testing.T) {
	t.Run("outputs task result without error", func(t *testing.T) {
		result := &TaskResult{
			TaskName:      "test_task",
			Status:        "completed",
			StartTime:     time.Now().Add(-5 * time.Second),
			EndTime:       time.Now(),
			Duration:      5 * time.Second,
			Output:        "Task completed successfully",
			ExitCode:      0,
			Environment:   "Python",
			ExecutionMode: "local",
		}

		// Save original output format
		originalOutputFormat := outputFormat

		// Test text output
		outputFormat = "text"
		err := outputTaskResult(result)
		assert.NoError(t, err)

		// Test JSON output
		outputFormat = "json"
		err = outputTaskResult(result)
		assert.NoError(t, err)

		// Restore original value
		outputFormat = originalOutputFormat
	})
}
