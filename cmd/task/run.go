package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/hasura/go-graphql-client"
	"github.com/spf13/cobra"
	"numerous.com/cli/internal/gql"
	"numerous.com/cli/internal/output"
)

var (
	taskName       string
	taskArgs       []string
	taskDir        string
	localExecution bool
	noDocker       bool
	verbose        bool
	outputFormat   string
	timeoutSeconds int
	orgSlug        string
	collectionName string
	follow         bool
)

var runCmd = &cobra.Command{
	Use:   "run [task-name]",
	Short: "Run a task from a task collection",
	Long: `Run a specific task from a deployed task collection.

By default, this command runs tasks remotely on the Numerous platform using deployed task collections.
Use --local to run tasks locally for development and testing.

Organization is required for all executions to identify which workspace to use.

Remote execution examples:
  numerous task run validate_environment --org my-org --collection my-tasks
  numerous task run process_data --org my-org --collection my-tasks --args '["input.csv"]'
  numerous task run validate_environment --org my-org  # Will list collections if only one exists

Local execution examples:
  numerous task run validate_environment --org my-org --local --task-dir ./examples/validator/python-tasks
  numerous task run process_data --org my-org --local --task-dir ./examples/validator/python-tasks
  numerous task run --org my-org --local --list-tasks --task-dir ./examples/validator/docker-tasks`,
	RunE: runTask,
}

func init() {
	runCmd.Flags().StringVarP(&taskName, "task", "t", "", "Name of the task to run")
	runCmd.Flags().StringSliceVarP(&taskArgs, "args", "a", []string{}, "Arguments to pass to the task (JSON array for remote execution)")
	runCmd.Flags().BoolVar(&localExecution, "local", false, "Run task locally instead of on the platform")
	runCmd.Flags().StringVarP(&taskDir, "task-dir", "d", ".", "Directory containing the task collection (local execution only)")
	runCmd.Flags().BoolVar(&noDocker, "no-docker", false, "Run Docker tasks locally without Docker (local execution only)")
	runCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	runCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format: text, json")
	runCmd.Flags().IntVar(&timeoutSeconds, "timeout", 300, "Task execution timeout in seconds")
	runCmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow task execution and wait for completion")

	// Organization is mandatory for all modes
	runCmd.Flags().StringVar(&orgSlug, "organization", "", "Organization slug (required)")
	runCmd.MarkFlagRequired("organization")

	// Collection is optional - if not specified, we can list available collections or infer from context
	runCmd.Flags().StringVarP(&collectionName, "collection", "c", "", "Task collection name (optional for remote execution)")

	// Mark directories for local execution
	runCmd.MarkFlagDirname("task-dir")
}

type TaskManifest struct {
	Name        string        `toml:"name"`
	Version     string        `toml:"version"`
	Description string        `toml:"description"`
	Task        []TaskDef     `toml:"task"`
	Python      *PythonConfig `toml:"python,omitempty"`
	Docker      *DockerConfig `toml:"docker,omitempty"`
	Deployment  *DeployConfig `toml:"deployment,omitempty"`
}

type TaskDef struct {
	FunctionName      string   `toml:"function_name"`
	SourceFile        string   `toml:"source_file"`
	DecoratedFunction string   `toml:"decorated_function,omitempty"`
	Description       string   `toml:"description,omitempty"`
	Entrypoint        []string `toml:"entrypoint,omitempty"`
	APIEndpoint       string   `toml:"api_endpoint,omitempty"`
	PythonStub        string   `toml:"python_stub,omitempty"`
}

type PythonConfig struct {
	Version          string `toml:"version"`
	RequirementsFile string `toml:"requirements_file"`
}

type DockerConfig struct {
	Dockerfile string `toml:"dockerfile,omitempty"`
	Context    string `toml:"context,omitempty"`
}

type DeployConfig struct {
	OrganizationSlug string `toml:"organization_slug,omitempty"`
}

