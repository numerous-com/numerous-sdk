package task

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"numerous.com/cli/internal/output"
)

var (
	validateDir     string
	validateVerbose bool
	checkFiles      bool
	checkDocker     bool
)

var validateCmd = &cobra.Command{
	Use:   "validate [directory]",
	Short: "Validate a task collection configuration",
	Long: `Validate task collection manifest and configuration files.

This command checks:
- Task manifest syntax and structure
- Required files exist
- Docker configuration (if applicable)
- Python dependencies (if applicable)
- Task definitions are complete

Examples:
  numerous task validate
  numerous task validate ./examples/validator/python-tasks
  numerous task validate --check-files --check-docker ./examples/validator/docker-tasks`,
	RunE: validateTask,
}

func init() {
	validateCmd.Flags().StringVarP(&validateDir, "dir", "d", ".", "Directory containing the task collection")
	validateCmd.Flags().BoolVarP(&validateVerbose, "verbose", "v", false, "Enable verbose validation output")
	validateCmd.Flags().BoolVar(&checkFiles, "check-files", true, "Check that referenced files exist")
	validateCmd.Flags().BoolVar(&checkDocker, "check-docker", true, "Check Docker configuration and availability")

	validateCmd.MarkFlagDirname("dir")
}

type ValidationResult struct {
	Valid    bool                   `json:"valid"`
	Errors   []string               `json:"errors"`
	Warnings []string               `json:"warnings"`
	Summary  ValidationSummary      `json:"summary"`
	Details  map[string]interface{} `json:"details"`
}

type ValidationSummary struct {
	ManifestFound   bool   `json:"manifest_found"`
	TaskCount       int    `json:"task_count"`
	EnvironmentType string `json:"environment_type"`
	HasDocker       bool   `json:"has_docker"`
	HasPython       bool   `json:"has_python"`
	FilesChecked    int    `json:"files_checked"`
	MissingFiles    int    `json:"missing_files"`
}

func validateTask(cmd *cobra.Command, args []string) error {
	// Determine directory
	if len(args) > 0 {
		validateDir = args[0]
	}

	// Expand path
	absDir, err := filepath.Abs(validateDir)
	if err != nil {
		return fmt.Errorf("invalid directory path: %w", err)
	}
	validateDir = absDir

	if validateVerbose {
		fmt.Printf("🔍 Validating task collection in: %s\n", validateDir)
	}

	result := ValidationResult{
		Valid:    true,
		Errors:   []string{},
		Warnings: []string{},
		Details:  make(map[string]interface{}),
	}

	// Check if directory exists
	if _, err := os.Stat(validateDir); os.IsNotExist(err) {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Directory does not exist: %s", validateDir))
		return outputValidationResult(&result)
	}

	// Validate manifest
	manifestPath := filepath.Join(validateDir, "numerous-task.toml")
	manifest, err := validateManifest(manifestPath, &result)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Manifest validation failed: %v", err))
		return outputValidationResult(&result)
	}

	// Update summary
	result.Summary.ManifestFound = true
	result.Summary.TaskCount = len(manifest.Task)

	if manifest.Docker != nil {
		result.Summary.HasDocker = true
		result.Summary.EnvironmentType = "Docker"
	} else if manifest.Python != nil {
		result.Summary.HasPython = true
		result.Summary.EnvironmentType = "Python"
	} else {
		result.Summary.EnvironmentType = "Unknown"
	}

	// Validate environment configuration
	if err := validateEnvironment(manifest, &result); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Environment validation failed: %v", err))
	}

	// Validate tasks
	if err := validateTasks(manifest, &result); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf("Task validation failed: %v", err))
	}

	// Check files if requested
	if checkFiles {
		validateFiles(manifest, &result)
	}

	// Check Docker if requested and applicable
	if checkDocker && manifest.Docker != nil {
		validateDocker(manifest, &result)
	}

	return outputValidationResult(&result)
}

func validateManifest(manifestPath string, result *ValidationResult) (*TaskManifest, error) {
	if validateVerbose {
		fmt.Printf("📄 Checking manifest: %s\n", manifestPath)
	}

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		result.Summary.ManifestFound = false
		return nil, fmt.Errorf("task manifest not found: %s", manifestPath)
	}

	// Parse manifest
	var manifest TaskManifest
	if _, err := toml.DecodeFile(manifestPath, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate required fields
	if manifest.Name == "" {
		result.Errors = append(result.Errors, "manifest missing required field: name")
	}
	if manifest.Version == "" {
		result.Errors = append(result.Errors, "manifest missing required field: version")
	}

	// Check name format
	if !isValidSlug(manifest.Name) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("collection name '%s' should contain only lowercase letters, numbers, and hyphens", manifest.Name))
	}

	// Check version format
	if !isValidVersion(manifest.Version) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("version '%s' should follow semantic versioning (e.g., 1.0.0)", manifest.Version))
	}

	result.Details["manifest"] = map[string]interface{}{
		"name":        manifest.Name,
		"version":     manifest.Version,
		"description": manifest.Description,
		"task_count":  len(manifest.Task),
	}

	if validateVerbose {
		fmt.Printf("✅ Manifest parsed successfully: %s v%s\n", manifest.Name, manifest.Version)
	}

	return &manifest, nil
}

