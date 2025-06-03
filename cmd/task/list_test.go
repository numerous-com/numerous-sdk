package task

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListLocalTasks(t *testing.T) {
	t.Run("lists tasks from valid manifest", func(t *testing.T) {
		// Create a temporary directory with a valid manifest
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "numerous-task.toml")

		manifestContent := `name = "test-collection"
version = "1.0.0"
description = "Test task collection"

[[task]]
function_name = "task_one"
source_file = "task1.py"
description = "First task"

[[task]]
function_name = "task_two"
source_file = "task2.py"
description = "Second task"

[python]
version = "3.11"
requirements_file = "requirements.txt"
`

		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Save original values
		originalDir := listDir
		originalLocalExecution := localExecution

		// Set test values
		listDir = tempDir
		localExecution = true

		// Restore original values
		defer func() {
			listDir = originalDir
			localExecution = originalLocalExecution
		}()

		// This should not error out - in a real test, we'd capture output
		err = listLocalTasks()
		assert.NoError(t, err)
	})

	t.Run("handles missing manifest", func(t *testing.T) {
		tempDir := t.TempDir()

		originalDir := listDir
		listDir = tempDir
		defer func() { listDir = originalDir }()

		err := listLocalTasks()
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

		originalDir := listDir
		listDir = tempDir
		defer func() { listDir = originalDir }()

		err = listLocalTasks()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse task manifest")
	})
}

func TestListTasks(t *testing.T) {
	t.Run("displays task information", func(t *testing.T) {
		manifest := &TaskManifest{
			Name:        "test-collection",
			Version:     "1.0.0",
			Description: "Test task collection",
			Task: []TaskDef{
				{
					FunctionName: "task_one",
					SourceFile:   "task1.py",
					Description:  "First task",
				},
				{
					FunctionName: "task_two",
					SourceFile:   "task2.py",
					Description:  "Second task",
				},
			},
			Python: &PythonConfig{
				Version:          "3.11",
				RequirementsFile: "requirements.txt",
			},
		}

		// This function mainly prints output, so we test that it doesn't error
		err := listTasks(manifest)
		assert.NoError(t, err)
	})

	t.Run("handles empty task list", func(t *testing.T) {
		manifest := &TaskManifest{
			Name:        "empty-collection",
			Version:     "1.0.0",
			Description: "Empty task collection",
			Task:        []TaskDef{}, // No tasks
		}

		err := listTasks(manifest)
		assert.NoError(t, err)
	})

	t.Run("displays Docker environment info", func(t *testing.T) {
		manifest := &TaskManifest{
			Name:        "docker-collection",
			Version:     "1.0.0",
			Description: "Docker task collection",
			Task: []TaskDef{
				{
					FunctionName: "docker_task",
					SourceFile:   "main.py",
					Description:  "Docker task",
					Entrypoint:   []string{"python", "main.py"},
				},
			},
			Docker: &DockerConfig{
				Dockerfile: "Dockerfile",
				Context:    ".",
			},
		}

		err := listTasks(manifest)
		assert.NoError(t, err)
	})
}

func TestListTasksCmd(t *testing.T) {
	t.Run("defaults to remote listing", func(t *testing.T) {
		// Save original values
		originalLocalExecution := localExecution
		originalListOrgSlug := listOrgSlug

		// Set test values
		localExecution = false
		listOrgSlug = "test-org"

		// Restore original values
		defer func() {
			localExecution = originalLocalExecution
			listOrgSlug = originalListOrgSlug
		}()

		// This will try to make remote calls, but the stub implementation just prints
		// and returns nil, so we don't expect an error
		err := listTasksCmd(nil, []string{})
		assert.NoError(t, err) // The stub implementation doesn't fail
	})

	t.Run("uses provided directory argument for local mode", func(t *testing.T) {
		tempDir := t.TempDir()
		manifestPath := filepath.Join(tempDir, "numerous-task.toml")

		manifestContent := `name = "test-collection"
version = "1.0.0"
description = "Test task collection"

[[task]]
function_name = "test_task"
source_file = "main.py"
description = "A test task"
`

		err := os.WriteFile(manifestPath, []byte(manifestContent), 0644)
		require.NoError(t, err)

		// Save original values
		originalLocalExecution := localExecution
		originalListDir := listDir

		// Set test values
		localExecution = true
		listDir = "." // Will be overridden by args

		// Restore original values
		defer func() {
			localExecution = originalLocalExecution
			listDir = originalListDir
		}()

		err = listTasksCmd(nil, []string{tempDir})
		assert.NoError(t, err)
	})
}

func TestListCollectionsInOrganization(t *testing.T) {
	t.Run("displays organization info", func(t *testing.T) {
		// This function mainly prints output and would make API calls
		// In a unit test, we just verify it doesn't panic
		err := listCollectionsInOrganization("test-org", "http://localhost:8080", "test-token")
		assert.NoError(t, err)
	})
}

func TestListTasksInCollection(t *testing.T) {
	t.Run("displays collection info", func(t *testing.T) {
		// This function mainly prints output and would make API calls
		// In a unit test, we just verify it doesn't panic
		err := listTasksInCollection("test-org", "test-collection", "http://localhost:8080", "test-token")
		assert.NoError(t, err)
	})
}