type TaskResult struct {
	TaskName      string                 `json:"task_name"`
	Status        string                 `json:"status"`
	StartTime     time.Time              `json:"start_time"`
	EndTime       time.Time              `json:"end_time"`
	Duration      time.Duration          `json:"duration"`
	Output        string                 `json:"output"`
	Error         string                 `json:"error,omitempty"`
	ExitCode      int                    `json:"exit_code"`
	Environment   string                 `json:"environment"`
	ExecutionMode string                 `json:"execution_mode"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// TaskExecutionLogsSubscription defines the subscription for task execution logs
type TaskExecutionLogsSubscription struct {
	TaskExecutionLogs TaskLogEntry `graphql:"taskExecutionLogs(input: {taskExecutionId: $taskExecutionId})"`
}

// TaskLogEntry represents a single log entry from the task execution
type TaskLogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
}

func runTask(cmd *cobra.Command, args []string) error {
	// Determine task name
	if len(args) > 0 {
		taskName = args[0]
	}

	// Validate required parameters based on execution mode
	if localExecution {
		return runTaskLocally(cmd, args)
	} else {
		return runTaskRemotely(cmd, args)
	}
}

func runTaskLocally(cmd *cobra.Command, args []string) error {
	// Load task manifest
	manifestPath := filepath.Join(taskDir, "numerous-task.toml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return fmt.Errorf("no task manifest found at %s (expected in --task-dir, which defaults to current directory)", manifestPath)
	}

	var manifest TaskManifest
	if _, err := toml.DecodeFile(manifestPath, &manifest); err != nil {
		return fmt.Errorf("failed to parse task manifest '%s': %w", manifestPath, err)
	}

	// If no task specified, list available tasks
	if taskName == "" {
		return listTasks(&manifest)
	}

	// Find the task
	var task *TaskDef
	for i := range manifest.Task { // Iterate by index to get a pointer
		if manifest.Task[i].FunctionName == taskName {
			task = &manifest.Task[i]
			break
		}
	}

	if task == nil {
		output.PrintError("Task not found: %s in manifest %s", "", taskName, manifestPath)
		fmt.Println("Available tasks:")
		for _, t := range manifest.Task {
			fmt.Printf("  - %s: %s\n", t.FunctionName, t.Description)
		}
		return fmt.Errorf("task '%s' not found in manifest %s", taskName, manifestPath)
	}

	// Determine environment type and check for potential misconfigurations
	envType := "unknown"
	dockerfilePresent := false
	dockerfilePath := filepath.Join(taskDir, "Dockerfile")
	if _, err := os.Stat(dockerfilePath); err == nil {
		dockerfilePresent = true
	}

	if manifest.Docker != nil {
		envType = "docker"
	} else if manifest.Python != nil {
		// [python] section exists, but no [docker] section.
		if dockerfilePresent {
			return fmt.Errorf("ambiguous configuration in '%s': numerous-task.toml has a [python] section and a 'Dockerfile' exists, but no [docker] section was found. "+
				"If this is a Docker-based task collection, add a [docker] section to numerous-task.toml. "+
				"If it is a Python task collection intended for direct execution, remove the 'Dockerfile' from '%s' to avoid ambiguity.", taskDir, taskDir)
		}
		envType = "python" // Clear Python task: [python] present, no [docker], no Dockerfile.
	} else if dockerfilePresent {
		// Dockerfile exists, but manifest has neither [docker] nor [python] task configuration sections.
		return fmt.Errorf("incomplete manifest in '%s': a 'Dockerfile' exists, but the numerous-task.toml lacks a [docker] section to define how to build/run tasks, and also lacks a [python] section. "+
			"Please define your tasks and their execution environment (e.g., [docker] or [python]) in numerous-task.toml.", taskDir)
	}
	// If envType remains "unknown" here, it means manifest is missing [docker] and [python] sections, and no Dockerfile exists.
	// The switch statement's default case will handle this.

	if verbose {
		fmt.Printf("🎯 Running task locally: %s\n", taskName)
		fmt.Printf(" Collection: %s v%s (from manifest: %s)\n", manifest.Name, manifest.Version, manifestPath)
		fmt.Printf("🔧 Determined Environment Type: %s\n", envType)
		if manifest.Docker != nil {
			fmt.Printf("   - Found [docker] section in manifest.\n")
			if manifest.Docker.Dockerfile != "" {
				fmt.Printf("     - Custom Dockerfile specified in manifest: %s\n", manifest.Docker.Dockerfile)
			}
			if manifest.Docker.Context != "" {
				fmt.Printf("     - Custom Docker context specified in manifest: %s\n", manifest.Docker.Context)
			}
		}
		if manifest.Python != nil {
			fmt.Printf("   - Found [python] section in manifest.\n")
		}
		if dockerfilePresent {
			fmt.Printf("   - Found 'Dockerfile' at: %s\n", dockerfilePath)
		} else {
			fmt.Printf("   - No 'Dockerfile' found at: %s\n", dockerfilePath)
		}
		fmt.Printf("📁 Task Directory (--task-dir): %s\n", taskDir)
	}

	// Execute the task based on environment type
	var result TaskResult
	var err error

	switch envType {
	case "python":
		result, err = runPythonTask(&manifest, task)
	case "docker":
		if noDocker {
			result, err = runDockerTaskLocally(&manifest, task)
		} else {
			result, err = runDockerTask(&manifest, task)
		}
	default:
		return fmt.Errorf("unknown environment type for task collection")
	}

	if err != nil {
		return fmt.Errorf("task execution failed: %w", err)
	}

	// Output results
	return outputTaskResult(&result)
}

func runTaskRemotely(cmd *cobra.Command, args []string) error {
	// Validate required parameters for remote execution
	if taskName == "" {
		return fmt.Errorf("task name is required for remote execution")
	}

	// Organization is now required by flag validation, but let's be explicit
	if orgSlug == "" {
		return fmt.Errorf("organization slug is required (use --organization flag)")
	}

	// Collection is optional - if not provided, we should try to infer or list available collections
	if collectionName == "" {
		return fmt.Errorf("collection name is required for task execution (use --collection flag)\nTo list available collections, use: numerous task list --organization %s", orgSlug)
	}

	if verbose {
		fmt.Printf("🌐 Running task remotely: %s\n", taskName)
		fmt.Printf("🏢 Organization: %s\n", orgSlug)
		fmt.Printf("📦 Collection: %s\n", collectionName)
	}

	// Execute the remote task using the GraphQL infrastructure
	result, err := executeRemoteTask(taskName, taskArgs, orgSlug, collectionName)
	if err != nil {
		return fmt.Errorf("remote task execution failed: %w", err)
	}

	// Output results
	return outputTaskResult(&result)
}

func listTasks(manifest *TaskManifest) error {
	fmt.Printf("📋 Task Collection: %s v%s\n", manifest.Name, manifest.Version)
	if manifest.Description != "" {
		fmt.Printf("📝 Description: %s\n", manifest.Description)
	}

	envType := "unknown"
	if manifest.Docker != nil {
		envType = "Docker"
	} else if manifest.Python != nil {
		envType = "Python"
	}
	fmt.Printf("🔧 Environment: %s\n", envType)

	fmt.Printf("\n📋 Available tasks (%d):\n", len(manifest.Task))

	for i, task := range manifest.Task {
		fmt.Printf("  %d. %s\n", i+1, output.Highlight(task.FunctionName))
		if task.Description != "" {
			fmt.Printf("     Description: %s\n", task.Description)
		}

		// Show execution method
		if len(task.Entrypoint) > 0 {
			fmt.Printf("     Entrypoint: %v\n", task.Entrypoint)
		} else if task.SourceFile != "" {
			fmt.Printf("     Source: %s", task.SourceFile)
			if task.DecoratedFunction != "" && task.DecoratedFunction != task.FunctionName {
				fmt.Printf(" (%s)", task.DecoratedFunction)
			}
			fmt.Println()
		}

		if task.APIEndpoint != "" {
			fmt.Printf("     API Endpoint: %s\n", task.APIEndpoint)
		}

		fmt.Println()
	}

	fmt.Println("Usage:")
	fmt.Printf("  numerous task run <task-name> --task-dir %s\n", taskDir)

	return nil
}

func runPythonTask(manifest *TaskManifest, task *TaskDef) (TaskResult, error) {
	result := TaskResult{
		TaskName:      task.FunctionName,
		Environment:   "python",
		ExecutionMode: "direct",
		StartTime:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	if task.SourceFile == "" {
		result.Status = "failed"
		result.Error = "no source file specified for Python task"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("no source file specified")
	}

	// Check if source file exists
	sourceFilePath := filepath.Join(taskDir, task.SourceFile)
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		result.Status = "failed"
		result.Error = fmt.Sprintf("source file not found: %s", task.SourceFile)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("source file not found: %s", sourceFilePath)
	}

	// Prepare Python execution
	var cmd *exec.Cmd

	// Try to call specific function if it's a module
	if strings.HasSuffix(task.SourceFile, ".py") {
		// Execute Python file directly
		cmd = exec.Command("python", sourceFilePath)

		// Add any arguments
		if len(taskArgs) > 0 {
			cmd.Args = append(cmd.Args, taskArgs...)
		}
	}

	if cmd == nil {
		result.Status = "failed"
		result.Error = "unable to determine Python execution method"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("unable to determine execution method")
	}

	// Set working directory
	cmd.Dir = taskDir

	// Set environment variables
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PYTHONPATH="+taskDir)

	if verbose {
		fmt.Printf("🔧 Executing: %s\n", strings.Join(cmd.Args, " "))
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmd.Args[0], cmd.Args[1:]...)
	cmd.Dir = taskDir
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "PYTHONPATH="+taskDir)

	output, err := cmd.CombinedOutput()
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.Status = "success"
		result.ExitCode = 0
	}

	// Add metadata
	result.Metadata["python_version"] = getPythonVersion()
	result.Metadata["working_directory"] = taskDir
	result.Metadata["source_file"] = task.SourceFile

	return result, nil
}

func runDockerTask(manifest *TaskManifest, task *TaskDef) (TaskResult, error) {
	result := TaskResult{
		TaskName:      task.FunctionName,
		Environment:   "docker",
		ExecutionMode: "container (local run of deployed image)",
		StartTime:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	if !isDockerAvailable() {
		result.Status = "failed"
		result.Error = "Docker is not available or not running"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("Docker not available")
	}

	// Determine the deployed image name
	// OrgSlug is a global flag, manifest.Name and manifest.Version come from the loaded manifest
	if orgSlug == "" {
		// This should ideally be caught by earlier validation if orgSlug is always required for local docker run
		result.Status = "failed"
		result.Error = "Organization slug not provided (required to identify deployed image)"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("organization slug not provided")
	}
	imageName := fmt.Sprintf("numerous-tasks/%s/%s:%s", orgSlug, manifest.Name, manifest.Version)

	if verbose {
		fmt.Printf("🐳 Attempting to run deployed Docker image: %s\n", imageName)
		fmt.Printf("Task to run in container: %s\n", task.FunctionName)
	}

	// Check if image exists locally (optional, docker run will fail if not found anyway)
	// inspectCmd := exec.Command("docker", "image", "inspect", imageName)
	// if err := inspectCmd.Run(); err != nil {
	// 	result.Status = "failed"
	// 	result.Error = fmt.Sprintf("Deployed image %s not found locally. Please deploy the collection first or pull the image.", imageName)
	// 	result.EndTime = time.Now()
	// 	result.Duration = result.EndTime.Sub(result.StartTime)
	// 	return result, fmt.Errorf("image %s not found: %w", imageName, err)
	// }

	// Prepare arguments for docker run
	// Base: docker run --rm <imageName>
	// Args to container: <taskDef.FunctionName> [taskArgs...]
	dockerRunArgs := []string{"run", "--rm"} // --network=host could be useful for local dev if tasks need to reach host services

	// Pass environment variables from current shell (optional, but can be useful for local dev)
	// for _, e := range os.Environ() {
	// 	dockerRunArgs = append(dockerRunArgs, "-e", e)
	// }

	dockerRunArgs = append(dockerRunArgs, imageName)
	dockerRunArgs = append(dockerRunArgs, task.FunctionName) // First arg to entrypoint is task name
	dockerRunArgs = append(dockerRunArgs, taskArgs...)       // Subsequent args are task inputs

	runCmd := exec.Command("docker", dockerRunArgs...)

	if verbose {
		fmt.Printf("🚀 Executing command: %s\n", strings.Join(runCmd.Args, " "))
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Re-assign runCmd to include context, ensuring the original args are used
	runCmd = exec.CommandContext(ctx, "docker", dockerRunArgs...)
	var combinedOutput bytes.Buffer
	runCmd.Stdout = &combinedOutput
	runCmd.Stderr = &combinedOutput

	err := runCmd.Run()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = combinedOutput.String()

	if err != nil {
		result.Status = "failed"
		// If context deadline exceeded, it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			result.Error = fmt.Sprintf("task timed out after %d seconds", timeoutSeconds)
		} else {
			result.Error = err.Error() // Includes docker command error and stderr
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1 // Default error code
		}
	} else {
		result.Status = "success"
		result.ExitCode = 0
	}

	result.Metadata["docker_image_used"] = imageName
	result.Metadata["ran_deployed_image"] = true

	return result, nil
}

func runDockerTaskLocally(manifest *TaskManifest, task *TaskDef) (TaskResult, error) {
	result := TaskResult{
		TaskName:      task.FunctionName,
		Environment:   "docker-local",
		ExecutionMode: "local",
		StartTime:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// For Docker tasks without Docker, try to run the entrypoint locally
	if len(task.Entrypoint) == 0 {
		result.Status = "failed"
		result.Error = "no entrypoint specified for Docker task"
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, fmt.Errorf("no entrypoint specified")
	}

	// Execute the entrypoint locally
	cmdArgs := append(task.Entrypoint, taskArgs...)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = taskDir

	if verbose {
		fmt.Printf("🔧 Running locally: %s\n", strings.Join(cmdArgs, " "))
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	cmd = exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
	cmd.Dir = taskDir
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = 1
		}
	} else {
		result.Status = "success"
		result.ExitCode = 0
	}

	// Add metadata
	result.Metadata["entrypoint"] = task.Entrypoint
	result.Metadata["local_execution"] = true

	return result, nil
}

func outputTaskResult(result *TaskResult) error {
	switch outputFormat {
	case "json":
		jsonData, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal result to JSON: %w", err)
		}
		fmt.Println(string(jsonData))
	case "text":
	default:
		// Text output
		fmt.Printf("\n🎯 Task Execution Result\n")
		fmt.Printf("Task: %s\n", result.TaskName)
		fmt.Printf("Status: %s\n", formatStatus(result.Status))
		fmt.Printf("Environment: %s\n", result.Environment)
		fmt.Printf("Execution Mode: %s\n", result.ExecutionMode)
		fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))
		fmt.Printf("Exit Code: %d\n", result.ExitCode)

		if result.Error != "" {
			fmt.Printf("Error: %s\n", result.Error)
		}

		if result.Output != "" {
			fmt.Printf("\n📄 Output:\n%s\n", result.Output)
		}

		if verbose && len(result.Metadata) > 0 {
			fmt.Printf("\n🔍 Metadata:\n")
			for key, value := range result.Metadata {
				fmt.Printf("  %s: %v\n", key, value)
			}
		}
	}

	return nil
}

func formatStatus(status string) string {
	switch status {
	case "success":
		return "✅ " + status
	case "failed":
		return "❌ " + status
	default:
		return status
	}
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	err := cmd.Run()
	return err == nil
}

func getPythonVersion() string {
	cmd := exec.Command("python", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func executeRemoteTask(taskName string, taskArgs []string, orgSlug, collectionName string) (TaskResult, error) {
	result := TaskResult{
		TaskName:      taskName,
		Environment:   "remote",
		ExecutionMode: "api",
		StartTime:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Use the main API with task execution endpoints
	apiURL := gql.GetHTTPURL()
	accessToken := gql.GetAccessToken()

	// Create HTTP client with timeout
	httpClient := &http.Client{Timeout: time.Duration(timeoutSeconds) * time.Second}

	// Prepare arguments for the task
	// For remote execution, we construct the function identifier and arguments
	functionIdentifier := fmt.Sprintf("%s.%s.%s", orgSlug, collectionName, taskName)

	// Convert arguments to JSON strings as expected by the API
	argsJSON := "[]"
	kwargsJSON := "{}"

	if len(taskArgs) > 0 {
		if len(taskArgs) == 1 {
			// Try to parse single argument as JSON
			argsJSON = taskArgs[0]
		} else {
			// Multiple arguments - convert to JSON array
			argsBytes, err := json.Marshal(taskArgs)
			if err != nil {
				result.Status = "failed"
				result.Error = fmt.Sprintf("failed to marshal arguments: %v", err)
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
				return result, err
			}
			argsJSON = string(argsBytes)
		}
	}

	// Create GraphQL mutation using the correct start_task mutation
	mutation := `
		mutation StartTask($function: String!, $args: String!, $kwargs: String!, $filePath: String, $sessionName: String) {
			start_task(function: $function, args: $args, kwargs: $kwargs, filePath: $filePath, sessionName: $sessionName) {
				id
				status
				error
				created_at
				started_at
				completed_at
				result
				logTopicId
			}
		}
	`

	variables := map[string]interface{}{
		"function":    functionIdentifier,
		"args":        argsJSON,
		"kwargs":      kwargsJSON,
		"filePath":    nil, // For deployed tasks, file path is not needed
		"sessionName": nil, // Could be made configurable in the future
	}

	requestBody := map[string]interface{}{
		"query":     mutation,
		"variables": variables,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to marshal request: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}

	req.Header.Set("Content-Type", "application/json")
	if accessToken != nil {
		req.Header.Set("Authorization", "Bearer "+*accessToken)
	}

	if verbose {
		fmt.Printf("🔗 Sending request to: %s\n", apiURL)
		fmt.Printf("📧 Function: %s\n", functionIdentifier)
		fmt.Printf("📥 Args: %s\n", argsJSON)
	}

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("request failed: %v", err)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, err
	}
	defer resp.Body.Close()

	// Parse response
	var response struct {
		Data struct {
			StartTask struct {
				ID          string   `json:"id"`
				Status      string   `json:"status"`
				Error       *string  `json:"error"`
				CreatedAt   float64  `json:"created_at"`   // Revert to float64 for numeric timestamps
				StartedAt   *float64 `json:"started_at"`   // Revert to *float64
				CompletedAt *float64 `json:"completed_at"` // Revert to *float64
				Result      *string  `json:"result"`
				LogTopicId  *string  `json:"logTopicId"`
			} `json:"start_task"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	bodyBytes, errRead := io.ReadAll(resp.Body)
	if errRead != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to read response body: %v", errRead)
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, errRead
	}
	// Re-initialize resp.Body as it has been read
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		result.Status = "failed"
		// Include raw body in error for better debugging
		result.Error = fmt.Sprintf("failed to decode response: %v. Raw Body: %s", err, string(bodyBytes))
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		// In verbose mode, also print the raw response for debugging
		if verbose {
			fmt.Printf("🐛 Debug: Raw API Response: %s\n", string(bodyBytes))
		}

		return result, err
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Check for GraphQL errors
	if len(response.Errors) > 0 {
		result.Status = "failed"
		result.Error = response.Errors[0].Message
		return result, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	// Check for task execution errors
	startTaskData := response.Data.StartTask
	if startTaskData.Error != nil {
		result.Status = "failed"
		result.Error = *startTaskData.Error
		return result, fmt.Errorf("task execution failed: %s", *startTaskData.Error)
	}

	// If task is already completed, return immediately
	if startTaskData.Status == "completed" || startTaskData.Status == "failed" {
		result.Status = startTaskData.Status
		if startTaskData.Status == "completed" {
			result.ExitCode = 0
		} else {
			result.ExitCode = 1
		}

		// Set timing information from numeric timestamps
		// CreatedAt is non-nullable, convert from Unix timestamp
		result.StartTime = time.Unix(int64(startTaskData.CreatedAt), 0)

		if startTaskData.StartedAt != nil {
			// Prefer StartedAt if available
			result.StartTime = time.Unix(int64(*startTaskData.StartedAt), 0)
		}

		if startTaskData.CompletedAt != nil {
			result.EndTime = time.Unix(int64(*startTaskData.CompletedAt), 0)
			result.Duration = result.EndTime.Sub(result.StartTime)
		} else {
			// If CompletedAt is not provided, use current time for completed/failed tasks
			if startTaskData.Status == "completed" || startTaskData.Status == "failed" {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
			}
		}

		// Add metadata
		result.Metadata["task_id"] = startTaskData.ID
		result.Metadata["execution_id"] = startTaskData.ID
		result.Metadata["organization"] = orgSlug
		result.Metadata["collection"] = collectionName
		result.Metadata["function"] = functionIdentifier

		if startTaskData.Result != nil {
			result.Metadata["result"] = *startTaskData.Result
		}

		fmt.Printf("⏳ Task is %s. Use --follow to wait for completion.\n", startTaskData.Status)
		fmt.Printf("💡 Check status with: numerous task status %s --organization %s\n", startTaskData.ID, orgSlug)

		return result, nil
	}

	// If --follow flag is not set, return immediately with current status
	if !follow {
		result.Status = startTaskData.Status
		result.StartTime = time.Unix(int64(startTaskData.CreatedAt), 0)
		if startTaskData.StartedAt != nil {
			result.StartTime = time.Unix(int64(*startTaskData.StartedAt), 0)
		}
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)

		// Add metadata
		result.Metadata["task_id"] = startTaskData.ID
		result.Metadata["execution_id"] = startTaskData.ID
		result.Metadata["organization"] = orgSlug
		result.Metadata["collection"] = collectionName
		result.Metadata["function"] = functionIdentifier

		fmt.Printf("⏳ Task is %s. Use --follow to wait for completion.\n", startTaskData.Status)
		fmt.Printf("💡 Check status with: numerous task status %s --organization %s\n", startTaskData.ID, orgSlug)

		return result, nil
	}

	// For running tasks with --follow flag, poll until completion
	if verbose {
		fmt.Printf("⏳ Following task execution, polling for completion...\n")
	}

	// If follow flag is set, start log streaming with shared completion context
	completionCtx, completionCancel := context.WithCancel(context.Background())
	logStreamDone := make(chan bool)
	if follow {
		go func() {
			// Determine which ID to use for subscription
			subscriptionTopicId := startTaskData.ID // Default to task ID
			if startTaskData.LogTopicId != nil && *startTaskData.LogTopicId != "" {
				subscriptionTopicId = *startTaskData.LogTopicId // Use log topic ID if available
				if verbose {
					fmt.Printf("📡 Using log topic ID for subscription: %s\n", subscriptionTopicId)
				}
			} else if verbose {
				fmt.Printf("📡 No log topic ID provided, using task ID for subscription: %s\n", subscriptionTopicId)
			}

			// Try WebSocket subscription first, fallback to HTTP polling if it fails
			err := streamTaskLogsViaSubscription(subscriptionTopicId, apiURL, accessToken, completionCtx)
			if err != nil {
				if verbose {
					fmt.Printf("⚠️  WebSocket subscription failed, falling back to HTTP polling: %v\n", err)
				}
				// Fallback to HTTP polling (still uses task ID for HTTP polling)
				err = streamTaskLogs(startTaskData.ID, apiURL, accessToken, httpClient, completionCtx)
				if err != nil && verbose {
					fmt.Printf("⚠️  Log streaming error: %v\n", err)
				}
			}
			completionCancel() // Signal completion detected
			logStreamDone <- true
		}()
	}

	finalResult, err := pollTaskCompletion(startTaskData.ID, apiURL, accessToken, httpClient, completionCtx, completionCancel)
	if err != nil {
		result.Status = "failed"
		result.Error = fmt.Sprintf("failed to poll task completion: %v", err)
		completionCancel() // Ensure cleanup
		return result, err
	}

	// Wait for log streaming to complete if follow was enabled
	if follow {
		<-logStreamDone
	}
	completionCancel() // Ensure cleanup

	// Merge the polling result with our base result
	finalResult.TaskName = result.TaskName
	finalResult.Environment = result.Environment
	finalResult.ExecutionMode = result.ExecutionMode
	finalResult.Metadata["task_id"] = startTaskData.ID
	finalResult.Metadata["execution_id"] = startTaskData.ID
	finalResult.Metadata["organization"] = orgSlug
	finalResult.Metadata["collection"] = collectionName
	finalResult.Metadata["function"] = functionIdentifier

	return finalResult, nil
}