func validateEnvironment(manifest *TaskManifest, result *ValidationResult) error {
	if validateVerbose {
		fmt.Printf("🔧 Validating environment configuration\n")
	}

	// Must have either Python or Docker configuration
	if manifest.Python == nil && manifest.Docker == nil {
		result.Errors = append(result.Errors, "manifest must specify either [python] or [docker] environment section")
		return fmt.Errorf("no environment configuration found")
	}

	// Cannot have both
	if manifest.Python != nil && manifest.Docker != nil {
		result.Warnings = append(result.Warnings, "manifest has both [python] and [docker] sections, Docker takes precedence")
	}

	// Validate Python configuration
	if manifest.Python != nil {
		if manifest.Python.Version == "" {
			result.Warnings = append(result.Warnings, "Python version not specified, using default")
		}

		result.Details["python_config"] = map[string]interface{}{
			"version":           manifest.Python.Version,
			"requirements_file": manifest.Python.RequirementsFile,
		}

		if validateVerbose {
			fmt.Printf("✅ Python environment: %s\n", manifest.Python.Version)
		}
	}

	// Validate Docker configuration
	if manifest.Docker != nil {
		if manifest.Docker.Dockerfile == "" {
			manifest.Docker.Dockerfile = "Dockerfile" // Default
		}
		if manifest.Docker.Context == "" {
			manifest.Docker.Context = "." // Default
		}

		result.Details["docker_config"] = map[string]interface{}{
			"dockerfile": manifest.Docker.Dockerfile,
			"context":    manifest.Docker.Context,
		}

		if validateVerbose {
			fmt.Printf("✅ Docker environment: %s (context: %s)\n",
				manifest.Docker.Dockerfile, manifest.Docker.Context)
		}
	}

	return nil
}

func validateTasks(manifest *TaskManifest, result *ValidationResult) error {
	if validateVerbose {
		fmt.Printf("📋 Validating %d tasks\n", len(manifest.Task))
	}

	if len(manifest.Task) == 0 {
		result.Warnings = append(result.Warnings, "no tasks defined in manifest")
		return nil
	}

	// Check for duplicate task names
	taskNames := make(map[string]bool)
	for _, task := range manifest.Task {
		if taskNames[task.FunctionName] {
			result.Errors = append(result.Errors,
				fmt.Sprintf("duplicate task name: %s", task.FunctionName))
		}
		taskNames[task.FunctionName] = true
	}

	// Validate individual tasks
	for _, task := range manifest.Task {
		if err := validateSingleTask(&task, manifest, result); err != nil {
			result.Errors = append(result.Errors,
				fmt.Sprintf("task '%s': %v", task.FunctionName, err))
		}
	}

	taskDetails := make([]map[string]interface{}, len(manifest.Task))
	for i, task := range manifest.Task {
		executionMode := "unknown"
		if task.APIEndpoint != "" {
			executionMode = "api"
		} else if len(task.Entrypoint) > 0 {
			executionMode = "entrypoint"
		} else if task.SourceFile != "" {
			executionMode = "function"
		}

		taskDetails[i] = map[string]interface{}{
			"name":           task.FunctionName,
			"description":    task.Description,
			"source_file":    task.SourceFile,
			"execution_mode": executionMode,
		}
	}
	result.Details["tasks"] = taskDetails

	if validateVerbose {
		fmt.Printf("✅ All tasks validated\n")
	}

	return nil
}

func validateSingleTask(task *TaskDef, manifest *TaskManifest, result *ValidationResult) error {
	// Task name is required
	if task.FunctionName == "" {
		return fmt.Errorf("missing function_name")
	}

	// Check task name format
	if !isValidTaskName(task.FunctionName) {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("task name '%s' should contain only letters, numbers, and underscores", task.FunctionName))
	}

	// For Python tasks, source file is usually required
	if manifest.Python != nil && task.SourceFile == "" && task.APIEndpoint == "" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Python task '%s' has no source_file or api_endpoint", task.FunctionName))
	}

	// For Docker tasks, either entrypoint or API endpoint should be specified
	if manifest.Docker != nil && len(task.Entrypoint) == 0 && task.APIEndpoint == "" {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("Docker task '%s' has no entrypoint or api_endpoint", task.FunctionName))
	}

	if validateVerbose {
		fmt.Printf("  ✅ Task: %s\n", task.FunctionName)
	}

	return nil
}