func pollTaskCompletion(taskID string, apiURL string, accessToken *string, httpClient *http.Client, completionCtx context.Context, completionCancel context.CancelFunc) (TaskResult, error) {
	result := TaskResult{
		TaskName:      "",
		Environment:   "remote",
		ExecutionMode: "api",
		StartTime:     time.Now(),
		Metadata:      make(map[string]interface{}),
	}

	// Query to get task status
	query := `
		query TaskStatus($taskId: ID!) {
			task_status(taskId: $taskId) {
				id
				status
				result
				error
				created_at
				started_at
				completed_at
				progress
				status_message
			}
		}
	`

	maxAttempts := timeoutSeconds / 2 // Poll every 2 seconds
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check if completion was detected by log streaming
		select {
		case <-completionCtx.Done():
			// Completion detected by log streaming, make one final status check and return
			if verbose {
				fmt.Printf("📊 Completion detected by log streaming, making final status check...\n")
			}
			// Fall through to make final status check below
		default:
			// Continue with normal polling
		}

		variables := map[string]interface{}{
			"taskId": taskID,
		}

		requestBody := map[string]interface{}{
			"query":     query,
			"variables": variables,
		}

		jsonBody, err := json.Marshal(requestBody)
		if err != nil {
			return result, fmt.Errorf("failed to marshal polling request: %w", err)
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
		if err != nil {
			return result, fmt.Errorf("failed to create polling request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		if accessToken != nil {
			req.Header.Set("Authorization", "Bearer "+*accessToken)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			return result, fmt.Errorf("polling request failed: %w", err)
		}

		var response struct {
			Data struct {
				TaskStatus *struct {
					ID            string   `json:"id"`
					Status        string   `json:"status"`
					Result        *string  `json:"result"`
					Error         *string  `json:"error"`
					CreatedAt     *float64 `json:"created_at"`   // Revert to *float64
					StartedAt     *float64 `json:"started_at"`   // Revert to *float64
					CompletedAt   *float64 `json:"completed_at"` // Revert to *float64
					Progress      *float64 `json:"progress"`
					StatusMessage *string  `json:"status_message"`
				} `json:"task_status"`
			} `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}

		bodyBytes, errRead := io.ReadAll(resp.Body)
		if errRead != nil {
			resp.Body.Close()
			return result, fmt.Errorf("failed to read polling response body: %w", errRead)
		}
		// Re-initialize resp.Body as it has been read
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			// Include raw body in error for better debugging
			return result, fmt.Errorf("failed to decode polling response: %w. Raw Body: %s", err, string(bodyBytes))
		}
		resp.Body.Close()

		// Check for GraphQL errors
		if len(response.Errors) > 0 {
			return result, fmt.Errorf("GraphQL polling error: %s", response.Errors[0].Message)
		}

		taskStatus := response.Data.TaskStatus
		if taskStatus == nil {
			return result, fmt.Errorf("task not found: %s", taskID)
		}

		if verbose && attempt%5 == 0 { // Log every 10 seconds
			fmt.Printf("📊 Status: %s", taskStatus.Status)
			if taskStatus.Progress != nil {
				fmt.Printf(" (%.1f%%)", *taskStatus.Progress*100)
			}
			if taskStatus.StatusMessage != nil {
				fmt.Printf(" - %s", *taskStatus.StatusMessage)
			}
			fmt.Println()
		}

		// Check if task is completed
		if taskStatus.Status == "completed" || taskStatus.Status == "failed" || taskStatus.Status == "cancelled" {
			result.Status = taskStatus.Status

			if taskStatus.Status == "completed" {
				result.ExitCode = 0
			} else {
				result.ExitCode = 1
			}

			// Set timing information from numeric timestamps
			if taskStatus.CreatedAt != nil {
				result.StartTime = time.Unix(int64(*taskStatus.CreatedAt), 0)
			}
			if taskStatus.StartedAt != nil {
				// Prefer StartedAt if available
				result.StartTime = time.Unix(int64(*taskStatus.StartedAt), 0)
			}
			if taskStatus.CompletedAt != nil {
				result.EndTime = time.Unix(int64(*taskStatus.CompletedAt), 0)
				result.Duration = result.EndTime.Sub(result.StartTime)
			} else {
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)
			}

			// Set output and error information
			if taskStatus.Error != nil {
				result.Error = *taskStatus.Error
			}
			if taskStatus.Result != nil {
				result.Output = *taskStatus.Result
				result.Metadata["result"] = *taskStatus.Result
			}
			if taskStatus.Progress != nil {
				result.Metadata["progress"] = *taskStatus.Progress
			}
			if taskStatus.StatusMessage != nil {
				result.Metadata["status_message"] = *taskStatus.StatusMessage
			}

			// Signal completion to stop log streaming
			completionCancel()
			return result, nil
		}

		// If completion was detected by log streaming but task status is not yet complete,
		// return immediately to avoid infinite polling
		select {
		case <-completionCtx.Done():
			// Context was cancelled, which means completion was detected
			// Return current status even if not marked as complete
			result.Status = taskStatus.Status
			result.EndTime = time.Now()
			result.Duration = result.EndTime.Sub(result.StartTime)

			if taskStatus.Error != nil {
				result.Error = *taskStatus.Error
			}
			if taskStatus.Result != nil {
				result.Output = *taskStatus.Result
				result.Metadata["result"] = *taskStatus.Result
			}

			if verbose {
				fmt.Printf("📊 Returning due to completion detection via logs\n")
			}

			return result, nil
		default:
			// Continue with normal polling - wait before next poll
			time.Sleep(2 * time.Second)
		}
	}

	// Timeout reached
	result.Status = "timeout"
	result.Error = fmt.Sprintf("task execution timed out after %d seconds", timeoutSeconds)
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.ExitCode = 1

	return result, fmt.Errorf("task execution timed out")
}

func streamTaskLogs(taskID string, apiURL string, accessToken *string, httpClient *http.Client, completionCtx context.Context) error {
	// Track the last log we've seen
	lastLogIndex := 0

	// Poll for logs every 1 second
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Track if we've seen the completion signal
	completionDetected := false

	for {
		select {
		case <-completionCtx.Done():
			if verbose && !completionDetected {
				fmt.Printf("📊 Log streaming stopped by completion context\n")
			}
			return nil // Completion detected elsewhere, stop streaming
		case <-ticker.C:
			// Query for task logs
			query := `
				query TaskLogs($taskId: ID!, $offset: Int!) {
					task_logs(taskId: $taskId, offset: $offset) {
						entries {
							timestamp
							message
							level
						}
						hasMore
					}
				}
			`

			variables := map[string]interface{}{
				"taskId": taskID,
				"offset": lastLogIndex,
			}

			requestBody := map[string]interface{}{
				"query":     query,
				"variables": variables,
			}

			jsonBody, err := json.Marshal(requestBody)
			if err != nil {
				continue // Skip this iteration on error
			}

			req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
			if err != nil {
				continue // Skip this iteration on error
			}

			req.Header.Set("Content-Type", "application/json")
			if accessToken != nil {
				req.Header.Set("Authorization", "Bearer "+*accessToken)
			}

			resp, err := httpClient.Do(req)
			if err != nil {
				continue // Skip this iteration on error
			}

			// Parse response
			var response struct {
				Data struct {
					TaskLogs *struct {
						Entries []struct {
							Timestamp string `json:"timestamp"`
							Message   string `json:"message"`
							Level     string `json:"level"`
						} `json:"entries"`
						HasMore bool `json:"hasMore"`
					} `json:"task_logs"`
				} `json:"data"`
				Errors []struct {
					Message string `json:"message"`
				} `json:"errors"`
			}

			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				resp.Body.Close()
				continue // Skip this iteration on error
			}
			resp.Body.Close()

			if len(response.Errors) > 0 {
				if verbose {
					fmt.Printf("⚠️  Log streaming error: %s\n", response.Errors[0].Message)
				}
				continue
			}

			taskLogs := response.Data.TaskLogs
			if taskLogs == nil {
				continue // No logs available yet
			}

			// Print new log entries
			for _, entry := range taskLogs.Entries {
				// Check for completion signal in log message first
				if strings.HasPrefix(entry.Message, "TASK_COMPLETED:") {
					// Parse completion signal: TASK_COMPLETED:STATUS:EXITCODE
					parts := strings.Split(entry.Message, ":")
					if len(parts) >= 3 {
						status := parts[1]
						if verbose {
							fmt.Printf("🏁 Task completed with status: %s\n", status)
						}
						completionDetected = true
					}
					continue // Don't print the completion signal itself
				}

				// Format timestamp for display and print the entry
				if timestamp, err := time.Parse(time.RFC3339, entry.Timestamp); err == nil {
					fmt.Printf("%s %s\n", timestamp.Format("15:04:05"), entry.Message)
				} else {
					fmt.Printf("%s\n", entry.Message)
				}
				lastLogIndex++
			}

			// If completion was detected, exit immediately
			if completionDetected {
				return nil
			}

			// If no more logs and task might be complete, check status
			if !taskLogs.HasMore {
				// Make a quick status check to see if task is complete
				statusQuery := `
					query TaskStatus($taskId: ID!) {
						task_status(taskId: $taskId) {
							status
						}
					}
				`

				statusVars := map[string]interface{}{
					"taskId": taskID,
				}

				statusReqBody := map[string]interface{}{
					"query":     statusQuery,
					"variables": statusVars,
				}

				if statusJSON, err := json.Marshal(statusReqBody); err == nil {
					if statusReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(statusJSON)); err == nil {
						statusReq.Header.Set("Content-Type", "application/json")
						if accessToken != nil {
							statusReq.Header.Set("Authorization", "Bearer "+*accessToken)
						}

						if statusResp, err := httpClient.Do(statusReq); err == nil {
							var statusResponse struct {
								Data struct {
									TaskStatus *struct {
										Status string `json:"status"`
									} `json:"task_status"`
								} `json:"data"`
							}

							if err := json.NewDecoder(statusResp.Body).Decode(&statusResponse); err == nil {
								if statusResponse.Data.TaskStatus != nil {
									status := statusResponse.Data.TaskStatus.Status
									statusLower := strings.ToLower(status)
									if statusLower == "completed" || statusLower == "failed" || statusLower == "cancelled" {
										statusResp.Body.Close()
										return nil // Task is complete, stop streaming
									}
								}
							}
							statusResp.Body.Close()
						}
					}
				}
			}
		}
	}
}

// streamTaskLogsViaSubscription uses WebSocket subscription to stream task logs in real-time
func streamTaskLogsViaSubscription(taskID string, apiURL string, accessToken *string, completionCtx context.Context) error {
	// Create WebSocket URL for subscriptions
	wsURL := strings.Replace(apiURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)

	// Prepare headers
	headers := map[string]interface{}{}
	if accessToken != nil {
		headers["Authorization"] = "Bearer " + *accessToken
	}

	// Create subscription client
	subscriptionClient := graphql.NewSubscriptionClient(wsURL).
		WithConnectionParams(headers).
		WithSyncMode(true)

	defer subscriptionClient.Close()

	// Define the subscription query with inline fragments for the union type
	subscriptionQuery := `
		subscription TaskExecutionLogs($taskExecutionId: ID!) {
			taskExecutionLogs(input: {taskExecutionId: $taskExecutionId}) {
				... on TaskLogEntry {
					timestamp
					message
					level
				}
				... on TaskStatusEvent {
					timestamp
					status
					statusMessage
				}
				... on TaskProgressEvent {
					timestamp
					progress
				}
				... on TaskResultEvent {
					timestamp
					executionResult {
						id
						status
						result
						error
						created_at
						started_at
						completed_at
						progress
						status_message
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"taskExecutionId": taskID,
	}

	// Handler for subscription messages
	handler := func(message []byte, err error) error {
		if err != nil {
			if verbose {
				fmt.Printf("⚠️  Subscription error: %v\n", err)
			}
			return err
		}

		// Parse the subscription response
		var response struct {
			Data struct {
				TaskExecutionLogs json.RawMessage `json:"taskExecutionLogs"`
			} `json:"data"`
		}

		if err := json.Unmarshal(message, &response); err != nil {
			if verbose {
				fmt.Printf("⚠️  Failed to parse subscription message: %v\n", err)
			}
			return err
		}

		// Try to determine the event type and handle accordingly
		var eventData map[string]interface{}
		if err := json.Unmarshal(response.Data.TaskExecutionLogs, &eventData); err != nil {
			if verbose {
				fmt.Printf("⚠️  Failed to parse event data: %v\n", err)
			}
			return err
		}

		// Handle different event types based on available fields
		if timestamp, hasTimestamp := eventData["timestamp"]; hasTimestamp {
			if message, hasMessage := eventData["message"]; hasMessage {
				// This is a TaskLogEntry
				level := "INFO"
				if l, hasLevel := eventData["level"]; hasLevel {
					level = fmt.Sprintf("%v", l)
				}
				fmt.Printf("📋 [%s] %s: %v\n", level, timestamp, message)
			} else if status, hasStatus := eventData["status"]; hasStatus {
				// This is a TaskStatusEvent
				statusMsg := ""
				if sm, hasStatusMsg := eventData["statusMessage"]; hasStatusMsg && sm != nil {
					statusMsg = fmt.Sprintf(" - %v", sm)
				}
				fmt.Printf("📊 Status: %v%s\n", status, statusMsg)

				// Check for completion
				if statusStr := fmt.Sprintf("%v", status); statusStr == "COMPLETED" || statusStr == "FAILED" {
					fmt.Printf("📊 Returning due to completion detection via logs\n")
					return graphql.ErrSubscriptionStopped // Signal to stop the subscription
				}
			} else if progress, hasProgress := eventData["progress"]; hasProgress {
				// This is a TaskProgressEvent
				if p, ok := progress.(float64); ok {
					fmt.Printf("📊 Progress: %.1f%%\n", p*100)
				}
			} else if executionResult, hasResult := eventData["executionResult"]; hasResult {
				// This is a TaskResultEvent
				fmt.Printf("✅ Task execution completed\n")
				if verbose {
					resultJSON, _ := json.MarshalIndent(executionResult, "", "  ")
					fmt.Printf("📋 Final result: %s\n", string(resultJSON))
				}
				fmt.Printf("📊 Returning due to completion detection via logs\n")
				return graphql.ErrSubscriptionStopped // Signal to stop the subscription
			}
		}

		return nil
	}

	// Subscribe using raw GraphQL operation
	_, err := subscriptionClient.SubscribeRaw(subscriptionQuery, variables, handler)
	if err != nil {
		if verbose {
			fmt.Printf("⚠️  Failed to subscribe to task logs: %v\n", err)
		}
		return err
	}

	// Run the subscription in a goroutine so we can handle completion context
	done := make(chan error, 1)
	go func() {
		done <- subscriptionClient.Run()
	}()

	// Wait for either subscription completion or context cancellation
	select {
	case <-completionCtx.Done():
		if verbose {
			fmt.Printf("📊 Log streaming stopped by completion context\n")
		}
		return nil
	case err := <-done:
		// If subscription stopped normally (due to completion), that's expected
		if err != nil && err != graphql.ErrSubscriptionStopped && verbose {
			fmt.Printf("⚠️  Subscription ended with error: %v\n", err)
		}
		return nil
	}
}