func validateFiles(manifest *TaskManifest, result *ValidationResult) {
	if validateVerbose {
		fmt.Printf("📁 Checking file references\n")
	}

	filesChecked := 0
	missingFiles := 0

	// Check Python requirements file
	if manifest.Python != nil && manifest.Python.RequirementsFile != "" {
		reqPath := filepath.Join(validateDir, manifest.Python.RequirementsFile)
		filesChecked++
		if _, err := os.Stat(reqPath); os.IsNotExist(err) {
			missingFiles++
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("requirements file not found: %s", manifest.Python.RequirementsFile))
		} else if validateVerbose {
			fmt.Printf("  ✅ Requirements file: %s\n", manifest.Python.RequirementsFile)
		}
	}

	// Check Dockerfile
	if manifest.Docker != nil {
		dockerfilePath := manifest.Docker.Dockerfile
		if dockerfilePath == "" {
			dockerfilePath = "Dockerfile"
		}
		fullPath := filepath.Join(validateDir, dockerfilePath)
		filesChecked++
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			missingFiles++
			result.Errors = append(result.Errors,
				fmt.Sprintf("Dockerfile not found: %s", dockerfilePath))
		} else if validateVerbose {
			fmt.Printf("  ✅ Dockerfile: %s\n", dockerfilePath)
		}
	}

	// Check task source files
	for _, task := range manifest.Task {
		if task.SourceFile != "" {
			sourcePath := filepath.Join(validateDir, task.SourceFile)
			filesChecked++
			if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
				missingFiles++
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("task '%s' source file not found: %s", task.FunctionName, task.SourceFile))
			} else if validateVerbose {
				fmt.Printf("  ✅ Source file: %s\n", task.SourceFile)
			}
		}

		// Check Python stub files
		if task.PythonStub != "" {
			stubPath := filepath.Join(validateDir, task.PythonStub)
			filesChecked++
			if _, err := os.Stat(stubPath); os.IsNotExist(err) {
				missingFiles++
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("task '%s' Python stub not found: %s", task.FunctionName, task.PythonStub))
			} else if validateVerbose {
				fmt.Printf("  ✅ Python stub: %s\n", task.PythonStub)
			}
		}
	}

	result.Summary.FilesChecked = filesChecked
	result.Summary.MissingFiles = missingFiles

	if validateVerbose {
		fmt.Printf("📁 File check complete: %d checked, %d missing\n", filesChecked, missingFiles)
	}
}

func validateDocker(manifest *TaskManifest, result *ValidationResult) {
	if validateVerbose {
		fmt.Printf("🐳 Checking Docker configuration\n")
	}

	// Check if Docker is available
	if !isDockerAvailable() {
		result.Warnings = append(result.Warnings,
			"Docker is not available - tasks cannot be tested locally")
	} else if validateVerbose {
		fmt.Printf("  ✅ Docker is available\n")
	}
}

func outputValidationResult(result *ValidationResult) error {
	// Print summary
	if result.Valid {
		fmt.Printf("✅ Task collection validation: %s\n", output.Highlight("PASSED"))
	} else {
		fmt.Printf("❌ Task collection validation: %s\n", output.Highlight("FAILED"))
	}

	// Print summary details
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("  Environment: %s\n", result.Summary.EnvironmentType)
	fmt.Printf("  Tasks: %d\n", result.Summary.TaskCount)
	if result.Summary.FilesChecked > 0 {
		fmt.Printf("  Files checked: %d\n", result.Summary.FilesChecked)
	}

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors (%d):\n", len(result.Errors))
		for _, err := range result.Errors {
			fmt.Printf("  • %s\n", err)
		}
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Printf("\n⚠️  Warnings (%d):\n", len(result.Warnings))
		for _, warning := range result.Warnings {
			fmt.Printf("  • %s\n", warning)
		}
	}

	// Print verbose details
	if validateVerbose && len(result.Details) > 0 {
		fmt.Printf("\n🔍 Details:\n")
		if manifest, ok := result.Details["manifest"].(map[string]interface{}); ok {
			fmt.Printf("  Manifest: %s v%s\n", manifest["name"], manifest["version"])
		}
	}

	if !result.Valid {
		return fmt.Errorf("validation failed")
	}

	return nil
}

// Helper functions
func isValidSlug(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '-' {
			return false
		}
	}
	return true
}

func isValidVersion(s string) bool {
	// Simple semantic version check
	parts := strings.Split(s, ".")
	return len(parts) >= 2 && len(parts) <= 3
}

func isValidTaskName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') && !(r >= 'A' && r <= 'Z') && !(r >= '0' && r <= '9') && r != '_' {
			return false
		}
	}
	return true
}
